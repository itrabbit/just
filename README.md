# JUST Web Framework

<img align="right" width="159px" src="https://raw.githubusercontent.com/itrabbit/just/master/logo.png">

[![Build Status](https://travis-ci.org/itrabbit/just.svg?branch=master)](https://travis-ci.org/itrabbit/just)
 [![CodeCov](https://codecov.io/gh/itrabbit/just/branch/master/graph/badge.svg)](https://codecov.io/gh/itrabbit/just)
 [![GoDoc](https://godoc.org/github.com/itrabbit/just?status.svg)](https://godoc.org/github.com/itrabbit/just)

**JUST** — Web application framework, written in Go (GoLang). Inspired by the Gin (GoLang) and Symfony (PHP). JUST was not created to handle the huge volume of data and the more pre-empting analogues (Gin, Iris, Martini, ...). First I want to achieve comfort and reducing product development time ;-)  

> Ping / Pong example

```go
package main

import "github.com/itrabbit/just"

func main() {
	a := just.New()
	a.GET("/ping", func(c *just.Context) just.IResponse {
		return c.S().Response(200, just.H{
			"message": "pong",
		})
	})
	a.Run(":80")
}
```

## CLI 

> Install JUST CLI

```sh
go install github.com/itrabbit/just/cli/just-cli
```

### Build i18n file from source

```sh
just-cli i18n:build -lang="en,ru" -dir="{full path to project dir}" -out="i18n.go"
```

> Example result i18n:build (i18n.go)

```go
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
			"Hello World": "Hello World",
			"Payload": "Payload",
		})
	}
}
```

> Usage i18n (main.go)

```go
package main

import "github.com/itrabbit/just"

func main() {
	// Create app
	app := just.New()

	// Use i18n.go
	loadTranslations(app.Translator())
    
	app.GET("", func(c *just.Context) just.IResponse {
		return c.Serializer().
			Response(200, &just.H{
				"msg":     c.Tr("Hello World"),
				"payload": c.Tr("Payload"),
			})
	})
	app.Run(":8081")
}
```

## Performance testing (on MacBook Pro 15 (2014))

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

## Serialize result finishing with filtration fields

```go
package main

import (
	"time"
	
	"github.com/itrabbit/just"	
	"github.com/itrabbit/just/components/finalizer"
)

type PhoneNumber struct{
	E164 string `json:"e164"` 
}

type User struct {
	ID          uint64          `json:"id"`
	Phone       *PhoneNumber    `json:"phone,omitempty" group:"private" export:"E164"`
	CreatedAt   time.Time       `json:"created_at" group:"private"`
	UpdatedAt   time.Time       `json:"update_at" group:"private" exclude:"equal:CreatedAt"`
}

func main() {
	// Create new JUST application
	a := just.New()
	
	// replace def serializers
	finalizer.ReplaceSerializers(a)	
        
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
    a.Run(":80")
}
```

> Result GET request http://localhost/public

```json
{
    "id": 1
}
```

> Result GET request http://localhost/private

```json
{
    "id": 1,
    "phone": "+79000000000",
    "created_at": "2017-12-11T22:23:36.709146+05:00"    
}
```

[More info](/components/finalizer/README.md)

## Routing examples

```
// Enums
http://localhost/api/{type:enum(object,item)}/{id:integer}

// True / False (0,1,t,f,true,false)
http://localhost/api/trigger/{value:boolean}

// Integer
http://localhost/api/object/{id:integer}

// Float
http://localhost/api/object/{id:float}

// Regexp
http://localhost/api/object/{id:regexp(\\d+)}

// String
http://localhost/api/object/{name}

// UUID
http://localhost/api/object/{uuid:uuid}
```

# Donation to development

`BTC: 1497z5VaY3AUEUYURS5b5fUTehVwv7wosX`

`DASH: XjBr7sqaCch4Lo1A7BctQz3HzRjybfpx2c`

`XRP: rEQwgdCr8Jma3GY2s55SLoZq2jmqmWUBDY`

`PayPal / Yandex Money: garin1221@yandex.ru`

## License

JUST is licensed under the [MIT](LICENSE).