package global

import "strings"
import "unicode"
import "os"
import "path/filepath"
import "runtime"


var defaultConfigIni = `
# Listen IP address of http server. Two or more IPs are separated by comma.
# If ip is auto, QReader will get the sever's IP Automaticly.
# Example:
#   ip = 127.0.0.1,192.168.1.123
#   ip = 0.0.0.0
#   ip = auto
ip = auto

# Listen port of http server
port = 4664

# Set usetls to true to use https, set to false to use http
usetls = false

# Path of logfile. If you want to output to stdout, leave it empty.
logfile =

# Log level: DEBUG, NOTICE, INFO, WARN, ERROR, FATAL
loglevel = INFO

# Permission of generated files
permission = 640

# Password
password =

# Used for hash
salt = 34682084954d47239577b53caad5baf4

# Debug mode
debug = false

# Socks5 proxy address. Example: 127.0.0.1:8080
proxy =

# Socks5 proxy username
proxy_username =

# Socks5 proxy password
proxy_password =

# always, try or never use proxy to fetch feed
# always: use proxy to fetch feed for each connection
# try: first use normal connection to fetch feed, if got error, try to use proxy to fetch
# never: use normal connection to fetch feed.
use_proxy = try
`

func DefaultConfigIni() string {
    s := strings.TrimLeftFunc(defaultConfigIni, unicode.IsSpace)
    if runtime.GOOS == "windows" {
        s = strings.Replace(s, "\n", "\r\n", -1)
    }
    return s
}


func CreateConfigIni() error {
    file, err := os.OpenFile(filepath.Join(Sitedata, "config.ini"),
                             os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
                             Permission)
    defer file.Close()
    if err == nil {
        _, err = file.WriteString(DefaultConfigIni())
    }
    return err
}


func IsConfigIniExist() bool {
    info, err := os.Stat(filepath.Join(Sitedata, "config.ini"))
    if err != nil || info.IsDir() {
        return false
    }
    return true
}

