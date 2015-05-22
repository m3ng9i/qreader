// Global variables and configuration of QReader.
package global

import "fmt"
import "os"
import "sync"
import "path/filepath"
import "strconv"
import "strings"
import "net"
import "github.com/Unknwon/goconfig"
import "github.com/go-xorm/xorm"
import "github.com/go-xorm/core"
import _ "github.com/mattn/go-sqlite3"
import "github.com/m3ng9i/go-utils/log"
import h "github.com/m3ng9i/go-utils/http"


type ProxyType string
const PROXY_ALWAYS  ProxyType = "always"
const PROXY_TRY     ProxyType = "try"
const PROXY_NEVER   ProxyType = "never"

type VersionType struct {
    Version     string // program version, from git tag
    Branch      string // git branch
    CommitId    string // git commit id
    BuildTime   string // build time
}

var Sitedata        string          // Directory of sitedata

var PathRoot        string          // Root directory of sitedata
var PathClient      string          // Directory of javascript client
var PathDB          string          // Path of database
var PathCertPem     string          // Path of cert.pem
var PathKeyPem      string          // Path of key.pem

var ConfigFile      string          // path of config file

var IP              string              // IP of http server
var Port            uint                // Port of http server
var Usetls          bool                // Set to true to use https, set to false to use http
var Password        string              // Password to log into QReader
var ProxyConfig     *h.ProxyConfig      // Proxy config. It's nil if no proxy is configured.
var UseProxy        ProxyType           // if use proxy, always, try or never
var Debug           bool                // If enable debug mode.
var Salt            string              // Used for authentication
var Permission      os.FileMode = 0640  // Permission of generated files
var Logger          *log.Logger         // Logger
var Orm             *xorm.Engine        // Xorm database engine
var NormalFetcher   *h.Fetcher          // Normal fetcher
var Socks5Fetcher   *h.Fetcher          // Socks5 proxy fetcher
var Version         VersionType

var Github          string

var loglevel        log.LevelType
var logfile         string


func createLogger(file string, l log.LevelType, p os.FileMode) (logger *log.Logger, err error) {
    var config log.Config
    config.Level = l
    config.TimeFormat = log.TF_DEFAULT
    config.Utc = false

    var f *os.File
    if file == "" {
        f = os.Stdout
    } else {
        f, err = log.OpenFile(file, p)
        if err != nil {
            return
        }
    }

    logger, err = log.New(f, config)
    return
}


// Read config file.
func loadConfig(filename string) error {

    var err error

    ConfigFile, err = filepath.Abs(filename)
    if err != nil {
        return err
    }

    c, err := goconfig.LoadConfigFile(filename)
    if err != nil {
        return fmt.Errorf("Cannot open config file: %s\n", filename)
    }

    IP = c.MustValue("", "ip")
    if IP == "auto" {
        IP, err = getIPAutomaticly()
        if err != nil {
            return err
        }
        if IP == "" {
            return fmt.Errorf("Cannot get IP address.\n")
        }
    } else if IP == "" {
        IP = "127.0.0.1"
    }

    Port = uint(c.MustInt("", "port"))
    if Port == 0 {
        return fmt.Errorf("Port cannot be 0.\n")
    }

    Usetls      = c.MustBool("", "usetls", false)
    Password    = c.MustValue("", "password")
    Debug       = c.MustBool("", "debug", false)
    Salt        = c.MustValue("", "salt")

    var proxy = new(h.ProxyConfig)
    proxy.Addr = c.MustValue("", "proxy")
    proxy.Username = c.MustValue("", "proxy_username")
    proxy.Password = c.MustValue("", "proxy_password")
    if proxy.Addr != "" {
        ProxyConfig = proxy
    }

    if ProxyConfig == nil {
        UseProxy = PROXY_NEVER
    } else {
        UseProxy = ProxyType(strings.ToLower(c.MustValue("", "use_proxy", "never")))
        if UseProxy != PROXY_ALWAYS && UseProxy != PROXY_TRY {
            UseProxy = PROXY_NEVER
        }
    }

    value := c.MustValue("", "permission")
    p, err := strconv.ParseUint(value, 8, 0)
    if err != nil {
        return fmt.Errorf("Value of Permission is not legal.\n")
    }
    Permission = os.FileMode(p)

    value = c.MustValue("", "loglevel")
    var ok bool
    loglevel, ok = log.String2Level(value)
    if !ok {
        return fmt.Errorf("loglevel is not correct.\n")
    }
    logfile = c.MustValue("", "logfile")

    return nil
}


// Get first IPv4 address in system's network interface.
// This may be a Lan IP or a public IP.
func getIPAutomaticly() (a string, e error) {
    addr, e := net.InterfaceAddrs()
    if e != nil {
        return
    }
    for _, i := range addr {
        ip := net.ParseIP(strings.SplitN(i.String(), "/", 2)[0])
        ipString := ip.String()
        if ip.To4() != nil && !ip.IsLoopback() && ipString != "0.0.0.0" {
            a = ipString
            goto END
        }
    }

    END:
    if a == "" {
        a = "127.0.0.1"
    }
    return
}


var once1, once2 sync.Once


// Init step 1: set path and database
func Init1() {
    once1.Do(func() {
        var err error
        PathRoot, err = filepath.Abs(Sitedata)
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
            os.Exit(1)
        }

        PathClient  = filepath.Join(PathRoot, "client")
        PathDB      = filepath.Join(PathRoot, "feed.db")
        PathCertPem = filepath.Join(PathRoot, "cert", "cert.pem")
        PathKeyPem  = filepath.Join(PathRoot, "cert", "key.pem")

        // set database

        Orm, err = xorm.NewEngine("sqlite3", PathDB)
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
            os.Exit(1)
        }
        Orm.SetMapper(core.SameMapper{})
    })
}


// Init step 2: read config file and create logger
func Init2() {
    once2.Do(func() {
        err := loadConfig(filepath.Join(Sitedata, "config.ini"))
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
            os.Exit(1)
        }

        headers := make(map[string]string)
        headers["User-Agent"] = fmt.Sprintf("QReader %s (%s)", Version.Version, Github)

        NormalFetcher = h.NewFetcher(nil, headers)

        if UseProxy != PROXY_NEVER {
            socks5client, err := h.Socks5Client(*ProxyConfig)
            if err != nil {
                fmt.Fprintf(os.Stderr, err.Error())
                os.Exit(1)
            }

            Socks5Fetcher = h.NewFetcher(socks5client, headers)
        }

        if Debug {
            Orm.ShowSQL = true
        }

        // create logger
        Logger, err = createLogger(logfile, loglevel, Permission)
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
            os.Exit(1)
        }
    })
}
