package iw4m

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"plugin/internal/config"
	"plugin/internal/logger"
	"strconv"
	"strings"
	"time"
)

type IW4MWrapper struct {
	Host     string
	serverID int64
	cookie   string

	log    *logger.Logger
	config *config.Config
	client *http.Client
}

func New(config *config.Config, log *logger.Logger) *IW4MWrapper {
	// making sure end of URL doesnt have a "/"
	host := strings.TrimRight(config.IW4MAdmin.Host, "/")
	return &IW4MWrapper{
		Host:     host,
		serverID: config.IW4MAdmin.ServerID,
		cookie:   config.IW4MAdmin.Cookie,

		log:    log,
		config: config,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (w *IW4MWrapper) do(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, w.Host+"/"+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", w.cookie)
	req.Header.Set("User-Agent", "plutoplugin-bot/1.0") // for shits n gigles

	return w.client.Do(req)
}

func (w *IW4MWrapper) ExecuteCommand(command string) error {
	endpoint := fmt.Sprintf(
		"Console/Execute?serverId=%s&command=%s",
		url.QueryEscape(strconv.FormatInt(w.serverID, 10)),
		url.QueryEscape(command),
	)
	_, err := w.do(endpoint)
	return err
}

func (w *IW4MWrapper) SetLevel(player, level string) error {
	return w.ExecuteCommand(fmt.Sprintf("!sl %s %s", player, level))
}

func (w *IW4MWrapper) Ban(clientID int, reason string) error {
	return w.ExecuteCommand(fmt.Sprintf("!ban @%d %s", clientID, reason))
}

func (w *IW4MWrapper) Unban(clientID int, reason string) error {
	return w.ExecuteCommand(fmt.Sprintf("!unban @%d %s", clientID, reason))
}

type findClient struct {
	TotalFoundClients int `json:"totalFoundClients"`

	Clients []struct {
		ClientID int    `json:"clientId"`
		XUID     string `json:"xuid"`
		Name     string `json:"name"`
	} `json:"clients"`
}

func (w *IW4MWrapper) ClientIDFromGUID(guid string) *int {
	endpoint := fmt.Sprintf(
		"api/client/find?name=&guid=%s&count=10&offset=0&direction=0",
		url.QueryEscape(guid),
	)

	res, err := w.do(endpoint)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil
	}

	var client findClient
	if err := json.NewDecoder(res.Body).Decode(&client); err != nil {
		return nil
	}

	if client.TotalFoundClients == 0 || len(client.Clients) == 0 {
		return nil
	}

	return &client.Clients[0].ClientID
}

type stats struct {
	Name               string
	Ranking            int
	Kills              int
	Deaths             int
	Performance        float64
	LastPlayed         string
	TotalSecondsPlayed int
	ServerName         string
	ServerGame         string
}

func (w *IW4MWrapper) Stats(clientID, index int) (*stats, error) {
	endpoint := fmt.Sprintf("api/stats/%d", clientID)

	res, err := w.do(endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var statList []stats
	if err := json.NewDecoder(res.Body).Decode(&statList); err != nil {
		return nil, err
	}

	if len(statList) == 0 {
		return nil, fmt.Errorf("no stats found for clientID %d", clientID)
	}

	return &statList[index], nil
}

func (w *IW4MWrapper) TestConnection() error {
	for i := range 5 {
		w.log.Infof("Attempt %d/5: Attempting to connect to IW4M-Admin\n", i+1)

		stat, err := w.Stats(2, 0)
		if err != nil {
			w.log.Errorf("couldn't connect to IW4M-Admin: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if stat == nil || stat.Name == "" {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return nil
	}

	return fmt.Errorf("couldn't connect to IW4M-Admin (%s)", w.Host)
}
