package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    []Server  `yaml:"server"`
	Discord   Discord   `yaml:"discord"`
	Economy   Economy   `yaml:"economy"`
	Gambling  Gambling  `yaml:"gambling"`
	Levels    Levels    `yaml:"levels"`
	IW4MAdmin IW4MAdmin `yaml:"iw4madmin"`
}

type Server struct {
	Host          string `yaml:"host"`
	Password      string `yaml:"password"`
	LogPath       string `yaml:"log_path"`
	CommandPrefix string `yaml:"command_prefix"`
}

type Gambling struct {
	Enabled     bool    `yaml:"enabled"`
	WinChance   float64 `yaml:"win_chance"`
	Currency    string  `yaml:"currency"`
	ConsoleName string  `yaml:"console_name"`
}

type Economy struct {
	Enabled         bool `yaml:"enabled"`
	KillReward      int  `yaml:"kill_reward"`
	JoinReward      int  `yaml:"join_reward"`
	FirstTimeReward int  `yaml:"first_time_reward"`
	DeathPenalty    int  `yaml:"death_penalty"`
}

type Levels struct {
	User  string `yaml:"user"`
	Admin string `yaml:"admin"`
	Owner string `yaml:"owner"`
}

type Discord struct {
	Enabled     bool   `yaml:"enabled"`
	InviteLink  string `yaml:"invite_link"`
	WebhookLink string `yaml:"webhook_link"`
	BotToken    string `yaml:"bot_token"`
}

type IW4MAdmin struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	ServerID int64  `yaml:"server_id"`
	Cookie   string `yaml:"cookie"`
}

func Default() *Config {
	return &Config{
		Gambling: Gambling{
			Enabled:     true,
			WinChance:   0.45,
			Currency:    "$",
			ConsoleName: "^6PlutoPlugin^7",
		},

		Economy: Economy{
			Enabled:         true,
			KillReward:      350,
			JoinReward:      500,
			FirstTimeReward: 10_000,
			DeathPenalty:    550,
		},

		Levels: Levels{
			User:  "user",
			Admin: "admin",
			Owner: "owner",
		},
	}
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile("config.yaml", data, 0644)
}

func (c *Config) Load() (*Config, error) {
	if _, err := os.Stat("config.yaml"); errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
