package iw4m

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"plugin/internal/config"
	"strconv"
	"time"
)

type IW4MWrapper struct {
	host     string
	serverID int64
	cookie   string

	config *config.Config
	client *http.Client
}

type findClient struct {
	TotalFoundClients int `json:"totalFoundClients"`
	Clients           []struct {
		ClientID int    `json:"clientId"`
		XUID     string `json:"xuid"`
		Name     string `json:"name"`
	} `json:"clients"`
}

func New(config *config.Config) *IW4MWrapper {
	return &IW4MWrapper{
		host:     config.IW4MAdmin.Host,
		serverID: config.IW4MAdmin.ServerID,
		cookie:   config.IW4MAdmin.Cookie,

		config: config,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (w *IW4MWrapper) do(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, w.host+"/"+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", w.cookie)
	req.Header.Set("User-Agent", "gambling-bot/1.0")

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

func (w *IW4MWrapper) BanPlayer(player, reason string) error {
	return w.ExecuteCommand(fmt.Sprintf("!ban %s %s", player, reason))
}

func (w *IW4MWrapper) ClientIDFromGUID(guid string) *int {
	endpoint := fmt.Sprintf(
		"/api/client/find?name=&guid=%s&count=10&offset=0&direction=0",
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
