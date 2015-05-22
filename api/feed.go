package api

import "net/http"
import "strconv"
import "strings"
import "fmt"
import "encoding/json"
import "github.com/go-martini/martini"
import httphelper "github.com/m3ng9i/go-utils/http"
import "github.com/m3ng9i/feedreader"
import "github.com/m3ng9i/qreader/model"
import "github.com/m3ng9i/qreader/utils"


/*
Check if a feed url is already subscribed.

method:     GET
path:       /api/feed/subscription?url={}
example:    /api/feed/subscription?url=http://127.0.0.1&token=xxxx

The output is like: {"request_id":"06442e9fa31f9c8620c68ab9e2df89c4","success":true,"error":{"errcode":0,"errmsg":""},"result":false}
*/
func IsSubscribed() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result

        r.ParseForm()
        url := httphelper.QueryValue(r, "url")

        ok, err := model.IsSubscribed(url)
        if err != nil {
            result.Success = false
            result.Error = ErrQueryDB
            result.Result = false
            result.IntError = err
        } else {
            result.Success = true
            result.Result = ok
        }

        result.Response(w, rid)
    }
}


/*
Subscribe a feed.

method:     POST
path:       /api/feed/subscription?url={}
example:    /api/feed/subscription?url=http://127.0.0.1&token=xxxx
postdata:   nothing

The output is like: {"request_id":"ed4149c76a0910e9b4d3f5921f1cfda3","success":true,"error":{"errcode":0,"errmsg":""},"result":{"id":2,"number":2}}
*/
func Subscribe() martini.Handler {

    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        r.ParseForm()
        url := httphelper.QueryValue(r, "url")

        ok, err := model.IsSubscribed(url)
        if err != nil {
            result.Success = false
            result.Error = ErrQueryDB
            result.Result = url
            result.IntError = err

            result.Response(w)
            return
        }
        if ok {
            result.Success = false
            result.Error = ErrAlreadySubscribed
            result.Result = nil

            result.Response(w)
            return
        }

        feed, items, err := model.FetchFeed(url)
        if err != nil {
            result.Success = false

            if _, ok := err.(*feedreader.ParseError); ok {
                result.Error = ErrParseError
            } else {
                result.Error = ErrFetchError
            }

            result.Result = url
            result.IntError = err
            result.Response(w)
            return
        }

        id, number, name, err := model.Subscribe(feed, items)
        if err != nil {
            result.Success = false
            result.Error = ErrQueryDB
            result.Result = url
            result.IntError = err

            result.Response(w)
            return
        }

        result.Success = true

        var t struct {
            Id      int64   `json:"id"`
            Number  int64   `json:"number"`
            Name    string  `json:"name"`
        }
        t.Id = id
        t.Number = number
        t.Name = name

        result.Result = t

        result.Response(w)
    }
}


/*
Get all feeds' info.

method:     GET
path:       /api/feed/feedlist

The output is like:
{"request_id":"a02d1d1b06cb38bcb4cc248da7d21ba3","success":true,"error":{"errcode":0,"errmsg":""},"result":[{"feed_id":1,"feed_name":"cnBeta.COM业界资讯","feed_feed_url":"http://cnbeta.feedsportal.com/c/34306/f/624776/index.rss","feed_url":"http://www.cnbeta.com","feed_desc":"cnBeta.COM - 简明IT新闻,网友媒体与言论平台","feed_type":"rss","feed_interval":0,"feed_last_fetch":"2015-03-15T19:07:16+08:00","feed_last_failed":"0001-01-01T08:00:00+08:00","feed_last_error":"","feed_max_number":0,"feed_filter":"","feed_use_proxy":false,"feed_note":"","unread":60},{"feed_id":2,"feed_name":"My*Candy","feed_feed_url":"http://mengqi.info/feed.xml","feed_url":"http://mengqi.info","feed_desc":"","feed_type":"atom","feed_interval":0,"feed_last_fetch":"2015-03-15T19:48:13+08:00","feed_last_failed":"0001-01-01T08:00:00+08:00","feed_last_error":"","feed_max_number":0,"feed_filter":"","feed_use_proxy":false,"feed_note":"","unread":0},{"feed_id":3,"feed_name":"Startup News","feed_feed_url":"http://news.dbanotes.net/rss","feed_url":"http://news.dbanotes.net/","feed_desc":"Startup News of China","feed_type":"rss","feed_interval":0,"feed_last_fetch":"2015-03-19T19:36:40+08:00","feed_last_failed":"0001-01-01T08:00:00+08:00","feed_last_error":"","feed_max_number":0,"feed_filter":"","feed_use_proxy":false,"feed_note":"","unread":395},{"feed_id":4,"feed_name":"于江水","feed_feed_url":"http://yujiangshui.com/atom.xml","feed_url":"http://yujiangshui.com/","feed_desc":"一入前端深似海。","feed_type":"atom","feed_interval":0,"feed_last_fetch":"2015-03-19T19:39:09+08:00","feed_last_failed":"0001-01-01T08:00:00+08:00","feed_last_error":"","feed_max_number":0,"feed_filter":"","feed_use_proxy":false,"feed_note":"","unread":20}]}

*/
func FeedList() martini.Handler {
    return func(w http.ResponseWriter, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        feedList, err := model.GetFeedListWithAmount(nil)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        for i, _ := range(feedList) {
            utils.SanitizeSelf(&feedList[i].Name)
            utils.SanitizeSelf(&feedList[i].Desc)
        }

        result.Success = true
        result.Result = feedList
        result.Response(w)
    }
}


