// log mw
//@author: baoqiang
//@time: 2021/10/15 21:54:59
package axeweb

import (
	"log"
	"time"
)

func Logger() HandleFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
