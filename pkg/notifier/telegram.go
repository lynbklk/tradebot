package notifier

import (
	"errors"
	"fmt"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/order"
	"github.com/rs/zerolog/log"
	tb "gopkg.in/tucnak/telebot.v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	buyRegexp  = regexp.MustCompile(`/buy\s+(?P<pair>\w+)\s+(?P<amount>[0-9]+(?:\.\d+)?)(?P<percent>%)?`)
	sellRegexp = regexp.MustCompile(`/sell\s+(?P<pair>\w+)\s+(?P<amount>[0-9]+(?:\.\d+)?)(?P<percent>%)?`)
)

type telegram struct {
	settings        model.Settings
	orderController *order.Controller
	defaultMenu     *tb.ReplyMarkup
	client          *tb.Bot
}

type Option func(telegram *telegram)

func NewTelegram(controller *order.Controller, settings model.Settings, options ...Option) (Notifier, error) {
	menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	poller := &tb.LongPoller{Timeout: 10 * time.Second}

	userMiddleware := tb.NewMiddlewarePoller(poller, func(u *tb.Update) bool {
		if u.Message == nil || u.Message.Sender == nil {
			log.Error().Msgf("no message, %v", u)
			return false
		}

		for _, user := range settings.Telegram.Users {
			if int(u.Message.Sender.ID) == user {
				return true
			}
		}

		log.Error().Msgf("invalid user: %v", u.Message)
		return false
	})

	client, err := tb.NewBot(tb.Settings{
		ParseMode: tb.ModeMarkdown,
		Token:     settings.Telegram.Token,
		Poller:    userMiddleware,
	})
	if err != nil {
		return nil, err
	}

	var (
		statusBtn  = menu.Text("/status")
		profitBtn  = menu.Text("/profit")
		balanceBtn = menu.Text("/balance")
		startBtn   = menu.Text("/start")
		stopBtn    = menu.Text("/stop")
		buyBtn     = menu.Text("/buy")
		sellBtn    = menu.Text("/sell")
	)

	err = client.SetCommands([]tb.Command{
		{Text: "/help", Description: "Display help instructions"},
		{Text: "/stop", Description: "Stop buy and sell coins"},
		{Text: "/start", Description: "Start buy and sell coins"},
		{Text: "/status", Description: "Check bot status"},
		{Text: "/balance", Description: "Wallet balance"},
		{Text: "/profit", Description: "Summary of last trade results"},
		{Text: "/buy", Description: "open a buy order"},
		{Text: "/sell", Description: "open a sell order"},
	})
	if err != nil {
		return nil, err
	}

	menu.Reply(
		menu.Row(statusBtn, balanceBtn, profitBtn),
		menu.Row(startBtn, stopBtn, buyBtn, sellBtn),
	)

	bot := &telegram{
		orderController: controller,
		client:          client,
		settings:        settings,
		defaultMenu:     menu,
	}

	for _, option := range options {
		option(bot)
	}

	client.Handle("/help", bot.HelpHandle)
	client.Handle("/start", bot.StartHandle)
	client.Handle("/stop", bot.StopHandle)
	client.Handle("/status", bot.StatusHandle)
	client.Handle("/balance", bot.BalanceHandle)
	client.Handle("/profit", bot.ProfitHandle)
	client.Handle("/buy", bot.BuyHandle)
	client.Handle("/sell", bot.SellHandle)

	return bot, nil
}

func (t telegram) Start() {
	go t.client.Start()
	for _, id := range t.settings.Telegram.Users {
		_, err := t.client.Send(&tb.User{ID: int64(id)}, "Bot initialized.", t.defaultMenu)
		if err != nil {
			log.Error().Err(err).Msg("bot start failed. ")
		}
	}
}

func (t telegram) Notify(text string) {
	for _, user := range t.settings.Telegram.Users {
		_, err := t.client.Send(&tb.User{ID: int64(user)}, text)
		if err != nil {
			log.Error().Err(err).Msg("bot notify failed. ")
		}
	}
}

