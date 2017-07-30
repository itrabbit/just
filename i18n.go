package just

import (
	"fmt"
	"sync"
)

type TranslationMap map[string]string

type ITranslator interface {
	DefaultLocale() string
	SetDefaultLocale(locale string) ITranslator
	AddTranslationMap(locale string, m TranslationMap) ITranslator
	Trans(locale string, message string, vars ...interface{}) string
}

type baseTranslator struct {
	sync.RWMutex
	defaultLocale string
	localisations map[string]TranslationMap
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

	if t.localisations == nil {
		t.localisations = make(map[string]TranslationMap)
	}
	if _, ok := t.localisations[locale]; !ok {
		t.localisations[locale] = make(TranslationMap)
	}
	if m != nil {
		for key, value := range m {
			t.localisations[locale][key] = value
		}
	}
	return t
}

func (t *baseTranslator) Trans(locale string, message string, vars ...interface{}) string {
	if t.localisations == nil {
		if m, ok := t.localisations[locale]; ok && m != nil {
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