/*
Get random items.

method:     GET
path:       /api/article/list/random

example:
    /api/article/list/random
    /api/article/list/random?n=20

You can control the number of results returned by add a "n" parameter, default number is 10.
*/
func RandomArticleList() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        r.ParseForm()
        num, err := strconv.Atoi(httphelper.QueryValue(r, "n"))
        if err != nil || num <= 0 {
            num = 10
        }

        list, err := model.GetRandomArticleList(num)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        for i, _ := range list.Articles {
            utils.SanitizeSelf(&list.Articles[i].Name)
            utils.SanitizeSelf(&list.Articles[i].Author)
            utils.SanitizeSelf(&list.Articles[i].Title)
        }

        result.Success = true
        result.Result = list
        result.Response(w)
    }
}


/*
Get one article.

method:     GET
path:       /api/article/{article id}
example:    /api/article/1
*/
func Article() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        id, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || id <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        article, ok, err := model.GetArticle(id)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if !ok {
            result.Error = ErrNoResultsFound
            result.IntError = fmt.Errorf("Article of id:%d is not found.", id)
            result.Response(w)
            return
        }

        if article.Read == false {
            // read article read
            _, err = model.MarkArticleRead(id, true)
            if err != nil {
                result.Error = ErrQueryDB
                result.IntError = err
                result.Response(w)
                return
            }
        }
        article.Read = true

        utils.SanitizeSelf(&article.Name)
        utils.SanitizeSelf(&article.Author)
        utils.SanitizeSelf(&article.Title)
        utils.SanitizeSelf(&article.Content, true)

        result.Success = true
        result.Result = article
        result.Response(w)
    }
}


/*
Mark article read or unread.
method:     PUT
path:       /api/article/read/{id}
            /api/article/unread/{id}
*/
func MarkReadStatus(markread bool) martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        id, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || id <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        var t struct {
            Id int64 `json:"id"`
            Status string `json:"status"`
        }
        t.Id = id
        if markread {
            t.Status = "read"
        } else {
            t.Status = "unread"
        }

        ok, err := model.MarkArticleRead(id, markread)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if !ok {
            result.Error = ErrNoDataChanged
            result.IntError = fmt.Errorf("No data changed when set article:%d as %s", id, t.Status)
            result.Response(w)
            return
        }

        result.Success = true
        result.Result = t
        result.Response(w)
    }
}


func readJsonPost(r *http.Request, data interface{}) error {
    decoder := json.NewDecoder(r.Body)
    return decoder.Decode(data)
}


/* Mark articles read, by ids, feedid or tag

by ids:
method:     PUT
path:       /api/articles/read
postdata:   {"type":"ids", "value":"1,2,3,4"}

by feedid:
method:     PUT
path:       /api/articles/read
postdata:   {"type":"feedid", "value":"5"}

by tag:
method:     PUT
path:       /api/articles/read
postdata:   {"type":"tag", "value":"blog"}


The output is like:
{"request_id":"5075115ebec5a26119eaac5e226dc6cc","success":true,"error":{"errcode":0,"errmsg":""},"result":{"affected":0}}
*/
func MarkArticlesRead() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        var data struct {
            Type string `json:"type"`
            Value interface{} `json:"value"`
        }

        err := readJsonPost(r, &data)
        if err != nil {
            result.Error = ErrBadRequest
            result.IntError = err
            result.Response(w)
            return
        }

        var affected int64

        // data.Type will be "ids", "feedid", or "tag"
        if data.Type == "ids" {
            respError := func() {
                var r Result
                r.RequestId = rid
                r.Error = ErrBadRequest
                r.IntError = fmt.Errorf("ids is not correct.")
                r.Response(w)
            }

            value, ok := data.Value.([]interface{})
            if !ok {
                respError()
                return
            }

            var ids []int64

            for _, item := range value {
                v, ok := item.(string)
                if !ok {
                    respError()
                    return
                }

                v = strings.TrimSpace(v)
                if v == "" {
                    continue
                }

                id, err := strconv.ParseInt(v, 10, 64)
                if err != nil {
                    respError()
                    return
                }
                ids = append(ids, id)
            }

            affected, err = model.MarkArticlesRead(ids)

        } else if data.Type == "feedid" {

            value, ok := data.Value.(float64)
            if !ok {
                result.Error = ErrBadRequest
                result.IntError = fmt.Errorf("'feedid' is not correct.")
                result.Response(w)
                return
            }

            affected, err = model.MarkArticlesReadByFid(int64(value))

        } else if data.Type == "tag" {

            value, ok := data.Value.(string)
            if !ok {
                result.Error = ErrBadRequest
                result.IntError = fmt.Errorf("'feedid' is not correct.")
                result.Response(w)
                return
            }

            affected, err = model.MarkArticlesReadByTag(value)

        }

        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if affected == 0 {
            result.Error = ErrNoDataChanged
            result.IntError = fmt.Errorf(ErrNoDataChanged.ErrMsg)
            result.Response(w)
            return
        }

        var t struct {
            Affected int64 `json:"affected"`
        }
        t.Affected = affected
        result.Success = true
        result.Result = t
        result.Response(w)
    }
}



