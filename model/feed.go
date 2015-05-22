package model

import "fmt"
import "strings"
import dbsql "database/sql"
import "github.com/go-xorm/xorm"
import "../global"


// Get number of subscribed feed.
func GetFeedNumber() (int64, error) {
    return global.Orm.Count(&Feed{})
}


type ArticleNumber struct {
    Feed        int64   `json:"feed"`
    Amounts     int64   `json:"amounts"`
    Read        int64   `json:"read"`
    Unread      int64   `json:"unread"`
    Starred     int64   `json:"starred"`
    Grabbed     int64   `json:"grabbed"`
}


/*
Get number of articles which grabbed from the feed.

Return value:

    number.Amounts  Current number of articles saved in the database.
    number.Read     Number of articles that is marked to read in the database.
    number.Unread   Equals to result[0] - resut[1], means unread number of articles.
    number.Starred  Number of starred articles.
    number.Grabbed  Number of articles the QReader has grabbed. This is larger than result[0] because old articles will be deleted.
*/
func GetArticleNumber() (number *ArticleNumber, err error){

    sql := `select  (select count(*) from Item) amounts,
                    (select count(*) from Item where Read=1) read,
                    (select count(*) from Item where Read=0) unread,
                    (select count(*) from Item where Starred=1) starred,
                    (select Id from Item order by Id desc limit 1) grabbed`

    number = new(ArticleNumber)
    var grabbed dbsql.NullInt64

    err = global.Orm.DB().QueryRow(sql).Scan(
                &number.Amounts,
                &number.Read,
                &number.Unread,
                &number.Starred,
                &grabbed)

    number.Grabbed = grabbed.Int64
    return
}


// Get all data in table Feed.
func GetFeedList() (feedlist []*Feed, err error) {

    feed := new(Feed)
    rows, err := global.Orm.Rows(feed)
    if err != nil {
        return
    }
    defer rows.Close()

    for rows.Next() {
        err = rows.Scan(feed)
        if err != nil {
            return
        }
        feedlist = append(feedlist, feed)
    }
    return
}


type FeedWithAmount struct {
    Feed `xorm:"extends"`
    Amount uint64 `json:"amount"`
    Unread uint64 `json:"unread"`
    Starred uint64 `json:"starred"`
}


// Get feed list with amount, unread and starred number.
func GetFeedListWithAmount(session *xorm.Session) (feedlist []*FeedWithAmount, err error) {

    sql := `select Feed.*,t1.Amount,t2.Unread,t3.Starred from Feed left join
            (select Fid,count(*) as Amount from Item group by Fid) as t1 on Feed.Id = t1.Fid left join
            (select Fid,count(*) as Unread from Item where Read=0 group by Fid) as t2 on t1.Fid=t2.Fid left join
            (select Fid,count(*) as Starred from Item where Starred=1 group by Fid) as t3 on t1.Fid=t3.Fid`

    if session == nil {
        err = global.Orm.Sql(sql).Find(&feedlist)
    } else {
        err = session.Sql(sql).Find(&feedlist)
    }

    return
}


type TagsWithFeedList map[string][]*FeedWithAmount


// Get tags with feed list. If getall is true, the function will get all data even if feed's unread item number is 0.
// If getall is false, the feeds which unread item number is 0 will not returned.
func GetTagsWithFeedList(getall ...bool) (feedlist TagsWithFeedList, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    tags, err := getTagWithFeedIds(session)
    if err != nil {
        session.Commit()
        return
    }

    list, err := GetFeedListWithAmount(session)
    if err != nil {
        session.Commit()
        return
    }

    all := false
    if len(getall) > 0 && getall[0] {
        all = true
    }

    feedlist = make(TagsWithFeedList)
    for k, v := range tags {
        for _, id := range v {
            for _, item := range list {
                if item.Id == id {
                    if all == true || (all == false && item.Unread > 0) {
                        feedlist[k] = append(feedlist[k], item)
                    } else {
                        break
                    }
                }
            }
        }
    }

    session.Commit()
    return
}


type TagAndFeedId struct {
    FeedId int64
    Tag string
}


func getTagAndFeedIds(session *xorm.Session) (result []*TagAndFeedId, err error) {
    err = session.Sql("select Feed.Id as FeedId,Tag.Name as Tag from Feed left join Tag on Feed.Id=Tag.Fid").Find(&result)
    return
}


type TagWithFeedIds map[string][]int64


/* Get tag with feed ids

example: map[IT:[1,2,3], blog:[2,3,4]]

if there tags: tag, Tag and TAG, they will be combined to TAG (uppercase form)
*/
func getTagWithFeedIds(session *xorm.Session) (result TagWithFeedIds, err error) {

    t, err := getTagAndFeedIds(session)
    if err != nil {
        return
    }

    r := make(TagWithFeedIds)
    result = make(TagWithFeedIds)
    tags := make(map[string]int)

    for _, i := range t {
        r[i.Tag] = append(r[i.Tag], i.FeedId)
    }

    for k, _ := range r {
        upperTag := strings.ToUpper(k)
        v, _ := tags[upperTag]
        tags[upperTag] = v + 1
    }

    for k, v := range r {
        upperTag := strings.ToUpper(k)
        num, _ := tags[upperTag]
        if num > 1 {
            _, ok := result[upperTag]
            if ok {
                result[upperTag] = append(result[upperTag], v...)
            } else {
                result[upperTag] = v
            }
        } else {
            result[k] = v
        }
    }

    return
}


