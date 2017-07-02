# JUST

GoLang package for fast development micro services

## Параметры роутинга

```
// Перечисление
http://localhost/api/{type:enum(object,item)}/{id:integer}

// Число
http://localhost/api/object/{id:integer}

// Регулярное выражение
http://localhost/api/object/{id:regexp(\\d+)}

// Любая строка
http://localhost/api/object/{name}

// UUID
http://localhost/api/object/{uuid:uuid}
```
