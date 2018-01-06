package main

import (
	"github.com/itrabbit/just"
)

func main() {
	// Создаем приложение
	app := just.New()

	// Загружаем перевод
	loadTranslations(app.Translator())

	// Обработка GET запроса
	app.GET("", func(c *just.Context) just.IResponse {
		return c.Serializer().
			Response(200, &just.H{
				"msg":     c.Tr("Hello World"),
				"payload": c.Tr("Payload"),
			})
	})

	// Запускаем приложеие
	app.Run(":8081")
}
