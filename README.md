# JUST 

Пакет для быстрой и удобной разработки Web микросервисов


## Параметры роутинга

```
// Перечисление
http://localhost/api/{type:enum(object,item)}/{id:integer}

// True / False ( поддерживаются значения: 0,1,t,f,true,false)
http://localhost/api/trigger/{value:boolean}

// Число
http://localhost/api/object/{id:integer}

// Число с плавающей точкой
http://localhost/api/object/{id:float}

// Регулярное выражение
http://localhost/api/object/{id:regexp(\\d+)}

// Любая строка
http://localhost/api/object/{name}

// UUID
http://localhost/api/object/{uuid:uuid}
```
