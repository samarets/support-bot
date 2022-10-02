package messages

import "fmt"

type Helper struct {
	defaultLocale Locale
	botPrefix     string
}

func NewMessagesHelper(defaultLocale, botPrefix string) (*Helper, error) {
	if !localeExist(Locale(defaultLocale)) {
		return nil, fmt.Errorf("default locale is not found among existing locales")
	}

	return &Helper{
		defaultLocale: Locale(defaultLocale),
		botPrefix:     botPrefix,
	}, nil
}

func (h *Helper) GetMessage(code Code, locale string) string {
	loc := Locale(locale)
	if !localeExist(loc) {
		loc = h.defaultLocale
	}

	return h.botPrefix + " " + fmt.Sprintf(LocalMessages[code][loc])
}
