package router

import (
	"context"
	"net/http"
	"strings"
)

/*
route table:
	HTTPMethod => Controllers tree
*/
type Route struct {
	table map[string]*routeNode
}

func NewRoute() *Route {
	route := &Route{
		make(map[string]*routeNode),
	}
	route.table[http.MethodPost] = newNode()
	route.table[http.MethodGet] = newNode()
	return route
}

// nodes of route tree
type routeNode struct {
	children map[string]*routeNode // child nodes
	wildcard *routeNode
	param    string
	handler  http.HandlerFunc
}

func newNode() *routeNode {
	return &routeNode{
		children: make(map[string]*routeNode),
		wildcard: nil,
		param:    "",
		handler:  nil,
	}
}

func (root *routeNode) add(name string, handler http.HandlerFunc) {
	paths := strings.Split(name, "/")
	node := root
	for _, path := range paths {
		if path == "" {
			continue
		}
		if child, ok := node.children[path]; ok {
			node = child
			continue
		} else { // create a child node
			// if it is a parameter
			if path[:1] == "#" {
				// we just override it if it exists
				node.wildcard = newNode()
				node = node.wildcard
				node.param = path[1:]
			} else {
				n := newNode()
				node.children[path] = n
				node = n
			}
		}
	}
	node.handler = handler
}

func (route *Route) POST(name string, handler http.HandlerFunc) {
	route.table[http.MethodPost].add(name, handler)
}

func (route *Route) GET(name string, handler http.HandlerFunc) {
	route.table[http.MethodGet].add(name, handler)
}

// serve http will forward the request to a given handler
func (route *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method // the HTTP method (GET, POST, PUT, etc.).
	//path := r.URL.Path
	query := make(map[string]string)
	for k, v := range r.Form {
		query[k] = strings.Join(v, "")
	}

	// parse path
	table := route.table[method]
	param := make(map[string]string)
	for _, path := range strings.Split(r.URL.Path, "/") {
		if path == "" {
			continue
		}
		if child, ok := table.children[path]; ok {
			table = child
		} else { // match parameters
			if table.wildcard == nil {
				http.NotFound(w, r)
				return
			} else {
				table = table.wildcard
				param[table.param] = path
			}
		}
	}
	if len(param) == 0 {
		table.handler(w, r)
	} else {
		ctx := context.WithValue(r.Context(), "param", param)
		table.handler(w, r.WithContext(ctx))
	}
	//http.NotFound(w, r)
	return
}
