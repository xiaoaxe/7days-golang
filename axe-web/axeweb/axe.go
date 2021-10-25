//engine impl
//@author: baoqiang
//@time: 2021/10/15 21:41:28
package axeweb

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandleFunc func(*Context)

type (
	RouterGroup struct {
		prefix      string
		middlewares []HandleFunc
		parent      *RouterGroup //nesting
		engine      *Engine      // ref
	}

	Engine struct {
		*RouterGroup
		router *router
		groups []*RouterGroup

		// render html
		htmlTemplates *template.Template
		funcMap       template.FuncMap
	}
)

// public funcs
func New() *Engine {
	engine := &Engine{router: newRouter()}

	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}

	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// RouterGroup impls
func (g *RouterGroup) Use(hs ...HandleFunc) {
	g.middlewares = append(g.middlewares, hs...)
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix: g.prefix + prefix,
		parent: g,
		engine: engine,
	}
	// update engine groups
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// RouterGroup GET/POST method
func (g *RouterGroup) GET(pattern string, h HandleFunc) {
	g.addRoute("GET", pattern, h)
}

func (g *RouterGroup) POST(pattern string, h HandleFunc) {
	g.addRoute("GET", pattern, h)
}

func (g *RouterGroup) addRoute(method string, pattern string, h HandleFunc) {
	fullPattern := g.prefix + pattern
	log.Printf("Route: %4s - %s", method, fullPattern)
	g.engine.router.addRoute(method, fullPattern, h)
}

// RouterGroup static impl
func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileSever := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// serve
		fileSever.ServeHTTP(c.Writer, c.Req)
	}
}

func (g *RouterGroup) Static(relativePath string, root string) {
	h := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// reg handler
	g.GET(urlPattern, h)
}

// engine impls
func (e *Engine) setFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// 通过通配符的方式加载HTML文件
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var mws []HandleFunc

	// add mws
	for _, g := range e.groups {
		if strings.HasPrefix(req.URL.RequestURI(), g.prefix) {
			mws = append(mws, g.middlewares...)
		}
	}

	// run with ctx
	c := newContext(w, req)
	c.handlers = mws
	c.engine = e //ref
	e.router.handle(c)
}
