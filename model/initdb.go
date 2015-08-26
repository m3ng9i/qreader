package model

import "bytes"
import "github.com/m3ng9i/qreader/global"

// SQL script for create tables.
const createTablesSql = `
drop table if exists 'Feed';
drop table if exists 'Item';
drop table if exists 'Tag';

create table if not exists 'Feed' (
    'Id'                integer not null primary key autoincrement,     -- primary key
    'Name'              text not null,                                  -- name of feed
    'Alias'             text not null default '',                       -- feed name's alias
    'Feedurl'           text not null,                                  -- url of feed
    'Url'               text not null,                                  -- url the feed point to
    'Desc'              text not null default '',                       -- feed description
    'Type'              text not null,                                  -- feed type: rss or atom
    'Interval'          integer not null default 0,                     -- refresh interval (minute), 0 for default interval. value below zero means not update.
    'LastFetch'         datetime not null,                              -- last successful fetch time (timestamp)
    'LastFailed'        datetime not null,                              -- last failed time for fetching (timestamp)
    'LastError'         text not null default '',                       -- last error for fetching
    'MaxUnread'         integer not null default 0,                     -- max number of unread items. 0 for keep all.
    'MaxKeep'           integer not null default 0,                     -- max number of items to keep. 0 for keep all, greater than 0 for keep n unread items.
    'Filter'            text not null default '',                       -- filter. (not to use now)
    'UseProxy'          integer not null default 0,                     -- whether to use proxy to fetch feed. 0: try, 1: always, 2: never.
    'Note'              text not null default ''                        -- comments for this feed (not to use now)
);

create table if not exists 'Item' (
    'Id'                integer not null primary key autoincrement,     -- primary key
    'Fid'               integer not null,                               -- Feed.id
    'Author'            text not null,                                  -- author
    'Url'               text not null,                                  -- url of the item
    'Guid'              text not null,                                  -- guid of the item
    'Title'             text not null,                                  -- title
    'Content'           text not null,                                  -- content
    'PubTime'           datetime not null,                              -- item pubtime
    'FetchTime'         datetime not null,                              -- item fetch time
    'Starred'           integer not null default 0,                     -- whether the item was starred. 0:no, 1:yes.
    'Read'              integer not null default 0,                     -- whether the item has been read. 0:no, 1:yes
    'Hash'              text not null                                   -- md5sum of content
);

create table if not exists 'Tag' (
    'Id'                integer not null primary key autoincrement,     -- primary key
    'Fid'               integer not null,                               -- Feed.id
    'Name'              text collate nocase not null                    -- tag name, case insensitive
);
`


// SQL script for create tables.
const createIndexesSql = `
create unique index if not exists i_feed_url on Feed(Feedurl);

create unique index if not exists i_item_combine_guid on Item(Fid, Guid);

create unique index if not exists i_tag_combine_name_fid on Tag(Name, Fid);
`


// Create tables for QReader, this will drop them first, you may lost data if the tables are already exists.
func CreateTables() error {
    sql := "begin;" + createTablesSql + "commit;"
    _, err := global.Orm.Import(bytes.NewReader([]byte(sql)))
    return err
}


// Create indexes.
func CreateIndexes() error {
    sql := "begin;" + createIndexesSql + "commit;"
    _, err := global.Orm.Import(bytes.NewReader([]byte(sql)))
    return err
}


// QReader database initialization
func InitDB() error {
    sql := "begin;" + createTablesSql + createIndexesSql + "commit;"
    _, err := global.Orm.Import(bytes.NewReader([]byte(sql)))
    return err
}