func (t telegram) BalanceHandle(m *tb.Message) {
	message := "*BALANCE*\n"
	quotesValue := make(map[string]float64)
	total := 0.0

	for _, pair := range t.settings.Pairs {
		assetPair, quotePair := exchange.SplitAssetQuote(pair)
		assetSize, quoteSize, err := t.orderController.Position(pair)
		if err != nil {
			log.Error().Err(err).Msg("bot balance handle failed.")
			t.OnError(err)
			return
		}

		quote, err := t.orderController.LastQuote(pair)
		if err != nil {
			log.Error().Err(err).Msg("bot balance handle failed.")
			t.OnError(err)
			return
		}

		assetValue := assetSize * quote
		quotesValue[quotePair] = quoteSize
		total += assetValue
		message += fmt.Sprintf("%s: `%.4f` ≅ `%.2f` %s \n", assetPair, assetSize, assetValue, quotePair)
	}

	for quote, value := range quotesValue {
		total += value
		message += fmt.Sprintf("%s: `%.4f`\n", quote, value)
	}

	message += fmt.Sprintf("-----\nTotal: `%.4f`\n", total)

	_, err := t.client.Send(m.Sender, message)
	if err != nil {
		log.Error().Err(err).Msg("bot balance handle failed.")
	}
}

func (t telegram) HelpHandle(m *tb.Message) {
	commands, err := t.client.GetCommands()
	if err != nil {
		log.Error().Err(err).Msg("bot help handle failed.")
		t.OnError(err)
		return
	}

	lines := make([]string, 0, len(commands))
	for _, command := range commands {
		lines = append(lines, fmt.Sprintf("/%s - %s", command.Text, command.Description))
	}

	_, err = t.client.Send(m.Sender, strings.Join(lines, "\n"))
	if err != nil {
		log.Error().Err(err).Msg("bot help handle failed.")
	}
}

func (t telegram) ProfitHandle(m *tb.Message) {
	if len(t.orderController.Results) == 0 {
		_, err := t.client.Send(m.Sender, "No trades registered.")
		if err != nil {
			log.Error().Err(err).Msg("bot profit handle failed.")
		}
		return
	}

	for pair, summary := range t.orderController.Results {
		_, err := t.client.Send(m.Sender, fmt.Sprintf("*PAIR*: `%s`\n`%s`", pair, summary.String()))
		if err != nil {
			log.Error().Err(err).Msg("bot profit handle failed.")
		}
	}
}

func (t telegram) BuyHandle(m *tb.Message) {
	match := buyRegexp.FindStringSubmatch(m.Text)
	if len(match) == 0 {
		_, err := t.client.Send(m.Sender, "Invalid command.\nExamples of usage:\n`/buy BTCUSDT 100`\n\n`/buy BTCUSDT 50%`")
		if err != nil {
			log.Error().Err(err).Msg("bot buy handle failed.")
		}
		return
	}

	command := make(map[string]string)
	for i, name := range buyRegexp.SubexpNames() {
		if i != 0 && name != "" {
			command[name] = match[i]
		}
	}

	pair := strings.ToUpper(command["pair"])
	amount, err := strconv.ParseFloat(command["amount"], 64)
	if err != nil {
		log.Error().Err(err).Msg("bot buy handle failed.")
		t.OnError(err)
		return
	} else if amount <= 0 {
		_, err := t.client.Send(m.Sender, "Invalid amount")
		if err != nil {
			log.Error().Err(err).Msg("bot buy handle failed.")
		}
		return
	}

	if command["percent"] != "" {
		_, quote, err := t.orderController.Position(pair)
		if err != nil {
			log.Error().Err(err).Msg("bot buy handle failed.")
			t.OnError(err)
			return
		}

		amount = amount * quote / 100.0
	}

	order, err := t.orderController.CreateOrderMarketQuote(model.SideTypeBuy, pair, amount)
	if err != nil {
		return
	}
	log.Info().Msgf("BUY ORDER CREATED: %v", order)
}