/*
Update a feed manually.

method:     GET
path:       /api/feed/update/{article id}
example:    /api/feed/update/1
*/
func Update() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        id, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || id <= 0 {
            result.Success = false
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        affected, err := model.RenewFeed(id)
        if err != nil {
            result.Success = false
            if err == model.ErrFeedNotFound {
                result.Error = ErrNoResultsFound
                result.IntError = fmt.Errorf("Feed of id:%d is not found.", id)
            } else {
                result.IntError = err

                if err == model.ErrFeedHasNoItems {
                    result.Error = ErrFetchError
                } else if _, ok := err.(*feedreader.FetchError); ok {
                    result.Error = ErrFetchError
                } else if _, ok := err.(*feedreader.ParseError); ok {
                    result.Error = ErrParseError
                } else {
                    result.Error = ErrUnexpectedError
                    result.Error.ErrMsg = err.Error()
                }
            }
            result.Response(w)
            return
        }

        result.Success = true
        result.Result = affected
        result.Response(w)
        return
    }
}


/*
Get unread article list.

method:     GET

route:      unread
path:       /api/articles/unread/{limit}/{offset}
example:    /api/articles/unread/10/100

route:      fid
path:       /api/articles/fid/{fid}/{limit}/{offset}
example:    /api/articles/fid/12/10/100

route:      tag
path:       /api/articles/tag/{tag}/{limit}/{offset}
example:    /api/articles/tag/blog/10/100

route:      starred
path:       /api/articles/starred/{limit}/{offset}
example:    /api/articles/starred/10/100
*/
func ArticleList(route string) martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        limit, err := strconv.Atoi(params["limit"])
        if err != nil || limit <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'limit' is not correct.")
            }
            result.Response(w)
            return
        }

        offset, err := strconv.Atoi(params["offset"])
        if err != nil || offset < 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'offset' is not correct.")
            }
            result.Response(w)
            return
        }

        var list model.ArticleList

        if route == "unread" {
            list, err = model.GetArticleList(limit, offset)

        } else if route == "fid" {
            var fid int64
            fid, err = strconv.ParseInt(params["fid"], 10, 64)
            if err != nil || fid < 0 {
                result.Error = ErrBadRequest
                if err != nil {
                    result.IntError = err
                } else {
                    result.IntError = fmt.Errorf("Parameter 'fid' is not correct.")
                }
                result.Response(w)
                return
            }

            list, err = model.GetArticleListByFid(fid, limit, offset)

        } else if route == "tag" {
            list, err = model.GetArticleListByTag(params["tag"], limit, offset)

        } else if route == "starred" {
            list ,err = model.GetStarredArticleList(limit, offset)

        } else {
            result.Error = ErrUnexpectedError
            result.IntError = fmt.Errorf(ErrUnexpectedError.ErrMsg)
            result.Response(w)
            return
        }

        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if len(list.Articles) == 0 {
            result.Error = ErrNoResultsFound
            result.IntError = fmt.Errorf(ErrNoResultsFound.Error())
            result.Response(w)
            return
        }

        for i, _ := range list.Articles {
            utils.SanitizeSelf(&list.Articles[i].Name)
            utils.SanitizeSelf(&list.Articles[i].Author)
            utils.SanitizeSelf(&list.Articles[i].Title)
        }

        result.Success = true
        result.Result = list
        result.Response(w)
    }
}


