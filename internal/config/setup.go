package config

import (
	"bufio"
	"fmt"
	"os"
	"plugin/internal/logger"
	"strings"
)

func Setup(log *logger.Logger) (*Config, error) {
	cfg := Default()

	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		log.Println("No config file found. Let's create one:")

		cfg.Server = promptServers()
		cfg.Discord = promptDiscord()
		cfg.IW4MAdmin = promptIW4M()

		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save config: %w", err)
		}

		log.Println("Config saved to config.yaml!")
	}

	loaded, err := cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return loaded, nil
}

func promptServers() []Server {
	var servers []Server
	reader := bufio.NewReader(os.Stdin)

	for {
		host := readLine(reader, "Enter RCON host (IP:Port): ")
		password := readLine(reader, "Enter RCON password: ")
		logPath := readLine(reader, "Enter server log path: ")
		prefix := readLine(reader, "Enter command prefix (default: !): ")
		if strings.TrimSpace(prefix) == "" {
			prefix = "!"
		}

		servers = append(servers, Server{
			Host:          host,
			Password:      password,
			LogPath:       logPath,
			CommandPrefix: prefix,
		})

		if !yesNo(reader, "Add another server? (Y/n): ") {
			break
		}
	}

	return servers
}

func promptDiscord() Discord {
	reader := bufio.NewReader(os.Stdin)
	if yesNo(reader, "Enable Discord integration? (yes/no): ") {
		return Discord{
			Enabled:     true,
			WebhookLink: readLine(reader, "Enter Discord webhook link: "),
			InviteLink:  readLine(reader, "Enter Discord invite link (optional): "),
			BotToken:    readLine(reader, "Enter Discord bot token: "),
		}
	}
	return Discord{}
}

func promptIW4M() IW4MAdmin {
	reader := bufio.NewReader(os.Stdin)
	if yesNo(reader, "Enable IW4M-Admin integration? (yes/no): ") {
		id := int64(0)
		fmt.Sscan(readLine(reader, "Enter IW4M-Admin server ID: "), &id)

		return IW4MAdmin{
			Enabled:  true,
			Host:     readLine(reader, "Enter IW4M-Admin host: "),
			ServerID: id,
			Cookie:   readLine(reader, "Enter IW4M-Admin cookie: "),
		}
	}
	return IW4MAdmin{}
}

func readLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func yesNo(reader *bufio.Reader, prompt string) bool {
	resp := strings.ToLower(readLine(reader, prompt))
	return resp == "yes" || resp == "y"
}
