package main

import "time"
import "flag"
import "os"
import "os/signal"
import "fmt"
import "strings"
import "net/http"
import "syscall"
import "path/filepath"
import "github.com/toqueteos/webbrowser"
import "github.com/m3ng9i/qreader/global"
import "github.com/m3ng9i/qreader/model"
import "github.com/m3ng9i/qreader/utils"
import "github.com/m3ng9i/qreader/server"


var _version_   = "v0.2.2"              // program version, from git tag
var _branch_    = "unknown"             // git branch
var _commitId_  = "0000000"             // git commit id
var _buildTime_ = "0000-00-00 00:00"    // build time

var Version = fmt.Sprintf("Version: %s, Branch: %s, Build: %s, Build time: %s",
        _version_, _branch_, _commitId_, _buildTime_)

var _github_ = "https://github.com/m3ng9i/qreader"

func usage() {
s := `QReader: a browser-server based feed reader

Usage:
    qreader [option...]

Options:
    -s, -sitedata <sitedata>    Directory of sitedata. If not provided, use "sitedata" under current wokring directory.
    -init                       Initialize QReader database and config.ini.
    -initdb                     Initialize QReader database, will delete all the data and recreate tables.
    -current-token              Show current api token.
    -defini                     Default content of config.ini.
    -open                       Open QReader web page on default browser.
    -h, -help                   Show this message.
    -v, -version                Show version information.

Github:
    <%s>

Author:
    m3ng9i <http://mengqi.info>
`
fmt.Printf(s, _github_)
os.Exit(0)
}


func checkDBFile() error {
    p := global.PathDB
    init := "You can use -initdb parameter to initialize database."
    info, err := os.Stat(p)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("Database: '%s' is not exist.\n%s\n", p, init)
        }
        return err
    }
    if info.IsDir() {
        return fmt.Errorf("'%s' is a directory, can not be used as database.\n", p)
    }
    if info.Size() == 0 {
        return fmt.Errorf("Size of database: '%s' is 0.\n%s\n", p, init)
    }
    return nil
}


func catchSignal() {
    signal_channel := make(chan os.Signal, 1)
    signal.Notify(signal_channel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
    go func() {
        for value := range signal_channel {
            global.Logger.Warnf("Catch signal: %s, QReader server is going to shutdown", value.String())
            global.Logger.Wait()
            os.Exit(0)
        }
    }()
}


func initDatabase() {
    err := model.InitDB()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error occurs when initializing database: %s\n", err.Error())
        os.Exit(1)
    } else {
        fmt.Println("Database initialized.")
    }
}


func initConfigIni() {
    err := global.CreateConfigIni()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error occurs when initializing config.ini: %s\n", err.Error())
        os.Exit(1)
    } else {
        fmt.Println("config.ini initialized.")
    }
}


func main() {

    global.Version = global.VersionType {
        Version:    _version_,
        Branch:     _branch_,
        CommitId:   _commitId_,
        BuildTime:  _buildTime_,
    }

    global.Github = _github_

    var sitedata, input string
    var init, initdb, help, version, currentToken, defini, open bool
    flag.StringVar(&sitedata, "sitedata", "", "Directory of sitedata")
    flag.StringVar(&sitedata, "s", "", "Directory of sitedata")
    flag.BoolVar(&init, "init", false, "-init")
    flag.BoolVar(&initdb, "initdb", false, "-initdb")
    flag.BoolVar(&help, "h", false, "-h")
    flag.BoolVar(&help, "help", false, "-help")
    flag.BoolVar(&version, "v", false, "-v")
    flag.BoolVar(&version, "version", false, "-version")
    flag.BoolVar(&currentToken, "current-token", false, "-current-token")
    flag.BoolVar(&defini, "defini", false, "-defini")
    flag.BoolVar(&open, "open", false, "-open")
    flag.Usage = usage
    flag.Parse()

    if help {
        usage()
    }

    if version {
        fmt.Println(Version)
        os.Exit(0)
    }

    if defini {
        fmt.Println(global.DefaultConfigIni())
        os.Exit(0)
    }

    // If sitedata is not provided, use default path.
    if sitedata == "" {
        sitedata = "sitedata"
    }
    global.Sitedata = sitedata

    // set path and database
    global.Init1()

    configIniExist := global.IsConfigIniExist()
    if configIniExist {
        // load config and create logger
        global.Init2()
        defer func() {
            global.Logger.Wait()
        }()
    }

    // initialize database and config.ini
    if init {
        fmt.Print("Are you sure to initialize QReader database and config.ini? This will delete all existing data and reset config.ini. ")
        fmt.Scanln(&input)
        if len(input) > 0 && strings.ToLower(string(input[0])) == "y" {
            initDatabase()
            initConfigIni()
        } else {
            fmt.Fprintln(os.Stderr, "Aborted to initialize database and config.ini.")
        }
        os.Exit(0)
    }

    // initialize database
    if initdb {
        fmt.Print("Are you sure to initialize QReader database? This will delete all existing data. ")
        fmt.Scanln(&input)
        if len(input) > 0 && strings.ToLower(string(input[0])) == "y" {
            initDatabase()
        } else {
            fmt.Fprintln(os.Stderr, "Aborted to initialize database.")
        }
        os.Exit(0)
    }

    if !configIniExist {
        fmt.Fprintf(os.Stderr, "%s is not exist or not a regular file.\n", filepath.Join(global.Sitedata, "config.ini"))
        os.Exit(1)
    }

    // check if database is correct
    err := checkDBFile()
    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }

    if currentToken {
        fmt.Println(utils.CurrentToken())
        os.Exit(0)
    }

    addr := fmt.Sprintf("%s:%d", global.IP, global.Port)
    url := ""
    if global.Usetls {
        url = "https://" + addr
    } else {
        url = "http://" + addr
    }

    catchSignal()

    server.Init()

    global.Logger.Infof("QReader %s.", Version)
    global.Logger.Infof("QReader is running. Open %s in your browser to use.", url)

    // Auto update feed. Feed will be updated every 120 minutes (2 hours) default.
    model.AutoUpdateFeed(120)

    if open {
        go func() {
            <- time.After(500 * time.Millisecond)
            err = webbrowser.Open(url)
            if err != nil {
                global.Logger.Error(err)
                err = nil
            }
        }()
    }

    if global.Usetls {
        err = http.ListenAndServeTLS(addr, global.PathCertPem, global.PathKeyPem, server.Mux)
    } else {
        err = http.ListenAndServe(addr, server.Mux)
    }
    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }
}