type Article struct {
    Item `xorm:"extends"`
    Feed `xorm:"extends"`
}


type ArticleList struct {
    Articles []*Article `xorm:"extends"`
    Number int64
}


// Get unread random items.
func GetRandomArticleList(limit int) (list ArticleList, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    list.Number, err = session.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Read=0").Count(&Article{})
    if err != nil {
        session.Commit()
        return
    }

    // BUG: Omit("Item.Content") doesn't work, still select all columns, waiting xorm team to fix it.
    // opened an issue at: https://github.com/go-xorm/xorm/issues/222
    err = session.Omit("Item.Content").Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Read=0").OrderBy("RANDOM()").Limit(limit).Find(&list.Articles)

    session.Commit()
    return
}


// Get unread article list, order by id desc.
func GetArticleList(limit, offset int) (list ArticleList, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    list.Number, err = session.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Read=0").Count(&Article{})
    if err != nil {
        session.Commit()
        return
    }

    // BUG: Omit("Item.Content") doesn't work, still select all columns, waiting xorm team to fix it.
    // opened an issue at: https://github.com/go-xorm/xorm/issues/222
    err = session.Omit("Item.Content").Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Read=0").Desc("Id").Limit(limit, offset).Find(&list.Articles)

    session.Commit()
    return
}


// Get starred article lsit.
func GetStarredArticleList(limit, offset int) (list ArticleList, err error) {
    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    list.Number, err = session.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Starred=1").Count(&Article{})
    if err != nil {
        session.Commit()
        return
    }

    // BUG: Omit("Item.Content") doesn't work, still select all columns, waiting xorm team to fix it.
    // opened an issue at: https://github.com/go-xorm/xorm/issues/222
    err = session.Omit("Item.Content").Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Starred=1").Desc("Id").Limit(limit, offset).Find(&list.Articles)

    session.Commit()
    return
}


// Get unread article list by fid, order by id desc.
func GetArticleListByFid(fid int64, limit, offset int) (list ArticleList, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    list.Number, err = session.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").
                            And(fmt.Sprintf("Item.Fid=%d", fid)).And("Item.Read=0").Count(&Article{})
    if err != nil {
        session.Commit()
        return
    }

    // BUG: Omit("Item.Content") doesn't work, still select all columns, waiting xorm team to fix it.
    // opened an issue at: https://github.com/go-xorm/xorm/issues/222
    err = session.Omit("Item.Content").Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").
            And(fmt.Sprintf("Item.Fid=%d", fid)).And("Item.Read=0").Desc("Id").Limit(limit, offset).Find(&list.Articles)

    session.Commit()
    return
}


// Get unread article list by tag name, order by id desc.
func GetArticleListByTag(tag string, limit, offset int) (list ArticleList, err error) {
    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    fids, err := getFeedIdsByTag(session, tag)
    if err != nil {
        session.Commit()
        return
    }

    if len(fids) == 0 {
        session.Commit()
        return
    }

    list.Number, err = session.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").
                            In("Item.Fid", fids).And("Item.Read=0").Count(&Article{})
    if err != nil {
        session.Commit()
        return
    }

    // BUG: Omit("Item.Content") doesn't work, still select all columns, waiting xorm team to fix it.
    // opened an issue at: https://github.com/go-xorm/xorm/issues/222
    err = session.Omit("Item.Content").Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").
            In("Item.Fid", fids).And("Item.Read=0").Desc("Id").Limit(limit, offset).Find(&list.Articles)

    //session.Table
    session.Commit()
    return
}


// Get feed ids by tag name.
func getFeedIdsByTag(session *xorm.Session, tag string) (fids []int64, err error) {
    var t []*Tag
    err = session.Cols("Fid").Where("Name = ?", tag).Find(&t)
    if err != nil {
        return
    }
    for _, i := range t {
        fids = append(fids, i.Fid)
    }
    return
}


/*
Get one article by Item.Id.

If no article get, ok is false.
*/
func GetArticle(id int64) (article *Article, ok bool, err error) {
    var a Article
    ok, err = global.Orm.Table("Item").Join("INNER", "Feed", "Item.Fid=Feed.Id").And("Item.Id=?", id).Get(&a)
    article = &a
    return
}


/*
Mark article read or unread.

    markread = true: mark read
    markread = false: mark unread
*/
func MarkArticleRead(id int64, markread bool) (ok bool, err error) {
    var read int
    if markread {
        read = 1
    } else {
        read = 0
    }

    result, err := global.Orm.Exec("update Item set read=? where id=?", read, id)
    if err != nil {
        return
    }

    affected, err := result.RowsAffected()
    if err != nil {
        return
    }

    if affected > 0 {
        ok = true
    }

    return
}


