package logger

import (
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"io/ioutil"
	"os"
)

var Log *logrus.Logger

func init() {

	Log = logrus.New()
	Log.SetOutput(ioutil.Discard)

	level, err := logrus.ParseLevel(config.Loglevel)
	if err != nil {
		Log.SetLevel(logrus.InfoLevel)
		Log.Error("Error when parsing logrus level from config")
	} else {
		Log.SetLevel(level)
	}

	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	Log.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})

	Log.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})

	Log.AddHook(lfshook.NewHook(
		config.LogLocation,
		&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		},
	))
}
