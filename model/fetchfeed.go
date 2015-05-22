package model

import "crypto/md5"
import "fmt"
import "time"
import "strings"
import "sync"
import "github.com/go-xorm/xorm"
import "github.com/m3ng9i/feedreader"
import h "github.com/m3ng9i/go-utils/http"
import "github.com/m3ng9i/qreader/global"


// Check if a feed url is already subscribed.
func IsSubscribed(url string) (s bool, e error) {
    total, e := global.Orm.Where("Feedurl = ?", url).Count(&Feed{})
    if e != nil {
        return
    }

    if total > 0 {
        s = true
    }

    return
}


/*
Subscribe a feed. If a feed is already subscribed, a "UNIQUE constraint failed: Feed.Url" error will be return.

return value:
    id      id in table feed
    num     amount of added items
    name    feed name
*/
func Subscribe(feed *Feed, items []*Item) (id int64, num int64, name string, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    // Feed.Filter and Feed.Note cannot be null.
    empty := ""
    feed.Alias = &empty
    feed.Filter = &empty
    feed.Note = &empty

    var zeroInt int = 0
    var zeroUint uint = 0
    feed.Interval = &zeroInt
    feed.MaxUnread = &zeroUint
    feed.MaxKeep = &zeroUint

    name = feed.Name

    // insert data to table Feed
    _, err = session.Insert(feed)
    if err != nil {
        session.Rollback()
        return
    }

    // get Feed.Id
    f := new(Feed)
    _, err = session.Cols("Id").Where("Feedurl = ?", feed.FeedUrl).Get(f)
    if err != nil {
        session.Rollback()
        return
    }
    id = f.Id

    var affected int64
    // insert data to table Item
    for _, item := range items {
        item.Fid = f.Id
        // ignore error like "UNIQUE constraint failed"
        affected, _ = session.Insert(item)
        num += affected
    }

    err = session.Commit()
    if err != nil {
        session.Rollback()
        return
    }

    return
}


func fetchFeed(url string, fetcher *h.Fetcher) (feed *Feed, items []*Item, err error) {
    fd, err := feedreader.Fetch(url, fetcher)
    if err != nil {
        return
    }

    feed, items = assembleFeed(fd)
    return
}


// Fetch a feed normally or behind a proxy.
func FetchFeed(url string) (feed *Feed, items []*Item, err error) {

    msgNormally := fmt.Sprintf("[FETCH] Fetch feed '%s' normally", url)
    msgProxy := fmt.Sprintf("[FETCH] Fetch feed '%s' behind proxy", url)

    if global.UseProxy == global.PROXY_ALWAYS {
        feed, items, err = fetchFeed(url, global.Socks5Fetcher)
        if err != nil {
            global.Logger.Errorf("%s: %s", msgProxy, err.Error())
        } else {
            global.Logger.Infof(msgProxy)
        }
        return
    }

    feed, items, err = fetchFeed(url, global.NormalFetcher)
    if err == nil {
        global.Logger.Infof(msgNormally)
        return
    }
    if err != nil {
        global.Logger.Errorf("%s: %s", msgNormally, err.Error())

        if global.UseProxy == global.PROXY_TRY {
            feed, items, err = fetchFeed(url, global.Socks5Fetcher)
            if err != nil {
                global.Logger.Errorf("%s: %s", msgProxy, err.Error())
            } else {
                global.Logger.Infof(msgProxy)
            }
        }
    }

    return
}

func assembleFeed(fd *feedreader.Feed) (feed *Feed, items []*Item)  {

    now := time.Now()

    feed            = new(Feed)
    feed.Name       = fd.Title
    feed.FeedUrl    = fd.FeedLink
    feed.Url        = fd.Link
    feed.Desc       = fd.Description
    feed.Type       = fd.Type
    feed.LastFetch  = now

    for _, i := range fd.Items {
        var item = new(Item)

        if i.Author != nil {
            item.Author = i.Author.Name
        }

        item.Url        = i.Link
        item.Guid       = i.Guid
        item.Title      = i.Title
        item.Content    = i.Content
        item.FetchTime  = now

        item.PubTime = i.PubDate
        if item.PubTime.IsZero() {
            item.PubTime = i.Updated
        }

        h := md5.New()
        fmt.Fprint(h, item.Content)
        item.Hash = fmt.Sprintf("%x", h.Sum(nil))

        items = append(items, item)
    }

    return
}


