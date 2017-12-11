# JUST Web Framework

<img align="right" width="159px" src="https://raw.githubusercontent.com/itrabbit/just/logo.png">

JUST — веб фреймворк, написанный на Go (GoLang). Вдохновлен Gin (GoLang) и Symfony (PHP). JUST не создавался с целью обработки огромных обьемов данных и тем более опережению аналогов (Git, Iris, ...) по скорости работы. В первую очерь хочется добиться удобства и уменьшения времени на разработку ;-)  

> Ping / Pong пример

```go
package main

import "github.com/itrabbit/just"

func main() {
	a := just.New()
	a.GET("/ping", func(c *just.Context) just.IResponse {
		return c.Serializer().Response(200, just.H{
			"message": "pong",
		})
	})
	a.Run(":80")
}
```

## Тестирование производительности (на MacBook Pro 15 (2014))

```
goos: darwin
goarch: amd64
pkg: github.com/itrabbit/just

BenchmarkOneRoute:
10000000	       183 ns/op	      48 B/op	       1 allocs/op

BenchmarkRecoveryMiddleware:
10000000	       176 ns/op	      48 B/op	       1 allocs/op

BenchmarkLoggerMiddleware:
10000000	       173 ns/op	      48 B/op	       1 allocs/op

BenchmarkManyHandlers:
10000000	       174 ns/op	      48 B/op	       1 allocs/op

Benchmark5Params:
  500000	      3708 ns/op	     720 B/op	       8 allocs/op
  
BenchmarkOneRouteJSON:
 1000000	      1059 ns/op	     592 B/op	       6 allocs/op
 
BenchmarkOneRouteHTML:
  300000	      4501 ns/op	    1840 B/op	      28 allocs/op
  
BenchmarkOneRouteSet:
 3000000	       420 ns/op	     384 B/op	       3 allocs/op
 
BenchmarkOneRouteString:
10000000	       220 ns/op	      80 B/op	       2 allocs/op

BenchmarkManyRoutesFist:
10000000	       175 ns/op	      48 B/op	       1 allocs/op

BenchmarkManyRoutesLast:
10000000	       194 ns/op	      48 B/op	       1 allocs/op

Benchmark404:
10000000	       169 ns/op	      48 B/op	       1 allocs/op

Benchmark404Many:
 2000000	       607 ns/op	      48 B/op	       1 allocs/op
```

## Финализаця резульата сериализации с фильтрацией полей

[Подробнее](/components/finalizer/README.md)

## Примеры роутинга

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
