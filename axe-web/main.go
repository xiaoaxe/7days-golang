package main

import (
	"net/http"

	"github.com/xiaoaxe/7days-golang/axe-web/axeweb"
)

func main() {
	r := axeweb.Default()

	r.GET("/", func(c *axeweb.Context) {
		c.String(http.StatusOK, "Hello axeweb\n")
	})

	r.GET("/panic", func(c *axeweb.Context) {
		names := []string{"axeweb"}
		c.JSON(http.StatusOK, axeweb.H{
			"name": names[1],
		})
	})

	r.Run(":8888")
}