func (t telegram) SellHandle(m *tb.Message) {
	match := sellRegexp.FindStringSubmatch(m.Text)
	if len(match) == 0 {
		_, err := t.client.Send(m.Sender, "Invalid command.\nExample of usage:\n`/sell BTCUSDT 100`\n\n`/sell BTCUSDT 50%")
		if err != nil {
			log.Error().Err(err).Msg("bot sell handle failed.")
		}
		return
	}

	command := make(map[string]string)
	for i, name := range sellRegexp.SubexpNames() {
		if i != 0 && name != "" {
			command[name] = match[i]
		}
	}

	pair := strings.ToUpper(command["pair"])
	amount, err := strconv.ParseFloat(command["amount"], 64)
	if err != nil {
		log.Error().Err(err).Msg("bot sell handle failed.")
		t.OnError(err)
		return
	} else if amount <= 0 {
		_, err := t.client.Send(m.Sender, "Invalid amount")
		if err != nil {
			log.Error().Err(err).Msg("bot sell handle failed.")
		}
		return
	}

	if command["percent"] != "" {
		asset, _, err := t.orderController.Position(pair)
		if err != nil {
			return
		}

		amount = amount * asset / 100.0
		order, err := t.orderController.CreateOrderMarket(model.SideTypeSell, pair, amount)
		if err != nil {
			return
		}
		log.Info().Msgf("SELL ORDER CREATED: %v", order)
		return
	}

	order, err := t.orderController.CreateOrderMarketQuote(model.SideTypeSell, pair, amount)
	if err != nil {
		return
	}
	log.Info().Msgf("SELL ORDER CREATED: %v", order)
}

func (t telegram) StatusHandle(m *tb.Message) {
	status := t.orderController.Status()
	_, err := t.client.Send(m.Sender, fmt.Sprintf("Status: `%s`", status))
	if err != nil {
		log.Error().Err(err).Msg("bot status handle failed.")
	}
}

func (t telegram) StartHandle(m *tb.Message) {
	if t.orderController.Status() == order.StatusRunning {
		_, err := t.client.Send(m.Sender, "Bot is already running.", t.defaultMenu)
		if err != nil {
			log.Error().Err(err).Msg("bot start handle failed.")
		}
		return
	}

	t.orderController.Start()
	_, err := t.client.Send(m.Sender, "Bot started.", t.defaultMenu)
	if err != nil {
		log.Error().Err(err).Msg("bot start handle failed.")
	}
}

func (t telegram) StopHandle(m *tb.Message) {
	if t.orderController.Status() == order.StatusStopped {
		_, err := t.client.Send(m.Sender, "Bot is already stopped.", t.defaultMenu)
		if err != nil {
			log.Error().Err(err).Msg("bot stop handle failed.")
		}
		return
	}

	t.orderController.Stop()
	_, err := t.client.Send(m.Sender, "Bot stopped.", t.defaultMenu)
	if err != nil {
		log.Error().Err(err).Msg("bot stop handle failed.")
	}
}

func (t telegram) OnOrder(order model.Order) {
	title := ""
	switch order.Status {
	case model.OrderStatusTypeFilled:
		title = fmt.Sprintf("✅ ORDER FILLED - %s", order.Pair)
	case model.OrderStatusTypeNew:
		title = fmt.Sprintf("🆕 NEW ORDER - %s", order.Pair)
	case model.OrderStatusTypeCanceled, model.OrderStatusTypeRejected:
		title = fmt.Sprintf("❌ ORDER CANCELED / REJECTED - %s", order.Pair)
	}
	message := fmt.Sprintf("%s\n-----\n%s", title, order)
	t.Notify(message)
}

func (t telegram) OnError(err error) {
	title := "🛑 ERROR"

	var orderError *exchange.OrderError
	if errors.As(err, &orderError) {
		message := fmt.Sprintf(`%s
		-----
		Pair: %s
		Quantity: %.4f
		-----
		%s`, title, orderError.Pair, orderError.Quantity, orderError.Err)
		t.Notify(message)
		return
	}

	t.Notify(fmt.Sprintf("%s\n-----\n%s", title, err))
}
