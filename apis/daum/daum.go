package daum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	apiUrl = "https://apis.daum.net/grammar-checker/v1/check.json"
)

// api result
type CheckResult struct {
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
func CheckGrammar(apiKey, text string) (CheckResult, error) {
	client := &http.Client{}

	var err error
	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequest("GET", apiUrl, nil); err == nil {
		q := req.URL.Query()
		q.Add("apikey", apiKey)
		q.Add("query", text)
		q.Add("help", "on")
		req.URL.RawQuery = q.Encode()

		if res, err = client.Do(req); err == nil {
			defer res.Body.Close()

			var bytes []byte
			if bytes, err = ioutil.ReadAll(res.Body); err == nil {
				var result CheckResult
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

	return CheckResult{}, err
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
func BuildResultMessage(result CheckResult) string {
	var message string

	var guide, corrected string

	for _, s := range result.Sentences {
		guide = ""

		if len(s.Sentence) <= 0 { // skip empty lines
			continue
		}

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
