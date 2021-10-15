// recovery middleware
//@author: baoqiang
//@time: 2021/10/15 22:03:07
package axeweb

import (
	"net/http"
	"runtime/debug"
)

func Recovery() HandleFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				//TODO impl tracelog
				debug.PrintStack()
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}
