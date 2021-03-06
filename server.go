package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	aelog "google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

func init() {
	bot, err := NewExampleBot(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/callback", bot)
}

const (
	ctxKeyBotClient = "bot-client"
)

// botClient returns a bot client inside context.
func botClient(ctx context.Context) *linebot.Client {
	return ctx.Value(ctxKeyBotClient).(*linebot.Client)
}

// The ExampleBot is type of the bot event handler.
type ExampleBot struct {
	webhookHandler *httphandler.WebhookHandler
}

// NewExampleBot constructs new ExampleBot instance.
func NewExampleBot(channelSecret, channelToken string) (*ExampleBot, error) {
	handler, err := httphandler.New(channelSecret, channelToken)
	if err != nil {
		return nil, err
	}
	bot := &ExampleBot{
		webhookHandler: handler,
	}
	handler.HandleEvents(bot.handleEvents)
	return bot, nil
}

// ServeHTTP implements for http.Handler
func (bot *ExampleBot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bot.webhookHandler.ServeHTTP(w, r)
}

// newContext returns new context with a bot client instance.
func (bot *ExampleBot) newContext(req *http.Request) (context.Context, error) {
	ctx := appengine.NewContext(req)
	httpclient := urlfetch.Client(ctx)
	client, err := bot.webhookHandler.NewClient(linebot.WithHTTPClient(httpclient))
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, ctxKeyBotClient, client)
	// 他にもGoogle Maps等のclientがある場合はcontext.WithValue()でcontextにinstanceを持たせる
	return ctx, nil
}

// handleEvents is handler function for handler.HandleEvents.
func (bot *ExampleBot) handleEvents(events []*linebot.Event, req *http.Request) {
	ctx, err := bot.newContext(req)
	if err != nil {
		aelog.Errorf(ctx, "%v", err)
		return
	}
	// 大量のeventを処理するならeventsをジョブキューに入れるのがベター
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				bot.handleMessageText(ctx, event, message)
			case *linebot.ImageMessage:
				// bot.handleMessageImage(ctx, event, message)
			case *linebot.VideoMessage:
				// bot.handleMessageVideo(ctx, event, message)
			case *linebot.AudioMessage:
				// bot.handleMessageAudio(ctx, event, message)
			case *linebot.LocationMessage:
				// bot.handleMessageLocation(ctx, event, message)
			case *linebot.StickerMessage:
				// bot.handleMessageSticker(ctx, event, message)
			}
		case linebot.EventTypeFollow:
			// bot.handleFollow(ctx, event)
		case linebot.EventTypeUnfollow:
			// bot.handleUnfollow(ctx, event)
		case linebot.EventTypeJoin:
			// bot.handleJoin(ctx, event)
		case linebot.EventTypeLeave:
			// bot.handleLeave(ctx, event)
		case linebot.EventTypePostback:
			// bot.handlePostback(ctx, event)
		case linebot.EventTypeBeacon:
			bot.handleBeacon(ctx, event)
		}
	}
}

// handleMessageText processes a text message event.
func (bot *ExampleBot) handleMessageText(ctx context.Context, event *linebot.Event, message *linebot.TextMessage) {
	client := botClient(ctx)
	_, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).WithContext(ctx).Do()
	if err != nil {
		aelog.Errorf(ctx, "%v", err)
	}
}

// handleBeacon processes a beacon event.
func (bot *ExampleBot) handleBeacon(ctx context.Context, event *linebot.Event) {
	client := botClient(ctx)
	_, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("ビーコン見つけた！")).WithContext(ctx).Do()
	if err != nil {
		aelog.Errorf(ctx, "%v", err)
	}
}
