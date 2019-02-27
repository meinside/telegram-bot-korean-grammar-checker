# 한글 맞춤법 검사 Telegram Bot

_(2019년 2월 27일 현재, HTTP 404 오류로 API가 더 이상 작동하지 않음)_

_(아래 API는 deprecate 되어 더 이상 새로 사용이 불가능합니다. 이미 쓰고 있던 경우는 2016년 10월 11일 현재까지는 잘 되는 듯 합니다.)_

이 봇은 Daum Kakao Corp.의 [맞춤법 검사 API](https://developers.daum.net/services/apis/grammar-checker/v1/check.json)를 활용, 입력된 문장에 대해 맞춤법 검사를 수행하여 그 결과를 응답해주는 Go로 개발된 Telegram Bot입니다.

![screen shot 2016-08-17 at 15 31 56](https://cloud.githubusercontent.com/assets/185988/17726738/df9505fe-648f-11e6-9677-a28b76eb8bec.png)

## 설치

설치를 위해 Go가 먼저 설치/설정되어 있어야 합니다.

```
$ go get -d github.com/meinside/telegram-bot-korean-grammar-checker
$ cd $GOPATH/src/github.com/meinside/telegram-bot-korean-grammar-checker
$ go build
```

## 설정

### Telegram Bot 토큰 생성

[가이드](https://core.telegram.org/bots#6-botfather) 참고.

### 설정 파일 생성

샘플 설정 파일을 복사한 뒤,

```bash
$ cp config.json.sample config.json
```

```json
{
	"telegram_api_token": "0123456789:abcdefghijklmnopqrstuvwyz-x-0a1b2c3d4e",
	"daum_api_key": "abcd1234efgh567890ijklmn",
	"monitor_interval": 1,
	"is_verbose": false
}
```

*telegram_api_token*, *daum_api_key* 등의 값을 본인의 설정에 맞게 수정하십시오.

*monitor_interval*은 Telegram Bot API로부터 새로 받은 메시지를 가져오는 주기(초)로, 짧을수록 빠르게 응답합니다.

*is_verbose*를 true로 설정하면 장황한 디버깅용 로그 메시지를 볼 수 있습니다.

## 실행

```bash
$ ./telegram-bot-korean-grammar-checker
```

### 서비스로 실행 (linux systemd)

```bash
$ sudo cp systemd/telegram-bot-korean-grammar-checker.service /lib/systemd/system/
$ sudo vi /lib/systemd/system/telegram-bot-korean-grammar-checker.service
```

**User**, **Group**, **WorkingDirectory**, **ExecStart** 값을 각각 환경에 맞게 수정합니다.

다음의 명령으로 부팅 시 자동으로 서비스로 실행됩니다:

```bash
$ sudo systemctl enable telegram-bot-korean-grammar-checker.service
```

서비스 시작/중단은 다음과 같습니다:

```bash
$ sudo systemctl start telegram-bot-korean-grammar-checker.service
$ sudo systemctl restart telegram-bot-korean-grammar-checker.service
$ sudo systemctl stop telegram-bot-korean-grammar-checker.service
```

## 활용 예

_(2019년 2월 27일 현재, API가 작동하지 않아 봇도 서비스를 중단합니다)_

* [@KoreanGrammarCheckerBot](https://telegram.me/KoreanGrammarCheckerBot): 개인 서버에서 서비스 중 **(예고 없이 중단될 수 있으며, 그 경우 응답이 없을 수 있음)**

## TODO

- [ ] test/benchmark 코드 추가
- [ ] 다른 open api 활용 옵션 추가
- [ ] 기타 옵션 추가

## License

MIT

