package utils

type Locale string

func (l Locale) ToString() string {
	return string(l)
}

const (
	EnLocale      = Locale("en")
	EsLocale      = Locale("es")
	ItLocale      = Locale("it")
	DefaultLocale = ItLocale
)

var supported = map[Locale]struct{}{
	ItLocale: struct{}{},
	EsLocale: struct{}{},
	EnLocale: struct{}{},
}

func SupportedLocale(l Locale) bool {
	_, ok := supported[l]

	return ok
}

func GetSupportedLocales() map[Locale]struct{} {
	return supported
}

func LocaleFromString(lang string) Locale {
	ll := Locale(lang)
	if SupportedLocale(ll) {
		return ll
	}

	return DefaultLocale
}