type FeedRenewInfo struct {
    Id          int64
    Feed        *Feed
    Items       []*Item
    FetchTime   time.Time
    FetchError  error
}


func fetchFeedAndItems(id int64) (info FeedRenewInfo, err error) {

    feed, ok, err := GetFeed(id)
    if err != nil {
        return
    }
    if !ok {
        err = ErrFeedNotFound
        return
    }

    info.Id = id
    info.Feed, info.Items, info.FetchError = FetchFeed(feed.FeedUrl)
    info.FetchTime = time.Now()

    if info.FetchError != nil {
        f := new(Feed)
        f.LastFailed = info.FetchTime
        f.LastError = info.FetchError.Error()
        _, e := global.Orm.Id(info.Id).Update(f)
        if e != nil {
            err = e
        } else {
            err = info.FetchError
        }
        return
    }

    // If feed has no items, record as an error.
    if len(info.Items) == 0 {
        info.FetchError = ErrFeedHasNoItems
        err = info.FetchError
    }

    return
}


func renewFeed(info FeedRenewInfo) (affected int64, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    _, err = session.Id(info.Id).Update(info.Feed)
    if err != nil {
        session.Rollback()
        return
    }

    for _, item := range info.Items {
        item.Fid = info.Id
        num, e := session.Insert(item)
        if e != nil {
            // Table Item has some unique indexes for preventing insert duplicate data.
            // So these errors should be ignored.
            if strings.HasPrefix(e.Error(), "UNIQUE constraint failed") {
                global.Logger.Noticef("[MODEL] insert item to table Item failed: %s, fid: %d, title: %s, url:%s, guid: %s",
                    e.Error(), info.Id, item.Title, item.Url, item.Guid)
                continue
            } else {
                session.Rollback()
                err = e
                return
            }
        }
        affected += num
    }

    err = session.Commit()
    return
}


// Renew a feed: fetch new items of feed, and insert them into Item table.
// If some information of remote feed has changed, e.g. feed name, description, they'll be synced to Feed table.
// If returned error is not nil, it will be feedreader.FetchError, feedreader.ParseError or common error.
func RenewFeed(id int64) (affected int64, err error) {
    feedInfo, err := fetchFeedAndItems(id)
    if err != nil {
        return
    }
    affected, err = renewFeed(feedInfo)
    return
}


// Get feed.Ids that need to update
func GetFidsNeedToUpdate(interval uint) (fids []int64, err error) {

    if interval == 0 {
        err = fmt.Errorf("interval cannot be zero.")
        return
    }

    var feeds []*Feed

    err = global.Orm.Cols("Id", "Interval", "LastFetch", "LastFailed").Asc("LastFetch").Find(&feeds)
    if err != nil {
        return
    }

    now := time.Now()
    var t time.Time

    for _, feed := range feeds {

        if feed == nil {
            continue
        }

        if *feed.Interval < 0 {
            continue
        }

        if *feed.Interval == 0 {
            t = feed.LastFetch.Add(time.Duration(interval) * time.Minute)
        } else {
            t = feed.LastFetch.Add(time.Duration(*feed.Interval) * time.Minute)
        }

        if t.Before(now) {
            // if fetch failed, try again 1 hour later.
            t = feed.LastFailed.Add(time.Hour)
            if t.Before(now) {
                fids = append(fids, feed.Id)
            }
        }
    }

    return
}


