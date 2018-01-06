### Финализация моделей данных (фильтрация полей модели в зависомости от тегов)

> Теги финализации:

`group:"group1,group2...groupN"` - группы при которых поле включается в результирующую модель 

`exclude:"equal:FileName"` - условия исключения поля (equal - сравнение значения текущего поля с значение поля FileName в текущей модели данных)

`export:"FileName"` -  используется для замены значения поля значение поля вложенной структуры (только для структур)

> Пример:

```go
package main

import (
	"time"
	
	"github.com/itrabbit/just"	
	"github.com/itrabbit/just/components/finalizer"
)

// Структура телефонного номера
type PhoneNumber struct{
	E164 string `json:"e164"` 
}

// Структура пользователя
type User struct {
	ID          uint64          `json:"id"`
	Phone       *PhoneNumber    `json:"phone,omitempty" group:"private" export:"E164"`
	CreatedAt   time.Time       `json:"created_at" group:"private"`
	UpdatedAt   time.Time       `json:"update_at" group:"private" exclude:"equal:CreatedAt"`
}

func main() {
	// Создаем приложение
	a := just.New()
	
	// Производим замену стандартных сериализаторов
	finalizer.ReplaceSerializers(a)	
    
    // Обработка GET запроса
    a.GET("/{group:enum(public,private)}", func(c *just.Context) just.IResponse {
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
 
    // Запускаем приложение
    a.Run(":80")
}
```

> Результат GET запроса http://localhost/public

```json
{
    "id": 1
}
```

> Результат GET запроса http://localhost/private

```json
{
    "id": 1,
    "phone": "+79000000000",
    "created_at": "2017-12-11T22:23:36.709146+05:00"    
}
```