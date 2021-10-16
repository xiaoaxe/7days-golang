// recovery middleware
//@author: baoqiang
//@time: 2021/10/15 22:03:07
package axeweb

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandleFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				//TODO impl tracelog
				// debug.PrintStack()
				msg := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(msg))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

func trace(msg string) string {
	var pcs []uintptr
	n := runtime.Callers(3, pcs[:])

	var sb strings.Builder
	sb.WriteString(msg + "\nTraceback: ")

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s: %d", file, line))
	}

	return sb.String()
}
