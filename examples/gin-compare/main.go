package main

import (
	"github.com/gin-gonic/gin"
	"github.com/itrabbit/just"
	"os"
)

func main() {
	isGin := false
	for _, arg := range os.Args {
		if arg == "--gin" {
			isGin = true
			break
		}
	}
	if isGin {
		s0 := gin.Default()
		s0.GET("/welcome/:name", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"param_name": c.Param("name"),
				"framework":  "gin",
				"version":    gin.Version})
		})
		s0.Run(":8001")
	} else {
		s1 := just.New()
		s1.GET("/welcome/{name}", func(c *just.Context) just.IResponse {
			return just.JsonResponse(200, just.H{
				"param_name": c.ParamDef("name", "unknown"),
				"framework":  "just",
				"version":    just.Version})
		})
		s1.Run(":8001")
	}
}
