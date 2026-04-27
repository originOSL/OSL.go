// name: serve
// description: Gin-like HTTP server framework for OSL
// author: Mist
// requires: encoding/json, sync, time, strings, fmt

type HttpContext struct {
	w http.ResponseWriter
	r *http.Request
}

func (c *HttpContext) String(code int, format string, values ...any) {
	c.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.w.WriteHeader(code)
	if len(values) > 0 {
		fmt.Fprintf(c.w, format, values...)
	} else {
		c.w.Write([]byte(format))
	}
}

func (c *HttpContext) JSON(code int, obj any) {
	c.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.w.WriteHeader(code)
	json.NewEncoder(c.w).Encode(obj)
}

func (c *HttpContext) HTML(code int, html string) {
	c.w.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.w.WriteHeader(code)
	c.w.Write([]byte(html))
}

func (c *HttpContext) Data(code int, contentType string, data []byte) {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	c.w.Write(data)
}

func (c *HttpContext) Redirect(code int, url string) {
	http.Redirect(c.w, c.r, url, code)
}

func (c *HttpContext) Status(code int) {
	c.w.WriteHeader(code)
}

func (c *HttpContext) Query(key string) string {
	return c.r.URL.Query().Get(key)
}

func (c *HttpContext) Param(key string) string {
	return c.r.PathValue(key)
}

func (c *HttpContext) FormValue(key string) string {
	return c.r.FormValue(key)
}

func (c *HttpContext) Header(key string) string {
	return c.r.Header.Get(key)
}

func (c *HttpContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *HttpContext) Method() string {
	return c.r.Method
}

func (c *HttpContext) Path() string {
	return c.r.URL.Path
}

func (c *HttpContext) Host() string {
	return c.r.Host
}

func (c *HttpContext) RemoteAddr() string {
	return c.r.RemoteAddr
}

func (c *HttpContext) Body() string {
	data, err := OSLio.ReadAll(c.r.Body)
	if err != nil {
		return ""
	}
	return string(data)
}

func (c *HttpContext) BindJSON() map[string]any {
	var obj map[string]any
	decoder := json.NewDecoder(c.r.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&obj); err != nil {
		return map[string]any{}
	}
	return obj
}

func (c *HttpContext) Cookie(name string) string {
	cookie, err := c.r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (c *HttpContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	http.SetCookie(c.w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *HttpContext) Request() *http.Request {
	return c.r
}

func (c *HttpContext) Writer() http.ResponseWriter {
	return c.w
}

type HttpRouter struct {
	mux        *http.ServeMux
	server     *http.Server
	middleware []func(*HttpContext) bool
	addr       string
}

type HTTPFramework struct{}

func (HTTPFramework) New() *HttpRouter {
	return &HttpRouter{
		mux: http.NewServeMux(),
	}
}

func (rt *HttpRouter) handle(method, pattern string, handler func(*HttpContext)) {
	fullPattern := method + " " + pattern
	rt.mux.HandleFunc(fullPattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := &HttpContext{w: w, r: r}
		for _, mw := range rt.middleware {
			if !mw(ctx) {
				return
			}
		}
		handler(ctx)
	})
}

func (rt *HttpRouter) GET(pattern string, handler func(*HttpContext)) {
	rt.handle("GET", pattern, handler)
}

func (rt *HttpRouter) POST(pattern string, handler func(*HttpContext)) {
	rt.handle("POST", pattern, handler)
}

func (rt *HttpRouter) PUT(pattern string, handler func(*HttpContext)) {
	rt.handle("PUT", pattern, handler)
}

func (rt *HttpRouter) PATCH(pattern string, handler func(*HttpContext)) {
	rt.handle("PATCH", pattern, handler)
}

func (rt *HttpRouter) DELETE(pattern string, handler func(*HttpContext)) {
	rt.handle("DELETE", pattern, handler)
}

func (rt *HttpRouter) OPTIONS(pattern string, handler func(*HttpContext)) {
	rt.handle("OPTIONS", pattern, handler)
}

func (rt *HttpRouter) HEAD(pattern string, handler func(*HttpContext)) {
	rt.handle("HEAD", pattern, handler)
}

func (rt *HttpRouter) ANY(pattern string, handler func(*HttpContext)) {
	rt.mux.HandleFunc(pattern+"/", func(w http.ResponseWriter, r *http.Request) {
		ctx := &HttpContext{w: w, r: r}
		for _, mw := range rt.middleware {
			if !mw(ctx) {
				return
			}
		}
		handler(ctx)
	})
}

func (rt *HttpRouter) Handle(pattern string, handler http.Handler) {
	if strings.Contains(pattern, " ") {
		rt.mux.Handle(pattern, handler)
	} else {
		rt.mux.Handle("GET "+pattern, handler)
	}
}

func (rt *HttpRouter) HandleFunc(pattern string, handler http.HandlerFunc) {
	rt.mux.HandleFunc(pattern, handler)
}

func (rt *HttpRouter) Use(mw func(*HttpContext) bool) {
	rt.middleware = append(rt.middleware, mw)
}

func (rt *HttpRouter) Static(prefix, dir string) {
	rt.mux.Handle(prefix+"/", http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

func (rt *HttpRouter) Group(prefix string) *HttpRouter {
	sub := &HttpRouter{
		mux:        rt.mux,
		middleware: make([]func(*HttpContext) bool, len(rt.middleware)),
	}
	copy(sub.middleware, rt.middleware)
	return sub
}

func (rt *HttpRouter) Serve(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           rt.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return srv.ListenAndServe()
}

func (rt *HttpRouter) ServeTLS(addr, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           rt.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return srv.ListenAndServeTLS(certFile, keyFile)
}

func (rt *HttpRouter) ServeHandler() http.Handler {
	return rt.mux
}

func (HTTPFramework) LoggingMiddleware() func(*HttpContext) bool {
	return func(ctx *HttpContext) bool {
		start := time.Now()
		method := ctx.Method()
		path := ctx.Path()
		fmt.Printf("%s %s %v\n", method, path, time.Since(start))
		return true
	}
}

func (HTTPFramework) CorsMiddleware(allowOrigin, allowMethods, allowHeaders string) func(*HttpContext) bool {
	return func(ctx *HttpContext) bool {
		ctx.SetHeader("Access-Control-Allow-Origin", allowOrigin)
		ctx.SetHeader("Access-Control-Allow-Methods", allowMethods)
		ctx.SetHeader("Access-Control-Allow-Headers", allowHeaders)
		if ctx.Method() == "OPTIONS" {
			ctx.Status(204)
			return false
		}
		return true
	}
}

var serve = HTTPFramework{}
