package alertmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"strings"
)

type Message struct{}

func (m *Message) FormatTemplate(c *gin.Context, tmpH *template.Template) ([]string, error) {
	var chunks []string
	var bytesBuff bytes.Buffer

	writer := io.Writer(&bytesBuff)

	mesg := map[string]interface{}{}
	err := binding.JSON.Bind(c.Request, &mesg)
	if err != nil {
		return nil, err
	}

	jsonMesg, err := json.MarshalIndent(&mesg, "", "    ")
	if err != nil {
		return nil, err
	}
	log.Infof("+------------------  A L E R T  J S O N  -------------------+")
	fmt.Printf("%s\r\n", jsonMesg)
	log.Infof("+-----------------------------------------------------------+\n\n")

	err = tmpH.Execute(writer, &mesg)
	if err != nil {
		return nil, err
	}
	message := bytesBuff.String()
	tmpChunks := strings.Split(message, "[delimiter]")
	for _, chunk := range tmpChunks {
		chunk = strings.TrimRightFunc(chunk, func(c rune) bool {
			//In windows newline is \r\n
			return c == '\r' || c == '\n'
		})
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
	}

	for _, chunk := range chunks {
		// Print in Log result message
		log.Infof("+---------------  F I N A L   M E S S A G E  ---------------+")
		fmt.Printf("%s\r\n", chunk)
		log.Infof("+-----------------------------------------------------------+")
	}
	return chunks, nil
}
