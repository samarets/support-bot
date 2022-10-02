package translations

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Translator struct {
	bundle *i18n.Bundle
}

func NewTranslations(defaultLanguage, languageDirPath string) (*Translator, error) {
	tag, err := language.Parse(defaultLanguage)
	if err != nil {
		return nil, err
	}

	bundle := i18n.NewBundle(tag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	files, err := os.ReadDir(languageDirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileSepName := strings.Split(file.Name(), ".")
		if len(fileSepName) < 3 {
			continue
		}
		if fileSepName[0] != "active" {
			continue
		}

		bundle.MustLoadMessageFile(languageDirPath + "/" + file.Name())
	}

	return &Translator{
		bundle: bundle,
	}, nil
}

type LocalizeBuilder struct {
	loc          *i18n.Localizer
	message      *i18n.Message
	templateData map[string]string
}

func (t *Translator) NewLocalize(lang, id string) *LocalizeBuilder {
	return &LocalizeBuilder{
		loc: i18n.NewLocalizer(t.bundle, lang),
		message: &i18n.Message{
			ID: id,
		},
		templateData: make(map[string]string),
	}
}

func (lb *LocalizeBuilder) AddMessage(message string) *LocalizeBuilder {
	lb.message.Other = message
	return lb
}

func (lb *LocalizeBuilder) AddTemplate(templateID, message string) *LocalizeBuilder {
	lb.templateData[templateID] = message
	return lb
}

func (lb *LocalizeBuilder) Localize() (string, error) {
	localizeConfig := &i18n.LocalizeConfig{
		DefaultMessage: lb.message,
	}
	if len(lb.templateData) > 0 {
		localizeConfig.TemplateData = lb.templateData
	}

	message, err := lb.loc.Localize(localizeConfig)
	if err != nil {
		return "", err
	}

	return message, nil
}
