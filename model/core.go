package model

import "time"


/* Map to table "Feed"
Alias, Filter and Note is pointer to string. When update use xorm, if it's value is nil, it means no need to up update this field.
If it's pointer to empty string, it means update this field and set it to "".
*/
type Feed struct {
    Id          int64       `json:"feed_id"             xorm:"pk autoincr"`                 // primary key
    Name        string      `json:"feed_name"           xorm:"notnull"`                     // name of feed
    Alias       *string     `json:"feed_alias"          xorm:"notnull default ''"`          // feed name's alias
    FeedUrl     string      `json:"feed_feed_url"       xorm:"notnull unique"`              // url of feed
    Url         string      `json:"feed_url"            xorm:"notnull"`                     // url the feed point to
    Desc        string      `json:"feed_desc"           xorm:"notnull default ''"`          // feed description
    Type        string      `json:"feed_type"           xorm:"notnull"`                     // feed type: rss or atom
    Interval    *int        `json:"feed_interval"       xorm:"notnull default 0"`           // refresh interval (minute), 0 for default interval. value below zero means not update.
    LastFetch   time.Time   `json:"feed_last_fetch"     xorm:"notnull"`                     // last successful fetch time
    LastFailed  time.Time   `json:"feed_last_failed"    xorm:"notnull"`                     // last failed time for fetching
    LastError   string      `json:"feed_last_error"     xorm:"notnull default ''"`          // last error for fetching
    MaxUnread   *uint       `json:"feed_max_unread"     xorm:"notnull default 0"`           // max number of unread items. 0 for keep all.
    MaxKeep     *uint       `json:"feed_max_keep"       xorm:"notnull default 0"`           // max number of items to keep. 0 for keep all, greater than 0 for keep n unread items.
    Filter      *string     `json:"feed_filter"         xorm:"notnull default ''"`          // filter. (not to use now)
    UseProxy    int         `json:"feed_use_proxy"      xorm:"notnull default 0"`           // whether to use proxy to fetch feed, 0: try, 1: always, 2: never
    Note        *string     `json:"feed_note"           xorm:"notnull default ''"`          // comments for this feed
}


// Map to table "Item"
type Item struct {
    Id          int64       `json:"item_id"             xorm:"pk autoincr"`                 // primary key
    Fid         int64       `json:"item_fid"            xorm:"notnull unique(Fid_Guid)"`    // Feed.Id
    Author      string      `json:"item_author"         xorm:"notnull"`                     // author
    Url         string      `json:"item_url"            xorm:"notnull unique"`              // url of the item
    Guid        string      `json:"item_guid"           xorm:"notnull unique(Fid_Guid)"`    // guid of the item
    Title       string      `json:"item_title"          xorm:"notnull"`                     // title
    Content     string      `json:"item_content"        xorm:"notnull"`                     // content
    PubTime     time.Time   `json:"item_pub_time"       xorm:"notnull"`                     // item pubtime
    FetchTime   time.Time   `json:"item_fetch_time"     xorm:"notnull"`                     // item fetch time
    Starred     bool        `json:"item_starred"        xorm:"notnull default 0"`           // whether the item was starred
    Read        bool        `json:"item_read"           xorm:"notnull default 0"`           // whether the item has been read
    Hash        string      `json:"-"                   xorm:"notnull"`                     // md5sum of content
}


// Map to table "Tag"
type Tag struct {
    Id          int64       `xorm:"pk autoincr"`                // primary key
    Name        string      `xorm:"notnull unique(Name_Fid)"`   // tag name, case insensitive
    Fid         int64       `xorm:"notnull unique(Name_Fid)"`   // Feed.Id
}


