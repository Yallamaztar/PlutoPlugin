package bot

import (
	"fmt"
	"plugin/internal/commands/pay"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/logger"
	"plugin/internal/service/link"
	"plugin/internal/service/player"
	"plugin/internal/service/stats"
	"plugin/internal/service/wallet"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	cfg         *config.Config
	log         *logger.Logger
	player      *player.Service
	wallet      *wallet.Service
	walletStats *stats.WalletStatsService
	link        *link.Service
	webhook     *webhook.Webhook
	session     *discordgo.Session
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "pay",
		Description: "Pay another player",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Target player",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Amount to pay",
				Required:    true,
			},
		},
	},
	{
		Name:        "balance",
		Description: "Check your or another player's balance",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Optional target player",
				Required:    false,
			},
		},
	},
	{
		Name:        "link",
		Description: "Link your in-game account",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "code",
				Description: "Link code from in-game",
				Required:    true,
			},
		},
	},
}

func New(cfg *config.Config, log *logger.Logger, player *player.Service, wallet *wallet.Service, walletStats *stats.WalletStatsService, link *link.Service, webhook *webhook.Webhook) (*Bot, error) {
	s, err := discordgo.New("Bot " + cfg.Discord.BotToken)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		cfg:         cfg,
		log:         log,
		player:      player,
		wallet:      wallet,
		walletStats: walletStats,
		link:        link,
		webhook:     webhook,
		session:     s,
	}

	s.AddHandler(b.onInteraction)

	return b, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	_, err := b.session.ApplicationCommandBulkOverwrite(
		b.session.State.User.ID,
		"",
		commands,
	)
	if err != nil {
		return err
	}

	b.log.Infoln("Discord bot started")
	return nil
}

func (b *Bot) Close() {
	_ = b.session.Close()
}

func (b *Bot) onInteraction(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "pay":
		b.handlePay(s, i)
	case "balance":
		b.handleBalance(s, i)
	case "link":
		b.handleLink(s, i)
	}
}

func (b *Bot) handlePay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	amount := int(opts[1].IntValue())

	fromPlayer, err := b.player.GetPlayerByDiscordID(i.Member.User.ID)
	if err != nil || fromPlayer == nil {
		respondEphemeral(s, i, "❌ You are not linked")
		return
	}

	toPlayer, err := b.player.GetPlayerByDiscordID(target.ID)
	if err != nil || toPlayer == nil {
		respondEphemeral(s, i, "❌ Target user not linked")
		return
	}

	_, err = pay.Pay(fromPlayer.ID, toPlayer.ID, amount, b.cfg, b.player, b.wallet, b.walletStats, b.webhook)
	if err != nil {
		respondEphemeral(s, i, "❌ "+err.Error())
		return
	}

	respondEphemeral(s, i, fmt.Sprintf("✅ paid **%s %s%d**", toPlayer.Name, b.cfg.Gambling.Currency, amount))
}

func (b *Bot) handleBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	var discordID string
	var name string

	if len(opts) > 0 {
		u := opts[0].UserValue(s)
		discordID = u.ID
		name = u.Username
	} else {
		discordID = i.Member.User.ID
		name = i.Member.User.Username
	}

	player, err := b.player.GetPlayerByDiscordID(discordID)
	if err != nil || player == nil {
		respondEphemeral(s, i, "❌ Account not linked")
		return
	}

	balance, err := b.wallet.GetBalance(player.ID)
	if err != nil {
		respondEphemeral(s, i, "❌ Failed to get balance")
		return
	}

	respondEphemeral(s, i, fmt.Sprintf("💰 **%s's balance:** %s%d", name, b.cfg.Gambling.Currency, balance))
}

func (b *Bot) handleLink(s *discordgo.Session, i *discordgo.InteractionCreate) {
	code := i.ApplicationCommandData().Options[0].StringValue()

	playerID, err := b.link.GetPlayerIDByCode(code)
	if err != nil {
		respondEphemeral(s, i, "❌ Invalid or expired code")
		return
	}

	if err := b.player.UpdateDiscordID(playerID, i.Member.User.ID); err != nil {
		respondEphemeral(s, i, "❌ Failed to link account")
		return
	}

	_ = b.link.DeleteByPlayerID(playerID)
	respondEphemeral(s, i, "✅ Successfully linked your discord to account!")
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
