package main

import (
	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
	//"flag"
)

type Config struct {
	Debug             bool
	TelegramToken     string
	TemplateFolder    string
	TimeZone          string
	TimeOutFormat     string
	SplitChart        string
	SplitMessageBytes int
}

var cfg *Config

func initFlags() (cfg *Config) {
	cfg = &Config{}
	flag.BoolVar(&cfg.Debug, "debug", false, "enable debug mod")
	flag.StringVar(&cfg.TelegramToken, "tg-token", "", "telegram token (env: TG_TOKEN)")
	flag.StringVar(&cfg.TemplateFolder, "template-folder", "templates", "path to the folder containing .tmpl templates (env: TEMPLATE_FOLDER)")
	flag.StringVar(&cfg.TimeZone, "time-zone", "Europe/Moscow", "time zone (env: TIME_ZONE)")
	flag.StringVar(&cfg.TimeOutFormat, "time-outdata", "02/01/2006 15:04:05", "time format (env: TIME_OUTDATA)")
	flag.StringVar(&cfg.SplitChart, "split-chart", "|", "separator character for str_Format_MeasureUnit function (env: SPLIT_CHART)")
	flag.IntVar(&cfg.SplitMessageBytes, "message-bytes", 4000, "max bytes per message (env: MESSAGE_BYTES)")
	flag.Parse()

	if cfg.TelegramToken == "" {
		log.Fatalf("-tg-token: telegram token must be specified")
	}
	return
}