// Mark old articles read.
func MarkOldArticlesRead(session *xorm.Session) (affected int64, err error) {

    list, err := GetFeedListWithAmount(session)
    if err != nil {
        return
    }

    for _, item := range list {
        if int(*item.MaxUnread) > 0 && item.Unread > uint64(*item.MaxUnread) {
            n := int(item.Unread) - int(*item.MaxUnread)

            sql := "update Item set read=1 where id in (select id from Item where Fid = ? and Read = 0 order by Id asc limit ?)"
            result, e := session.Exec(sql, item.Id, n)
            if e != nil {
                err = e
                return
            }
            num, e := result.RowsAffected()
            if e != nil {
                err = e
                return
            }
            global.Logger.Infof("[TRIM DATA] in transaction: mark old articles read: feed id: %d, affected: %d", item.Id, num)
            affected += num
        }
    }

    return
}


// Delete old articles which read=1.
func DeleteOldArticles(session *xorm.Session) (affected int64, err error) {

    list, err := GetFeedListWithAmount(session)
    if err != nil {
        return
    }

    for _, item := range list {
        if int(*item.MaxKeep) > 0 && (item.Amount - item.Unread > uint64(*item.MaxKeep)) {
            n := int(item.Amount - item.Unread - uint64(*item.MaxKeep))

            sql := "delete from Item where id in (select id from Item where Fid = ? and Read = 1 and starred = 0 order by Id asc limit ?)"
            result, e := session.Exec(sql, item.Id, n)
            if e != nil {
                err = e
                return
            }
            num, e := result.RowsAffected()
            if e != nil {
                err = e
                return
            }
            global.Logger.Infof("[TRIM DATA] in transaction: delete old articles which is read: feed id: %d, affected: %d", item.Id, num)
            affected += num
        }
    }

    return
}


// Mark old articles read, then delete old articles which read=1.
func TrimData() (markread, deleted int64, err error) {
    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    markread, err = MarkOldArticlesRead(session)
    if err != nil {
        session.Rollback()
        global.Logger.Errorf("[TRIM DATA] rollback: mark old articles read: %s", err.Error())
        return
    }

    deleted, err = DeleteOldArticles(session)
    if err != nil {
        session.Rollback()
        global.Logger.Errorf("[TRIM DATA] rollback: mark delete old articles: %s", err.Error())
        return
    }

    session.Commit()
    global.Logger.Infof("[TRIM DATA] commit: mark read: %d, delete: %d", markread, deleted)
    return
}


func AutoUpdateFeed(interval uint) {

    renewInfo := make(chan FeedRenewInfo, 30)

    // a goroutine for renewing feed
    go func(renew chan FeedRenewInfo) {
        for feed := range renew {
            affected, err := renewFeed(feed)
            if err != nil {
                global.Logger.Errorf("[SYSTEM] Auto update failed: fid:%d, %s", feed.Id, err.Error())
            } else {
                global.Logger.Infof("[SYSTEM] Auto update success: fid:%d, add %d articles.", feed.Id, affected)
            }
        }
    }(renewInfo)

    // a goroutine for fetching feed
    go func(renew chan FeedRenewInfo) {
        for {
            fids, err := GetFidsNeedToUpdate(interval)
            if err != nil {
                global.Logger.Errorf("[SYSTEM] Auto update failed: %s", err.Error())
                goto NEXT
            }

            if len(fids) > 0 {
                var wg sync.WaitGroup

                // fetch 5 feeds at one time at most
                maxFetch := make(chan bool, 5)

                for _, fid := range fids {
                    wg.Add(1)
                    go func(feedid int64) {
                        maxFetch <- true

                        feedInfo, err := fetchFeedAndItems(feedid)
                        if err != nil {
                            global.Logger.Errorf("[SYSTEM] Auto update failed: fid:%d, %s", feedid, err.Error())
                        } else {
                            renew <- feedInfo
                        }

                        <- maxFetch
                        wg.Done()
                    }(fid)
                }
                wg.Wait()
                global.Logger.Info("[SYSTEM] Auto update: finish fetching.")
            } else {
                global.Logger.Info("[SYSTEM] Auto update: no feed need to update.")
            }

            // after fetching, wait a minute for database updating, then trim data.
            <- time.After(time.Minute)
            TrimData()

            // try again after few minutes
            NEXT:
            <- time.After(10 * time.Minute)
        }
    }(renewInfo)
}
