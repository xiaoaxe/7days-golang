// context
//@author: baoqiang
//@time: 2021/10/15 21:42:06
package axeweb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]string

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request

	// req&resp vars
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int

	// internal vars
	handlers []HandleFunc
	index    int

	// refs
	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		// vars
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// core func
func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"msg": err})
}

// req vars
func (c *Context) Param(key string) string {
	val, _ := c.Params[key]
	return val
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// write impl
func (c *Context) SetHeader(k, v string) {
	c.Writer.Header().Add(k, v)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// renders
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
