package translations

import (
	"os"

	"github.com/kataras/i18n"
)

type Translator struct {
	i18n   *i18n.I18n
	prefix string
}

func NewTranslator(localesDir, defaultLanguage, botPrefix string) (*Translator, error) {
	dirDirs, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	dirs = append(dirs, defaultLanguage)

	for _, dir := range dirDirs {
		if !dir.IsDir() {
			continue
		}
		if defaultLanguage == dir.Name() {
			continue
		}

		dirs = append(dirs, dir.Name())
	}

	i, err := i18n.New(
		i18n.Glob(
			localesDir+"/*/*", i18n.LoaderConfig{},
		), dirs...,
	)
	if err != nil {
		return nil, err
	}

	return &Translator{
		i18n:   i,
		prefix: botPrefix,
	}, nil
}

func (t *Translator) GetMessage(lang, id string, args ...interface{}) string {
	return t.prefix + " " + t.i18n.Tr(lang, id, args...)
}