/* Get feedinfo.

method:     GET
path:       /api/feed/id/{id}
example:    /api/feed/id/1
*/
func FeedInfo() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        id, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || id <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        info, err := model.GetFeedInfo(id)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if info == nil {
            result.Error = ErrNoResultsFound
            result.IntError = fmt.Errorf(ErrNoResultsFound.ErrMsg)
            result.Response(w)
            return
        }

        utils.SanitizeSelf(&info.Name)
        utils.SanitizeSelf(info.Alias)
        utils.SanitizeSelf(&info.Desc)
        utils.SanitizeSelf(info.Note)

        result.Success = true
        result.Result = info
        result.Response(w)
    }
}


/*
Update table Feed and Tag.
Affected columns: Feed.FeedUrl, Feed.Note, Tag.Name, Tag.Fid.

method:     PUT
path:       /api/feed/id/{id}
example:    /api/feed/id/1
postdata:   {"alias":"xxx", "feed_url":"xxx", "feed_note":"xxx", "tags":["t1", "t2"]}
*/
func UpdateFeedAndTags() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        id, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || id <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        var data struct {
            FeedAlias       string      `json:"feed_alias"`
            FeedUrl         string      `json:"feed_url"`
            FeedNote        string      `json:"feed_note"`
            FeedMaxKeep     uint        `json:"feed_max_keep"`
            FeedMaxUnread   uint        `json:"feed_max_unread"`
            FeedInterval    int         `json:"feed_interval"`
            Tags            []string    `json:"tags"`
        }
        err = readJsonPost(r, &data)
        if err != nil {
            result.Error = ErrBadRequest
            result.IntError = err
            result.Response(w)
            return
        }

        var feed model.Feed
        feed.Alias      = &data.FeedAlias
        feed.FeedUrl    = data.FeedUrl
        feed.Note       = &data.FeedNote
        feed.MaxKeep    = &data.FeedMaxKeep
        feed.MaxUnread  = &data.FeedMaxUnread
        feed.Interval   = &data.FeedInterval

        ok, err := model.UpdateFeed(id, &feed)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }
        if !ok {
            result.Error = ErrNoDataChanged
            result.IntError = fmt.Errorf(ErrNoDataChanged.ErrMsg)
            result.Response(w)
            return
        }

        err = model.UpdateTags(id, data.Tags)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        result.Success = true
        result.Response(w)
    }
}


func TagsList() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        r.ParseForm()
        param := httphelper.QueryValue(r, "getall", "none") // "none" means no "getall" parameter provided.

        getall := false
        if param != "none" {
            getall = true
        }

        list, err := model.GetTagsWithFeedList(getall)
        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        result.Success = true
        result.Result = list
        result.Response(w)
    }
}


/*
Mark one or more articles starred or unstarred.

method:     PUT
path:       /api/articles/starred
postdata:   {"status":true/false, "ids": [123", 456]}

response:   {"request_id":"37d5e4acc438bd3fff608f6ecf870a7d","success":true,"error":{"errcode":0,"errmsg":""},"result":{"affected":1}} 
*/
func MarkArticlesStarred() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        var data struct {
            Status bool `json:"status"`
            Ids []float64 `json:"ids"`
        }

        err := readJsonPost(r, &data)
        if err != nil {
            result.Error = ErrBadRequest
            result.IntError = err
            result.Response(w)
            return
        }

        var ids []int64
        var affected int64

        for _, item := range data.Ids {
            ids = append(ids, int64(item))
            affected, err = model.MarkArticlesStarred(ids, data.Status)
        }

        if err != nil {
            result.Error = ErrQueryDB
            result.IntError = err
            result.Response(w)
            return
        }

        if affected == 0 {
            result.Error = ErrNoDataChanged
            result.IntError = fmt.Errorf(ErrNoDataChanged.ErrMsg)
            result.Response(w)
            return
        }

        var t struct {
            Affected int64 `json:"affected"`
        }
        t.Affected = affected

        result.Success = true
        result.Result = t
        result.Response(w)
    }
}


/*
Delete a feed by feed id.

method:     DELETE
path:       /api/feed/id/{id}
example:    /api/feed/id/1
*/
func DeleteFeed() martini.Handler {
    return func(w http.ResponseWriter, params martini.Params, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid

        fid, err := strconv.ParseInt(params["id"], 10, 64)
        if err != nil || fid <= 0 {
            result.Error = ErrBadRequest
            if err != nil {
                result.IntError = err
            } else {
                result.IntError = fmt.Errorf("Parameter 'id' is not correct.")
            }
            result.Response(w)
            return
        }

        err = model.DeleteFeed(fid)
        if err != nil {
            if err == model.ErrFeedCannotBeDeleted {
                result.Error = ErrFeedCannotBeDeleted
            } else {
                result.Error = ErrQueryDB
            }
            result.IntError = err
            result.Response(w)
            return
        }

        result.Success = true
        result.Response(w)
        return
    }
}
