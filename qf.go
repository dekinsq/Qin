package qf

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type HandlerFunc func(*Context)

type (
	RouterGroup struct {
		prefix      string
		middlewares []HandlerFunc
		parent      *RouterGroup
		engine      *Engine
	}
	Engine struct {
		*RouterGroup
		router        *router
		groups        []*RouterGroup
		htmlTemplates *template.Template
		funcMap       template.FuncMap
		baseDirectory string
	}
)

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.baseDirectory = engine.setBaseDirectory()
	return engine
}

//func Default() *Engine {
//	engine := New()
//	engine.Use(plugins.Logger())
//	return engine
//}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	var AppPath string
	var appConfigPath string
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	workPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var filename = "app.conf"
	if os.Getenv("BEEGO_RUNMODE") != "" {
		filename = os.Getenv("BEEGO_RUNMODE") + ".app.conf"
	}
	appConfigPath = filepath.Join(workPath, "conf", filename)
	appConfigPath = filepath.Join(AppPath, "conf", filename)

	fmt.Printf(appConfigPath)

	absolutePath := path.Join(engine.baseDirectory, pattern)
	temp, _ := template.New("").Funcs(engine.funcMap).ParseGlob(absolutePath)
	if temp != nil {
		engine.htmlTemplates = template.Must(temp, nil)
	}
}

func (engine *Engine) LoadHTMLFiles(files ...string) {
	temp := template.Must(template.New("").Funcs(engine.funcMap).ParseFiles(files...))
	if temp != nil {
		engine.htmlTemplates = template.Must(temp, nil)
	}
}

func (engine *Engine) Run(addr string) (err error) {
	err = http.ListenAndServe(addr, engine)
	return
}

func (engine *Engine) RunTLS(addr, certFile, keyFile string) (err error) {
	err = http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	return
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func (engine *Engine) setBaseDirectory() string {
	var baseDirectory string
	if _, filename, _, ok := runtime.Caller(1); ok {
		baseDirectory = path.Dir(filename)
	}
	engine.baseDirectory = baseDirectory
	return baseDirectory
}

func (engine *Engine) getBaseDirectory() string {
	return engine.baseDirectory
}
