package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"
)

type BotMessage interface {
	FormatTemplate(c *gin.Context, tmpH *template.Template) ([]string, error)
}

var botTemplates *BotTemplates

type BotTemplates struct {
	mx   sync.RWMutex
	tmpl map[string]*template.Template
}

func (b *BotTemplates) Load(key string) (*template.Template, bool) {
	b.mx.RLock()
	defer b.mx.RUnlock()
	val, ok := b.tmpl[key]
	return val, ok
}

func (b *BotTemplates) Store(key string, value *template.Template) {
	b.mx.Lock()
	defer b.mx.Unlock()
	if value != nil {
		b.tmpl[key] = value
	} else {
		delete(b.tmpl, key)
	}
}

func (b *BotTemplates) Len() int {
	return len(b.tmpl)
}

func NewBotTemplates() *BotTemplates {
	return &BotTemplates{
		tmpl: make(map[string]*template.Template),
	}
}

// Loads template on the given path and includes functions defined in FuncMap var
func loadTemplate(tmplPath string) (tmpH *template.Template, err error) {
	tmpH, err = template.New(path.Base(tmplPath)).Funcs(funcMap).ParseFiles(tmplPath)

	if err != nil {
		err = fmt.Errorf("problem reading parsing template file: %v", err)
	}
	return tmpH, err
}

// Load templates from config path folder. Searches .tmpl files and add it to global var botTemplates
func loadAllTemplatesFromPath(tmplFolder string, currRetryCount int, lastErr error) error {
	if currRetryCount > 2 && lastErr != nil {
		return lastErr
	}
	botTemplates = NewBotTemplates()
	files, err := ioutil.ReadDir(cfg.TemplateFolder)
	if err != nil {
		time.Sleep(time.Millisecond * 500)
		return loadAllTemplatesFromPath(tmplFolder, currRetryCount+1, err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmpl") {
			templatePath := tmplFolder + "/" + f.Name()
			tmpH, err := loadTemplate(templatePath)
			if err != nil {
				time.Sleep(time.Millisecond * 500)
				return loadAllTemplatesFromPath(tmplFolder, currRetryCount+1, err)
			}
			templateName := strings.Replace(f.Name(), ".tmpl", "", 1)
			botTemplates.Store(templateName, tmpH)
			log.Infof("loaded template: %s", f.Name())
		}
	}
	if botTemplates.Len() == 0 {
		return fmt.Errorf("no templates found for given path: %s", tmplFolder)
	}
	return nil
}

// Track for any changes by given path and reload global var botTemplates
func trackTemplateChanges(tmplFolder string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = watcher.Close()
	}()
	log.Infof("Adding file listener for path: %s", tmplFolder)
	err = watcher.Add(tmplFolder)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if strings.Contains(event.Name, ".tmpl") {
				log.Infof("modified file: %s", event.Name)
				err = loadAllTemplatesFromPath(tmplFolder, 0, nil)
				if err != nil {
					log.Fatal(err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("error: %v", err)
		}
	}
}

func formatTemplate(botMessage BotMessage, c *gin.Context, tmpH *template.Template) (chunks []string, err error) {
	chunks, err = botMessage.FormatTemplate(c, tmpH)
	return
}
