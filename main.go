package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	bot "github.com/meinside/telegram-bot-go"
)

const (
	ConfigFilename                 = "config.json"
	DaumKoreanGrammarCheckerApiUrl = "https://apis.daum.net/grammar-checker/v1/check.json"

	CommandStart   = "/start"
	WelcomeMessage = "맞춤법을 검사할 문장을 입력해 주세요."
)

type config struct {
	TelegramApiToken              string `json:"telegram_api_token"`
	DaumApiKey                    string `json:"daum_api_key"`
	TelegramUpdateIntervalSeconds int    `json:"monitor_interval"`
	IsVerbose                     bool   `json:"is_verbose"`
}

// api result
type checkResult struct {
	Sentences []struct {
		Sentence string `json:"sentence"`
		Result   []struct {
			Input   string   `json:"input"`
			Output  string   `json:"output"`
			Etype   string   `json:"etype"`
			Help    []string `json:"help,omitempty"`
			HelpMo  []string `json:"help_mo,omitempty"`
			Example []string `json:"example,omitempty"`
		} `json:"result"`
	} `json:"sentences"`

	// when error
	ErrorType string `json:"errorType,omitempty"`
	Message   string `json:"message,omitempty"`
}

// check grammar of given text
//
// https://developers.daum.net/services/apis/grammar-checker/v1/check.json
func checkGrammar(apiKey, text string) (checkResult, error) {
	client := &http.Client{}

	var err error
	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequest("GET", DaumKoreanGrammarCheckerApiUrl, nil); err == nil {
		q := req.URL.Query()
		q.Add("apikey", apiKey)
		q.Add("query", text)
		q.Add("help", "on")
		req.URL.RawQuery = q.Encode()

		if res, err = client.Do(req); err == nil {
			defer res.Body.Close()

			var bytes []byte
			if bytes, err = ioutil.ReadAll(res.Body); err == nil {
				var result checkResult
				if err = json.Unmarshal(bytes, &result); err == nil {
					if result.ErrorType == "" {
						return result, nil
					} else {
						err = fmt.Errorf("%s - %s", result.ErrorType, result.Message)
					}
				}
			}
		}
	}

	return checkResult{}, err
}

// build up result message string
/*
  * example:

  ► 이 대화의 목적은

  ► 내가 뭐라꼬 그랫는지 ⇨ 내가 뭐라고 그랬는지
  ▸ 뭐라꼬 ⇐ '뭐라꼬'는 '뭐라고'의 방언입니다.
  ▸ 그랫는지 ⇐ '그랫는지'의 옳은 표기는 '그랬는지'입니다.

  ► 이놈의 봇시키가 알지 모를지
  ▸ 봇시키가 ⇐ 맞춤법 오류가 의심되는 구절입니다.

  ► 그냥 테스트해 보는 거여따. ⇨ 그냥 테스트해 보는 거였다.
  ▸ 거여따. ⇐ 올바르지 않은 어미의 사용입니다. '거였다.'로 고쳐 씁니다.
*/
func buildResultMessage(result checkResult) string {
	var message string

	var guide, corrected string

	for _, s := range result.Sentences {
		guide = ""
		corrected = s.Sentence

		for _, r := range s.Result {
			if r.Etype != "no_error" {
				guide += fmt.Sprintf("▸ %s ⇐ %s\n", r.Input, strings.Join(r.Help, " "))

				corrected = strings.Replace(corrected, r.Input, r.Output, 1)
			}
		}

		if corrected == s.Sentence { // no correction
			message += fmt.Sprintf("► _%s_\n%s\n", s.Sentence, guide)
		} else {
			message += fmt.Sprintf("► _%s_ ⇨ *%s*\n%s\n", s.Sentence, corrected, guide)
		}
	}

	return message
}

func main() {
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

	client := bot.NewClient(conf.TelegramApiToken)
	client.Verbose = conf.IsVerbose

	// get info about this bot
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
								if result, err := checkGrammar(conf.DaumApiKey, *update.Message.Text); err == nil {
									message = buildResultMessage(result)
								} else {
									message = fmt.Sprintf("API 호출 오류: %s", err)

									log.Printf("*** failed to call api: %s", err)
								}
							}
						} else {
							message = WelcomeMessage
						}

						// send message back (with markup support)
						if sent := b.SendMessage(
							update.Message.Chat.Id,
							&message,
							map[string]interface{}{
								"parse_mode": bot.ParseModeMarkdown,
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
