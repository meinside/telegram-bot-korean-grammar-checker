package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	bot "github.com/meinside/telegram-bot-go"

	"github.com/meinside/telegram-bot-korean-grammar-checker/apis/daum"
)

const (
	ConfigFilename = "config.json"

	CommandStart   = "/start"
	WelcomeMessage = "맞춤법을 검사할 문장을 입력해 주세요."
)

type config struct {
	TelegramApiToken              string `json:"telegram_api_token"`
	DaumApiKey                    string `json:"daum_api_key"`
	TelegramUpdateIntervalSeconds int    `json:"monitor_interval"`
	IsVerbose                     bool   `json:"is_verbose"`
}

func main() {
	// read config file
	var conf config
	if file, err := ioutil.ReadFile(ConfigFilename); err == nil {
		if err := json.Unmarshal(file, &conf); err != nil {
			panic(err)
		} else {
			// XXX - check conf values
			if conf.TelegramUpdateIntervalSeconds <= 0 {
				conf.TelegramUpdateIntervalSeconds = 1
			}
		}
	} else {
		panic(err)
	}

	// setup telegram bot
	client := bot.NewClient(conf.TelegramApiToken)
	client.Verbose = conf.IsVerbose
	if me := client.GetMe(); me.Ok {
		log.Printf("Bot information: @%s (%s)\n", *me.Result.Username, *me.Result.FirstName)

		// delete webhook (getting updates will not work when wehbook is set up)
		if unhooked := client.DeleteWebhook(); unhooked.Ok {
			// wait for new updates
			client.StartMonitoringUpdates(0, conf.TelegramUpdateIntervalSeconds, func(b *bot.Bot, update bot.Update, err error) {
				if err == nil {
					if update.HasMessage() {
						// 'is typing...'
						b.SendChatAction(update.Message.Chat.Id, bot.ChatActionTyping)

						var message string
						if update.Message.HasText() {
							if *update.Message.Text == CommandStart { // skip /start command
								message = WelcomeMessage
							} else {
								if result, err := daum.CheckGrammar(conf.DaumApiKey, *update.Message.Text); err == nil {
									message = daum.BuildResultMessage(result)
								} else {
									message = fmt.Sprintf("API 호출 오류: %s", err)

									log.Printf("*** failed to call api: %s", err)
								}
							}
						} else {
							message = WelcomeMessage
						}

						// send message back
						if sent := b.SendMessage(
							update.Message.Chat.Id,
							&message,
							map[string]interface{}{
								"parse_mode": bot.ParseModeMarkdown, // with markup support
							},
						); !sent.Ok {
							log.Printf("*** failed to send message: %s\n", *sent.Description)
						}
					}
				} else {
					log.Printf("*** error while receiving update (%s)\n", err.Error())
				}
			})
		} else {
			panic("failed to delete webhook")
		}
	} else {
		panic("failed to get info of the bot")
	}
}
