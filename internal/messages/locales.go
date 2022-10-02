package messages

type Locale string

const (
	LocaleUK Locale = "uk"
	LocaleEN Locale = "en"
)

func localeExist(locale Locale) bool {
	switch locale {
	case LocaleUK:
		return true
	case LocaleEN:
		return true
	}

	return false
}
