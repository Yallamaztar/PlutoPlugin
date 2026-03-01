package rcon

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"plugin/internal/config"
	"plugin/internal/logger"
)

const (
	readBufferSize       = 4096
	defaultReadTimeout   = time.Second
	defaultReadExtension = 350 * time.Millisecond
	defaultTimeout       = 1 * time.Second
	defaultRetryCount    = 3

	infoResponseMarker   = "inforesponse"
	statusResponseMarker = "statusresponse"
	kvDelimiter          = "\\"
	statusHeaderPattern  = `(?i)^num\s+score\s+ping`

	playerLinePattern = `(?P<num>\d+)\s+` +
		`(?P<score>-?\d+)\s+` +
		`(?P<bot>\w+)?\s*` +
		`(?P<ping>\d+|LOAD)\s+` +
		`(?P<guid>[0-9a-fA-F]+)\s+` +
		`(?P<name>.+?)\s+` +
		`(?P<lastmsg>\d+)\s+` +
		`(?P<ipport>\S+)\s+` +
		`(?P<qport>\d+)\s+` +
		`(?P<rate>\d+)`
)

type RCON struct {
	host     string
	password string
	config   *config.Config

	log  *logger.Logger
	conn *net.UDPConn
	mu   sync.Mutex
}

func New(host, password string, cfg *config.Config, log *logger.Logger) (*RCON, error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, errors.New("failed to resolve address")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, errors.New("failed to connect to RCON server")
	}

	return &RCON{
		host:     host,
		password: password,
		config:   cfg,

		log:  log,
		conn: conn,
		mu:   sync.Mutex{},
	}, nil
}

func (r *RCON) TestConnection() error {
	println()
	r.SetDvar("plutoplugin_enabled", "1")
	for i := range 5 {
		r.log.Infof("Attempt %d/5: Attempting to connect to PlutoPlugin GSC\n", i)

		r.SetDvar("plutoplugin_in", "plugin_ready")
		time.Sleep(100 * time.Millisecond)

		d, err := r.GetDvar("plutoplugin_out")
		if err != nil || d.Value == "" {
			r.log.Errorf("reading plutoplugin_out: %v\n", err)
			println()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if len(d.Value) >= 29 && d.Value[22:29] == "success" {
			r.log.Println("PlutoPlugin RCON ready")
			println()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return errors.New("PlutoPlugin not found on the server")
}

func (r *RCON) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.conn.Close()
}

func (r *RCON) GetInfo() (*GetInfo, error) {
	packet := r.buildPacket("getinfo", false)
	if err := r.sendPacket(packet); err != nil {
		return nil, fmt.Errorf("failed to send getinfo request: %w", err)
	}

	lines, err := r.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read getinfo response: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty getinfo response")
	}

	kvPairs := parseKeyValueResponse(lines)
	if len(kvPairs) == 0 {
		return nil, fmt.Errorf("no valid key-value pairs found in getinfo response")
	}

	info := &GetInfo{RetrievedAt: time.Now()}
	populateServerInfo(info, kvPairs)

	return info, nil
}

func (r *RCON) Status() (*Status, error) {
	packet := r.buildPacket("status", true)
	if err := r.sendPacket(packet); err != nil {
		return nil, fmt.Errorf("failed to send status request: %w", err)
	}

	lines, err := r.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty status response")
	}

	status := &Status{
		Raw:         lines,
		RetrievedAt: time.Now(),
	}

	status.Map = extractMapName(lines)
	players, err := parsePlayerList(lines)
	if err != nil {
		return nil, fmt.Errorf("failed to parse player list: %w", err)
	}

	status.Players = players
	return status, nil
}

func (r *RCON) GetStatus() (*GetStatus, error) {
	packet := r.buildPacket("getstatus", false)
	if err := r.sendPacket(packet); err != nil {
		return nil, fmt.Errorf("failed to send getstatus request: %w", err)
	}

	lines, err := r.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read getstatus response: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty getstatus response")
	}

	kvPairs := parseStatusKeyValueResponse(lines)
	if len(kvPairs) == 0 {
		return nil, fmt.Errorf("no valid key-value pairs found in getstatus response")
	}

	info := &GetStatus{RetrievedAt: time.Now()}
	populateServerStatusInfo(info, kvPairs)

	return info, nil
}

func (r *RCON) SetDvar(dvar, value string) {
	if strings.ContainsAny(value, " \t\"") {
		value = fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}

	command := fmt.Sprintf("set %s %s", dvar, value)
	packet := r.buildPacket(command, true)

	_ = r.sendPacket(packet)
}

func (r *RCON) SetInDvar(value string) {
	if strings.ContainsAny(value, " \t\"") {
		value = fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}

	r.SetDvar("plutoplugin_in", value)
}

func (r *RCON) GetDvar(dvar string) (*Dvar, error) {
	if dvar == "" {
		return nil, fmt.Errorf("dvar cannot be empty")
	}

	regexes := r.compileDvarPatterns(dvar)

	var fallbackValue string
	for attempt := range defaultRetryCount {
		lines, err := r.queryDvar(dvar)
		if err != nil {
			return nil, fmt.Errorf("failed to query dvar %q: %w", dvar, err)
		}

		if value := extractDvarValueFromLines(lines, regexes); value != "" {
			return &Dvar{Name: dvar, Value: value}, nil
		}

		if fallbackValue == "" {
			fallbackValue = findFallbackValue(lines)
		}

		if !shouldRetryDvarQuery(lines) {
			break
		}

		if attempt < defaultRetryCount-1 {
			waitTime := time.Duration(attempt+1) * 150 * time.Millisecond
			time.Sleep(waitTime)
		}
	}

	if fallbackValue != "" {
		return &Dvar{Name: dvar, Value: fallbackValue}, nil
	}

	return nil, fmt.Errorf("empty dvar response for %q", dvar)
}

func (r *RCON) Tell(clientNum uint8, message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	packet := r.buildPacket(
		fmt.Sprintf("%d %s %s",
			clientNum, r.config.Gambling.ConsoleName, message,
		), true,
	)

	if err := r.sendPacket(packet); err != nil {
		return fmt.Errorf("failed to send getstatus request: %w", err)
	}

	return nil
}

func (r *RCON) Say(message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	packet := r.buildPacket(fmt.Sprintf("say [%s]: %s", r.config.Gambling.ConsoleName, message), true)
	return r.sendPacket(packet)
}

func (r *RCON) SayRaw(message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	packet := r.buildPacket(fmt.Sprintf("say %s", message), true)
	return r.sendPacket(packet)
}

func (r *RCON) GUIDByClientNum(clientNum uint8) string {
	status, err := r.Status()
	if err != nil {
		return ""
	}

	for _, p := range status.Players {
		if p.ClientNum == int(clientNum) {
			return p.GUID
		}
	}

	return ""
}

func (r *RCON) ClientNumByGUID(guid string) int {
	status, err := r.Status()
	if err != nil {
		return -1
	}

	for _, p := range status.Players {
		if p.GUID == guid {
			return p.ClientNum
		}
	}

	return -1
}
