package server

import "fmt"
import "time"
import "net/http"
import "strings"
import "sync"
import "github.com/go-martini/martini"
import httphelper "github.com/m3ng9i/go-utils/http"
import "github.com/m3ng9i/qreader/api"
import "github.com/m3ng9i/qreader/global"
import "github.com/m3ng9i/qreader/utils"


// Generate a request id for each request.
func requestId() martini.Handler {
    // requestIdFunc return a httphelper.RequestId type of value.
    requestIdFunc := httphelper.RequestIdGenerator(32)

    return func(c martini.Context) {
        c.Map(requestIdFunc())
        c.Next()
    }
}


func recovery() martini.Handler {
    return func(w http.ResponseWriter, ctx martini.Context) {
        defer func() {
            if err := recover(); err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                global.Logger.Errorf("PANIC: %s", err)
            }
        }()

        ctx.Next()
    }
}


// Create the router.
func createRouter() martini.Router {
    router := martini.NewRouter()

    router.Get(     "/api/feed/list",                               api.FeedList())
    router.Get(     "/api/feed/subscription",                       api.IsSubscribed())
    router.Post(    "/api/feed/subscription",                       api.Subscribe())
    router.Post(    "/api/feed/id/:id",                             api.Update())
    router.Get(     "/api/feed/id/:id",                             api.FeedInfo())
    router.Put(     "/api/feed/id/:id",                             api.UpdateFeedAndTags())
    router.Delete(  "/api/feed/id/:id",                             api.DeleteFeed())
    router.Get(     "/api/articles/random",                         api.RandomArticleList())
    router.Get(     "/api/articles/unread/:limit/:offset",          api.ArticleList("unread"))
    router.Get(     "/api/articles/fid/:fid/:limit/:offset",        api.ArticleList("fid"))
    router.Get(     "/api/articles/tag/:tag/:limit/:offset",        api.ArticleList("tag"))
    router.Get(     "/api/articles/starred/:limit/:offset",         api.ArticleList("starred"))
    router.Get(     "/api/articles/search/:deflimit",               api.SearchList())
    router.Put(     "/api/articles/read",                           api.MarkArticlesRead())
    router.Put(     "/api/articles/starred",                        api.MarkArticlesStarred())
    router.Get(     "/api/article/content/:id",                     api.Article())
    router.Put(     "/api/article/read/:id",                        api.MarkReadStatus(true))   // mark read
    router.Put(     "/api/article/unread/:id",                      api.MarkReadStatus(false))  // mark unread
    router.Get(     "/api/tags/list",                               api.TagsList())
    router.Get(     "/api/system/settings",                         api.Settings())
    router.Put(     "/api/system/shutdown",                         api.CloseServer())
    router.Get(     "/api/",                                        api.Status())               // do not need api token
    router.Get(     "/api/checktoken",                              api.Status())               // check api token
    router.Any(     "/api/**",                                      api.Default())

    return router
}


// Create the mux.
func createMux(r martini.Router) *martini.Martini {
    mux := martini.New()
    mux.Use(recovery())
    mux.Use(requestId())    // generate a request id for each request
    mux.Use(httpLog())      // log every request
    mux.Use(checkToken())   // check if token is valid if a request path is begin with /api/

    // set global.PathClient to static file path, and if a path of an url start with /, that will be pointed to the static file path.
    mux.Use(martini.Static(global.PathClient, martini.StaticOptions{Prefix:"/", SkipLogging:true, IndexFile:"index.html"}))

    mux.Action(r.Handle)

    return mux
}


func httpLog() martini.Handler {

    return func(w http.ResponseWriter, r *http.Request, ctx martini.Context, rid httphelper.RequestId) {

        timer := time.Now() // start time of request

        ctx.Next() // execute the other handler

        rw := w.(martini.ResponseWriter)

        loginfo := fmt.Sprintf("[Access] [#%s] [status:%v] [ip:%s] [host:%s] [method:%s] [path:%s] [user-agent:%s] [ref:%s] [time:%.3fms]",
                        string(rid),                        // request id
                        rw.Status(),                        // http status code
                        httphelper.GetIP(r),                // client IP
                        r.Host,
                        r.Method,
                        r.URL,
                        r.Header["User-Agent"][0],
                        r.Referer(),
                        time.Since(timer).Seconds()*1000)   // request time (milliseconds)

        global.Logger.Info(loginfo)
    }
}


// Get token in query string and check if it's valid, if not, response error.
func checkToken() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, ctx martini.Context, rid httphelper.RequestId) {

        var apiPrefix = "/api/"

        if strings.HasPrefix(r.URL.Path, apiPrefix) && len(r.URL.Path) > len(apiPrefix) {
            token := r.Header.Get("X-QReader-Token")
            if !utils.ValidateToken(token) {
                var result api.Result
                result.RequestId = rid
                result.Success = false
                result.Error = api.ErrTokenInvalid
                result.IntError = fmt.Errorf("invalid token")
                result.Result = nil
                result.Response(w)
                return
            }
        }

        ctx.Next()
    }
}


var Router martini.Router

var Mux *martini.Martini

var once sync.Once

// server's Init() should run after global's Init().
func Init() {
    once.Do(func() {
        Router = createRouter()
        Mux = createMux(Router)
    })
}
