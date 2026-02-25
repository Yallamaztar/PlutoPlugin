package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"plugin/internal/helpers"
	"time"
)

type embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp"`
}

type payload struct {
	Content string  `json:"content"`
	Embeds  []embed `json:"embeds"`
}

type Webhook struct {
	URL string
}

var client = &http.Client{Timeout: 2 * time.Second}
var webhookQueue = make(chan struct{}, 10)

func (w *Webhook) SendWebhook(payload payload) {
	if w.URL == "" {
		return
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, w.URL, bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}

func (w *Webhook) WinWebhook(player string, amount int64) {
	w.SendWebhook(payload{
		Embeds: []embed{{
			Title: "🎉 Jackpot Hit!",
			Description: "💰 **Winner:** **" + player + "**\n" +
				"📈 **Payout:** **$" + helpers.FormatMoney(amount) + "**\n\n" +
				"🔥 *Luck was on their side today.*",
			Color:     0x2ecc71, // smooth emerald green
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}},
	})
}