// Mark articles to read by article ids.
func MarkArticlesRead(ids []int64) (affected int64, err error) {
    affected, err = global.Orm.In("id", ids).UseBool("Read").Update(&Item{Read:true})
    return
}


// Get feed information by id.
func GetFeed(id int64) (feed *Feed, ok bool, err error) {
    var f Feed
    ok, err = global.Orm.Id(id).Get(&f)
    feed = &f
    return
}


type FeedInfo struct {
    Feed    `xorm:"extends"`
    Tags    []string
    Amounts uint64
    Read    uint64
    Unread  uint64
    Starred uint64
}


// Get information of one feed by Feed.Id.
// This function executes 3 queries, and combine the results together.
// If no record found ,feedinfo will be nil.
func GetFeedInfo(fid int64) (feedinfo *FeedInfo, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    var feed FeedInfo

    ok, err := session.Id(fid).Get(&feed.Feed)
    if err != nil || !ok {
        session.Commit()
        return
    }

    var tags []*Tag
    err = session.Where("Fid = ?", fid).Find(&tags)
    if err != nil {
        session.Commit()
        return
    }
    for _, i := range tags {
        feed.Tags = append(feed.Tags, i.Name)
    }

    sql := `select  (select count(*) from Item where Fid = %d) amounts,
                    (select count(*) from Item where Fid = %d and Read = 1) read,
                    (select count(*) from Item where Fid = %d and Read = 0) unread,
                    (select count(*) from Item where Fid = %d and Starred = 1) starred`
    sql = fmt.Sprintf(sql, fid, fid, fid, fid)
    err = session.DB().QueryRow(sql).Scan(&feed.Amounts,
                                          &feed.Read,
                                          &feed.Unread,
                                          &feed.Starred)
    if err != nil {
        session.Commit()
        return
    }

    session.Commit()

    feedinfo = &feed

    return
}


// Modify table Feed.
func UpdateFeed(fid int64, feed *Feed) (ok bool, err error) {
    affected, err := global.Orm.Id(fid).Update(feed)
    if affected > 0 {
        ok = true
    }
    return
}


/*
Remove same name tags. Remove each tag's leading and trailing white space.

example:
input []string{"aa ", " bb", "cc", "AA", "cC"}
output []string{"aa", "bb", "cc"}
*/
func trimTags(tags []string) []string {
    var t []string

    LOOP:
    for _, i := range tags {
        tag := strings.TrimSpace(i)
        if tag == "" {
            continue
        }
        t1 := strings.ToLower(tag)
        for _, j := range t {
            t2 := strings.ToLower(j)
            if t1 == t2 {
                continue LOOP
            }
        }
        t = append(t, tag)
    }

    return t
}


// Update table Tag.
func UpdateTags(fid int64, tags []string) (err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    total, err := session.Where("Id = ?", fid).Count(&Feed{})
    if err != nil {
        session.Commit()
        return
    }

    _, err = session.Where("Fid = ?", fid).Delete(&Tag{})
    if err != nil {
        session.Rollback()
        return
    }

    // If a Feed is not exists (determin by Feed.Id), try to delete data in Tag Table, then return.
    if total == 0 {
        session.Commit()
        err = ErrFeedNotFound
        return
    }

    tags = trimTags(tags)

    for _, tag := range tags {
        _, err = session.Insert(&Tag{Name: tag, Fid: fid})
        if err != nil {
            session.Rollback()
            return
        }
    }

    session.Commit()
    return
}


// Mark articles read by fid.
func MarkArticlesReadByFid(fid int64) (affected int64, err error) {
    affected, err = global.Orm.Table("Item").Where("Fid = ?", fid).UseBool("Read").Update(&Item{Read: true})
    return
}


// Mark articles read by tag.
func MarkArticlesReadByTag(tag string) (affected int64, err error) {
    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    feedIds, err := getFeedIdsByTag(session, tag)
    if err != nil {
        session.Commit()
    }

    affected, err = session.Table("Item").In("Fid", feedIds).UseBool("Read").Update(&Item{Read: true})

    session.Commit()
    return
}


// Mark articles starred.
func MarkArticlesStarred(ids []int64, status bool) (affected int64, err error) {

    if len(ids) == 0 {
        return
    }

    affected, err = global.Orm.Table("Item").In("Id", ids).UseBool("Starred").Update(&Item{Starred: status})
    return
}


// Delete a feed, including associated articles and tags.
func DeleteFeed(fid int64) (err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    number, err := session.Table("Item").Where("Fid = ?", fid).And("Starred=1").Count(&Item{})
    if err != nil {
        session.Commit()
        return
    }

    if number > 0 {
        err = ErrFeedCannotBeDeleted
        session.Commit()
        return
    }

    _, err = session.Where("Fid = ?", fid).Delete(&Item{})
    if err != nil {
        session.Rollback()
        return
    }

    _, err = session.Where("Fid = ?", fid).Delete(&Tag{})
    if err != nil {
        session.Rollback()
        return
    }

    _, err = session.Where("Id = ?", fid).Delete(&Feed{})
    if err != nil {
        session.Rollback()
        return
    }

    session.Commit()
    return
}
