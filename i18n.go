package just

import (
	"fmt"
	"sync"
)

// Translation map for one language.
type TranslationMap map[string]string

// Translator interface.
type ITranslator interface {
	DefaultLocale() string
	SetDefaultLocale(locale string) ITranslator
	AddTranslationMap(locale string, m TranslationMap) ITranslator
	Trans(locale string, message string, vars ...interface{}) string // Translate text for locale and vars.
}

type baseTranslator struct {
	sync.RWMutex
	defaultLocale string
	localizations map[string]TranslationMap
}

func (t *baseTranslator) DefaultLocale() string {
	t.Lock()
	defer t.Unlock()

	return t.defaultLocale
}

func (t *baseTranslator) SetDefaultLocale(locale string) ITranslator {
	t.RLock()
	defer t.RUnlock()

	t.defaultLocale = locale
	return t
}

func (t *baseTranslator) AddTranslationMap(locale string, m TranslationMap) ITranslator {
	t.RLock()
	defer t.RUnlock()

	if t.localizations == nil {
		t.localizations = make(map[string]TranslationMap)
	}
	if _, ok := t.localizations[locale]; !ok {
		t.localizations[locale] = make(TranslationMap)
	}
	if m != nil {
		for key, value := range m {
			t.localizations[locale][key] = value
		}
	}
	return t
}

func (t *baseTranslator) Trans(locale string, message string, vars ...interface{}) string {
	if t.localizations != nil {
		if m, ok := t.localizations[locale]; ok && m != nil {
			if transMessage, ok := m[message]; ok {
				message = transMessage
			}
		}
	}
	if len(vars) > 0 {
		return fmt.Sprintf(message, vars...)
	}
	return message
}
