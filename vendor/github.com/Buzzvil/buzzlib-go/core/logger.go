package core

import (
	"os"

	pluginLogrus "github.com/Buzzvil/buzzlib-go/plugin/logrus"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

//Gets application logger
func getLogger(newLogger bool) *logrus.Logger {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 03:04:05",
	}
	var log *logrus.Logger
	if newLogger {
		log = logrus.New()
	} else {
		log = logrus.StandardLogger()
	}
	log.Formatter = formatter
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	return log
}

func registerSentryHook(log *logrus.Logger, client *raven.Client) {
	hook := pluginLogrus.NewSentryHook(client, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	log.Hooks.Add(hook)
}

// Customize logger from config
func initLogger(log *logrus.Logger, config map[string]string) {
	if config["level"] != "" {
		lvl, err := logrus.ParseLevel(config["level"])
		if err == nil {
			log.SetLevel(lvl)
		}
	}
	if config["file"] != "" {
		file, err := os.OpenFile(config["file"], os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic("Failed to open log file" + config["file"])
		}
		log.Out = file
	}
	if config["formatter"] == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	} else if config["color"] == "true" {
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 03:04:05",
		}
	}
}
