package model

import "fmt"
import "strings"
import "strconv"
import qp "github.com/m3ng9i/go-utils/query-parser"
import "github.com/m3ng9i/go-utils/slice"
import "github.com/m3ng9i/qreader/global"

type SearchQuery struct {
    Fid     *[]int64
    Title   *[]string
    Content *[]string
    Read    *bool // nil: unread
    Orderby *[]string
    Asc     *bool
    Tag     *[]string
    Starred *bool // nil: any
    Num     *int // nil: default value
}


type SearchQueryError struct {
    qp.Node
    Msg string
}


func (e *SearchQueryError) Error() string {
    return e.Msg
}


// Determine if a string is a legal column used in sql orderby statement
func legalOrderbyColumn(c string) bool {
    c = strings.ToLower(c)

    columns := []string {
        "id",
        "fid",
        "author",
        "url",
        "guid",
        "title",
        "content",
        "pubtime",
        "fetchtime",
        "starred",
        "read",
    }

    for _, col := range columns {
        if c == col {
            return true
        }
    }

    return false
}

// Get SearchQuery structure base on a search query.
// err may be qp.InvalidCharError or SearchQueryError.
// E.g. sq, err := Search("fid:22 title:'article title' orderby:title order:asc")
func Search(q string) (sq SearchQuery, err error) {

    nodes, err := qp.Parse(q)
    if err != nil {
        return
    }

    t := true
    f := false

    readAny := false

    for _, node := range *nodes {

        if node.Negative {
            err = &SearchQueryError {
                Node: node,
                Msg: fmt.Sprintf("Incorrect use of '-' in key: %s", node.Key),
            }
            return
        }

        switch strings.ToLower(node.Key) {
            case "fid":
                var fids []int64
                for _, value := range node.Values {
                    fid, e := strconv.ParseInt(value, 10, 64)
                    if e != nil || fid <= 0 {
                        err = &SearchQueryError {
                            Node: node,
                            Msg: fmt.Sprintf("Incorrect fid value: %s", value),
                        }
                        return
                    }
                    fids = append(fids, fid)
                }
                if len(fids) > 0 {
                    if sq.Fid != nil {
                        *sq.Fid = append(*sq.Fid, fids...)
                    } else {
                        sq.Fid = &fids
                    }
                }

            case "": fallthrough
            case "keyword":
                v := node.Values

                if sq.Title != nil {
                    *sq.Title = append(*sq.Title, node.Values...)
                } else {
                    sq.Title = &v
                }

                if sq.Content != nil {
                    *sq.Content = append(*sq.Content, node.Values...)
                } else {
                    sq.Content = &v
                }

            case "title":
                if sq.Title != nil {
                    *sq.Title = append(*sq.Title, node.Values...)
                } else {
                    v := node.Values
                    sq.Title = &v
                }

            case "content":
                if sq.Content != nil {
                    *sq.Content = append(*sq.Content, node.Values...)
                } else {
                    v := node.Values
                    sq.Content = &v
                }

            case "read":
                if sq.Read != nil {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'read' has already been set.",
                    }
                    return
                }

                valueNum := len(node.Values)
                if valueNum > 1 {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'read' could only have one value.",
                    }
                    return
                }

                switch strings.ToLower(node.Values[0]) {
                    case "true": fallthrough
                    case "yes":
                        sq.Read = &t

                    case "false": fallthrough
                    case "no":
                        sq.Read = &f

                    case "any":
                        readAny = true

                    default:
                        err = &SearchQueryError {
                            Node: node,
                            Msg: fmt.Sprintf("Value of 'read' is not correct: %s", node.Values[0]),
                        }
                        return
                }

            case "orderby":
                if sq.Orderby != nil {
                    *sq.Orderby = append(*sq.Orderby, node.Values...)
                } else {
                    v := node.Values
                    sq.Orderby = &v
                }

            case "order":
                if sq.Asc != nil {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'order' has already been set.",
                    }
                    return
                }

                valueNum := len(node.Values)
                if valueNum > 1 {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'order' could only have one value.",
                    }
                    return
                }

                switch strings.ToLower(node.Values[0]) {
                    case "asc":
                        sq.Asc = &t
                    case "desc":
                        sq.Asc = &f
                    default:
                        err = &SearchQueryError {
                            Node: node,
                            Msg: fmt.Sprintf("Value of 'order' is not correct: %s", node.Values[0]),
                        }
                        return
                }

            case "tag":
                if sq.Tag != nil {
                    *sq.Tag = append(*sq.Tag, node.Values...)
                } else {
                    v := node.Values
                    sq.Tag = &v
                }

            case "starred":
                if sq.Starred != nil {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'starred' has already been set.",
                    }
                    return
                }

                valueNum := len(node.Values)
                if valueNum > 1 {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'starred' could only have one value.",
                    }
                    return
                }

                switch strings.ToLower(node.Values[0]) {
                    case "true": fallthrough
                    case "yes":
                        sq.Starred = &t

                    case "false": fallthrough
                    case "no":
                        sq.Starred = &f

                    default:
                        err = &SearchQueryError {
                            Node: node,
                            Msg: fmt.Sprintf("Value of 'starred' is not correct: %s", node.Values[0]),
                        }
                        return
                }

            case "num":
                if sq.Num != nil {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'num' has already been set.",
                    }
                    return
                }

                valueNum := len(node.Values)
                if valueNum > 1 {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: "'num' could only have one value.",
                    }
                    return
                }

                num, e := strconv.Atoi(node.Values[0])
                if e != nil || num <= 0 {
                    err = &SearchQueryError {
                        Node: node,
                        Msg: fmt.Sprintf("Incorrect num value: %s", node.Values[0]),
                    }
                    return
                }
                sq.Num = &num


            default:
                err = &SearchQueryError {
                    Node: node,
                    Msg: fmt.Sprintf("Do not support key: %s", node.Key),
                }
                return
        }
    }

    // set default values
    if sq.Orderby == nil || len(*sq.Orderby) == 0 {
        sq.Orderby = &[]string{"Id"}
    }
    if sq.Asc == nil {
        sq.Asc = &f // desc
    }
    if sq.Read == nil && readAny == false {
        sq.Read = &f // unread
    }

    // remove duplicate values

    if sq.Fid != nil {
        *sq.Fid = slice.Unique(*sq.Fid).([]int64)
    }

    if sq.Title != nil {
        *sq.Title = slice.Unique(*sq.Title).([]string)
    }

    if sq.Content != nil {
        *sq.Content = slice.Unique(*sq.Content).([]string)
    }

    if sq.Orderby != nil {
        *sq.Orderby = slice.Unique(*sq.Orderby).([]string)
    }

    if sq.Tag != nil {
        *sq.Tag = slice.Unique(*sq.Tag).([]string)
    }

    return
}


