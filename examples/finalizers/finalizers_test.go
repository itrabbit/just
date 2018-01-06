package main

import (
	"time"

	"github.com/itrabbit/just"
	"github.com/itrabbit/just/components/cors"
	"github.com/itrabbit/just/components/finalizer"
)

// Структура телефонного номера
type PhoneNumber struct {
	E164 string `json:"e164"`
}

// Структура пользователя
type User struct {
	ID        uint64       `json:"id"`
	Phone     *PhoneNumber `json:"phone,omitempty" group:"private" export:"E164"`
	CreatedAt time.Time    `json:"created_at" group:"private"`
	UpdatedAt time.Time    `json:"update_at" group:"private" exclude:"equal:CreatedAt"`
}

func ExampleFinalizers() {
	// Создаем чистое приложение (без сериализаторов и шаблонизаторов)
	app := just.NewClear()

	// Добавляем необходимые сериализаторы
	app.SerializerManager().
		SetSerializer("json", []string{
			"application/json",
		}, finalizer.NewJsonSerializer("utf-8")).
		SetSerializer("xml", []string{
			"text/xml",
			"application/xml",
		}, finalizer.NewXmlSerializer("utf-8"))

	// Добавляем поддержку CORS
	app.Use(cors.Middleware(cors.Options{
		AllowHeaders: []string{
			"Origin", "Accept", "Content-Type",
			"Authorization", "X-Auth", "Token",
			"Connection", "Upgrade", "Sec-Websocket-Version",
			"Sec-Websocket-Key", "Sec-Websocket-Protocol", "Sec-Websocket-Accept",
		},
	}))

	// Обработка GET запроса
	app.GET("/{group:enum(public,private)}", func(c *just.Context) just.IResponse {
		now := time.Now()
		return c.Serializer().
			Response(200, finalizer.Input(
				&User{
					ID: 1,
					Phone: &PhoneNumber{
						E164: "+79000000000",
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
				c.ParamDef("group", "public"),
			))
	})

	// Запускаем приложеие
	app.Run(":8081")
}
