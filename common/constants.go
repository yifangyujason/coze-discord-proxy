package common

import (
	"os"
	"strings"
	"time"
)

var ProxySecrets = strings.Split(os.Getenv("PROXY_SECRET"), ",")
var RequestOutTime = os.Getenv("REQUEST_OUT_TIME")
var StreamRequestOutTime = os.Getenv("STREAM_REQUEST_OUT_TIME")

var DebugEnabled = os.Getenv("DEBUG") == "true"

var Version = "v4.0.8" // this hard coding will be replaced automatically when building, no need to manually change

const (
	RequestIdKey = "X-Request-Id"
	OutTime      = "out-time"
)

// Shouldn't larger then RateLimitKeyExpirationDuration
var (
	GlobalApiRateLimitNum            = 60
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = 60
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60
)

var RateLimitKeyExpirationDuration = 20 * time.Minute

var RequestOutTimeDuration = 18 * time.Second

var CozeErrorMsg = "Something wrong occurs, please retry. If the error persists, please contact the support team."

var CozeErrorMessages = []string{"You have exceeded the daily limit for sending messages to the bot. Please try again later.",
	"Something wrong occurs, please retry. If the error persists, please contact the support team.",
	"There are too many users now. Please try again a bit later.",
	"I'm sorry, but I can't assist with that."}

var DrawMessages = []string{"画",
	"绘",
	"draw"}

var FastModel = []string{"3.5"}

var CozeDailyLimitErrorMessages = []string{"You have exceeded the daily limit for sending messages to the bot. Please try again later."}
