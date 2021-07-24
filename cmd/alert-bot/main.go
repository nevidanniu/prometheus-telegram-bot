package main

import (
	"alert-bot/internal/templateFormatters/alertmanager"
	"alert-bot/internal/templateFormatters/defaultformatter"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
	"net/http"
	"strconv"
)

func SplitMessage(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func main() {
	fmt.Sprint("lol")
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	cfg = initFlags()

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	err := loadAllTemplatesFromPath(cfg.TemplateFolder, 0, nil)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Debug {
		go trackTemplateChanges(cfg.TemplateFolder)
	}

	bot_tmp, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal(err)
	}
	bot = bot_tmp
	log.Infof("Authorised on account %s", bot.Self.UserName)
	go telegramBot(bot)

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.POST("/:templatename/:chatid", POST_Handling)
	router.Run(":9087")

}

func POST_Handling(c *gin.Context) {
	var chunks []string
	chatID, err := strconv.ParseInt(c.Param("chatid"), 10, 64)
	if err != nil {
		log.Errorf("Cat't parse chat id: %q", c.Param("chatid"))
		c.JSON(http.StatusBadRequest, gin.H{
			"err": fmt.Sprintf("no such channel: %v", c.Param("chatid")),
		})
		return
	}

	// Find template in library for request
	templateName := c.Param("templatename")
	tmpH, ok := botTemplates.Load(templateName)
	if !ok {
		err = fmt.Errorf("no template with given name found: %v", templateName)
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"err": fmt.Sprint(err),
		})
		return
	}

	// Check for special formatters like alertmanager messages
	switch templateName {
	case "alertmanager":
		alertmanagerMessage := &alertmanager.Message{}
		chunks, err = formatTemplate(alertmanagerMessage, c, tmpH)
		if err != nil {
			log.Errorf("Error sending message: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"err": fmt.Sprint(err),
			})
			msg := tgbotapi.NewMessage(chatID, "Error sending message, checkout logs")
			bot.Send(msg)
			return
		}
	default:
		defaultMessage := &defaultformatter.Message{}
		chunks, err = formatTemplate(defaultMessage, c, tmpH)
		if err != nil {
			log.Errorf("Error sending message: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"err": fmt.Sprint(err),
			})
			msg := tgbotapi.NewMessage(chatID, "Error sending message, checkout logs")
			bot.Send(msg)
			return
		}
	}

	if chunks == nil {
		err = fmt.Errorf("Error sending message: no messages generated with selected template")
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"err": fmt.Sprint(err),
		})
		return
	}

	for _, chunk := range chunks {
		for _, subString := range SplitMessage(chunk, cfg.SplitMessageBytes) {
			msg := tgbotapi.NewMessage(chatID, subString)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.DisableWebPagePreview = true
			sendmsg, err := bot.Send(msg)
			if err == nil {
				c.String(http.StatusOK, "telegram msg sent.")
			} else {
				log.Errorf("Error sending message: %s", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"err":     fmt.Sprint(err),
					"message": sendmsg,
					"srcmsg":  fmt.Sprint(subString),
				})
				msg := tgbotapi.NewMessage(chatID, "Error sending message, checkout logs")
				bot.Send(msg)
			}
		}
	}
}
