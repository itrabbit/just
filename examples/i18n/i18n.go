// The file is generated using the CLI JUST.
// Change only translation strings!
// Everything else can be removed when re-generating!
// - - - - - 
// Last generated time: Sun, 07 Jan 2018 00:56:22 +05

package main

import "github.com/itrabbit/just"

func loadTranslations(t just.ITranslator) {
	if t != nil {
		t.AddTranslationMap("en", just.TranslationMap{
			"Hello World": "Hello World",
			"Payload": "Payload",
		})
		t.AddTranslationMap("ru", just.TranslationMap{
			"Payload": "Нагрузка",
			"Hello World": "Привет мир",
		})
	}
}