// Get article list of search query.
func (this *SearchQuery) List(page int) (list ArticleList, err error) {

    session := global.Orm.NewSession()
    defer session.Close()

    err = session.Begin()
    if err != nil {
        return
    }

    var fids []int64
    if this.Tag != nil && len(*this.Tag) > 0 {
        fids, err = getFeedIdsByTags(session, *this.Tag)
        if err != nil {
            session.Rollback()
            return
        }
    }

    if this.Fid != nil {
        fids = append(fids, *this.Fid...)
    }

    var where []string
    if len(fids) > 0 {
        var s []string
        for _, fid := range fids {
            s = append(s, strconv.FormatInt(fid, 10))
        }
        where = append(where, fmt.Sprintf("Item.Fid in (%s)", strings.Join(s, ",")))
    }

    var keywordSql []string // for combine sql of title and content

    // the code below escape the single quotation for used in sql.

    if this.Title != nil && len(*this.Title) > 0 {
        for _, title := range *this.Title {
            keywordSql = append(keywordSql, fmt.Sprintf("Item.Title like '%%%s%%'",
                strings.Replace(title, "'", "''", -1)))
        }
    }

    if this.Content != nil && len(*this.Content) > 0 {
        for _, content := range *this.Content {
            keywordSql = append(keywordSql, fmt.Sprintf("Item.Content like '%%%s%%'",
                strings.Replace(content, "'", "''", -1)))
        }
    }

    if len(keywordSql) > 0 {
        // where clause of title and content
        where = append(where, "(" + strings.Join(keywordSql, " or ") + ")")
    }

    if this.Read != nil {
        if *this.Read {
            where = append(where, "Item.Read=1")
        } else {
            where = append(where, "Item.Read=0")
        }
    }

    if this.Starred != nil {
        if *this.Starred {
            where = append(where, "Item.Starred=1")
        } else {
            where = append(where, "Item.Starred=0")
        }
    }

    whereSql := strings.Join(where, " and ")

    sql := "select count(*) from Item inner join Feed on Item.Fid=Feed.Id"
    if len(whereSql) > 0 {
        sql += " where " + whereSql
    }

    err = session.DB().QueryRow(sql).Scan(&list.Number)
    if err != nil {
        session.Rollback()
        return
    }

    // place Feed.* as the last column to fit list.Articles structure.
    sql = `select Item.Id, Item.Fid, Item.Author, Item.Url, Item.Guid, Item.Title, Item.PubTime,
                Item.FetchTime, Item.Starred, Item.Read, Item.Hash, Feed.* from Item
                inner join Feed on Item.Fid=Feed.Id`

    if len(whereSql) > 0 {
        sql += " where " + whereSql
    }

    if this.Asc != nil && this.Orderby != nil {
        var cols []string
        for _, e := range *this.Orderby {
            cols = append(cols, "`Item`.`" + e + "`")
        }

        s := fmt.Sprintf("order by %s", strings.Join(cols, ","))

        if *this.Asc {
            s += " asc"
        } else {
            s += " desc"
        }

        sql += " " + s
    }

    if this.Num == nil {
        err = fmt.Errorf("'num' not provide for search query.")
        session.Rollback()
        return
    }

    sql = fmt.Sprintf("%s limit %d, %d", sql, (page - 1) * *this.Num, *this.Num)

    err = session.Sql(sql).Find(&list.Articles)

    session.Commit()

    return
}

