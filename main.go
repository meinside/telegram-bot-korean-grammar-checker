package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	bot "github.com/meinside/telegram-bot-go"

	"github.com/meinside/telegram-bot-korean-grammar-checker/apis/daum"

	"github.com/meinside/loggly-go"
)

const (
	appName = "KoreanGrammarCheckerBot"

	configFilename = "config.json"

	commandStart   = "/start"
	welcomeMessage = "맞춤법을 검사할 문장을 입력해 주세요."
)

type config struct {
	TelegramApiToken              string `json:"telegram_api_token"`
	DaumApiKey                    string `json:"daum_api_key"`
	TelegramUpdateIntervalSeconds int    `json:"monitor_interval"`
	LogglyToken                   string `json:"loggly_token,omitempty"`
	IsVerbose                     bool   `json:"is_verbose"`
}

type LogglyLog struct {
	Application string      `json:"app"`
	Severity    string      `json:"severity"`
	Message     string      `json:"message,omitempty"`
	Object      interface{} `json:"obj,omitempty"`
}

var logger *loggly.Loggly

func main() {
	// read config file
	var conf config
	if file, err := ioutil.ReadFile(configFilename); err == nil {
		if err := json.Unmarshal(file, &conf); err != nil {
			panic(err)
		} else {
			if conf.TelegramUpdateIntervalSeconds <= 0 {
				conf.TelegramUpdateIntervalSeconds = 1
			}

			if conf.LogglyToken != "" {
				logger = loggly.New(conf.LogglyToken)
			} else {
				logger = nil
			}
		}
	} else {
		panic(err)
	}

	// setup telegram bot
	client := bot.NewClient(conf.TelegramApiToken)
	client.Verbose = conf.IsVerbose
	if me := client.GetMe(); me.Ok {
		logMessage(fmt.Sprintf("Starting bot: @%s (%s)", *me.Result.Username, me.Result.FirstName))

		// delete webhook (getting updates will not work when wehbook is set up)
		if unhooked := client.DeleteWebhook(); unhooked.Ok {
			// wait for new updates
			client.StartMonitoringUpdates(0, conf.TelegramUpdateIntervalSeconds, func(b *bot.Bot, update bot.Update, err error) {
				if err == nil {
					if update.HasMessage() {
						// 'is typing...'
						b.SendChatAction(update.Message.Chat.ID, bot.ChatActionTyping)

						var message, username string
						if update.Message.HasText() {
							if *update.Message.Text == commandStart { // skip /start command
								message = welcomeMessage
							} else {
								// log request
								if update.Message.From.Username == nil {
									username = update.Message.From.FirstName
								} else {
									username = *update.Message.From.Username
								}
								logRequest(*update.Message.Text, username)

								// check grammar
								if result, err := daum.CheckGrammar(conf.DaumApiKey, *update.Message.Text); err == nil {
									message = daum.BuildResultMessage(result)
								} else {
									message = fmt.Sprintf("API 호출 오류: %s", err)

									logError(fmt.Sprintf("Failed to call API: %s", err))
								}
							}
						} else {
							message = welcomeMessage
						}

						// send message back
						if sent := b.SendMessage(
							update.Message.Chat.ID,
							message,
							map[string]interface{}{
								"parse_mode": bot.ParseModeMarkdown, // with markup support
							},
						); !sent.Ok {
							logError(fmt.Sprintf("Failed to send message: %s", *sent.Description))
						}
					}
				} else {
					logError(fmt.Sprintf("Error while receiving update: %s", err))
				}
			})
		} else {
			panic("failed to delete webhook")
		}
	} else {
		panic("failed to get info of the bot")
	}
}

func logMessage(message string) {
	log.Println(message)

	if logger != nil {
		logger.Log(LogglyLog{
			Application: appName,
			Severity:    "Log",
			Message:     message,
		})
	}
}

func logError(message string) {
	log.Println(message)

	if logger != nil {
		logger.Log(LogglyLog{
			Application: appName,
			Severity:    "Error",
			Message:     message,
		})
	}
}

func logRequest(text, username string) {
	if logger != nil {
		logger.Log(LogglyLog{
			Application: appName,
			Severity:    "Verbose",
			Object: struct {
				Username string `json:"username"`
				Text     string `json:"text"`
			}{
				Username: username,
				Text:     text,
			},
		})
	}
}
