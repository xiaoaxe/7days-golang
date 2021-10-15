// router
//@author: baoqiang
//@time: 2021/10/15 21:44:15
package axeweb

type router struct {
	roots    map[string]*node
	handlers map[string]HandleFunc
}

func newRouter() *router {
	return &router{
		roots:    map[string]*node{},
		handlers: map[string]HandleFunc{},
	}
}

// add & get route
func (r *router) addRoute(method string, pattern string, h HandleFunc) {

}

// get route with params
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	return nil, nil
}

func (r *router) getRoutes(method string, path string) []*node {
	return nil
}

// route helper
func parsePattern(pattern string) []string {
	return nil
}

// handle func
func (r *router) handle(c *Context) {

}
