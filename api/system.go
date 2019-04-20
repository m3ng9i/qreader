package api

import "os"
import "net/http"
import "github.com/go-martini/martini"
import httphelper "github.com/m3ng9i/go-utils/http"
import "github.com/m3ng9i/qreader/global"
import "github.com/m3ng9i/qreader/model"


func Settings() martini.Handler {
    return func(w http.ResponseWriter, rid httphelper.RequestId) {
        var result Result

        var data = make(map[string]interface{})
        data["ConfigFile"]  = global.ConfigFile
        data["PathRoot"]    = global.PathRoot
        data["PathDB"]      = global.PathDB
        data["IPs"]         = global.IPs
        data["Port"]        = global.Port
        data["Usetls"]      = global.Usetls
        data["UseProxy"]    = global.UseProxy
        data["Debug"]       = global.Debug

        if global.ProxyConfig != nil {
            data["ProxyAddr"] = global.ProxyConfig.Addr
        }

        var d = make(map[string]interface{})
        d["SystemInfo"] = data

        feedNumber, err := model.GetFeedNumber()
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        articleNumber, err := model.GetArticleNumber()
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        dbsize, err := model.DBSize()
        if err != nil {
            result.Error = ErrSystemError
            result.IntError = err
            result.Response(w)
            return
        }

        var t struct {
            Feed int64 `json:"feed"`
            *model.ArticleNumber
            DBSize int64 `json:"dbsize"`
        }
        t.Feed = feedNumber
        t.ArticleNumber = articleNumber
        t.DBSize = dbsize

        d["Summary"] = t
        d["Version"] = global.Version

        result.Success = true
        result.Result = d
        result.Response(w, rid)
    }
}


func CloseServer() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        result.Success = true
        result.Result = "QReader server is going to shutdown"
        result.Response(w)

        global.Logger.Warnf("[API] [#%s] The server is shutdown manually.", result.RequestId)
        global.Logger.Wait()
        os.Exit(0)
    }
}


