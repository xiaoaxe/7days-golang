package main

import (
	"fmt"
	"net/http"

	"github.com/xiaoaxe/7days-golang/axe-web/axeweb"
)

// curl http://127.0.0.1:8888/v1/greeting/xiaobao
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

	r.GET("/v1/greeting/:name", func(c *axeweb.Context) {
		c.String(http.StatusOK, fmt.Sprintf("We Got name: %v\n", c.Params["name"]))
	})

	r.Run(":8888")
}
