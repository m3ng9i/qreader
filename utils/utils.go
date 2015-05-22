package utils

import "crypto/sha1"
import "fmt"
import "time"
import "../global"
import "github.com/microcosm-cc/bluemonday"
import "github.com/m3ng9i/go-utils/timeslot"


func hash(s string) string {
    h := sha1.New()
    fmt.Fprint(h, s)
    return fmt.Sprintf("%x", h.Sum(nil))
}


func ValidateToken(token string) bool {
    if token == "" {
        return false
    }

    auth := hash(global.Password + global.Salt)

    t           := time.Now().UTC()
    now         := time.Date(t.Year(), t.Month(), 15, 0, 0, 0, 0, time.UTC)
    current     := now.Format("200601")
    previous    := now.AddDate(0, -1, 0).Format("200601")
    next        := now.AddDate(0, 1, 0).Format("200601")
    ts          := timeslot.Default()

    if  token == hash(hash(current   + auth) + ts.String()               + global.Salt) ||
        token == hash(hash(current   + auth) + ts.Previous().String()    + global.Salt) ||
        token == hash(hash(current   + auth) + ts.Next().String()        + global.Salt) ||
        token == hash(hash(previous  + auth) + ts.String()               + global.Salt) ||
        token == hash(hash(previous  + auth) + ts.Previous().String()    + global.Salt) ||
        token == hash(hash(previous  + auth) + ts.Next().String()        + global.Salt) ||
        token == hash(hash(next      + auth) + ts.String()               + global.Salt) ||
        token == hash(hash(next      + auth) + ts.Previous().String()    + global.Salt) ||
        token == hash(hash(next      + auth) + ts.Next().String()        + global.Salt) {
       return true
    }
    return false
}


func CurrentToken() string {
    auth := hash(global.Password + global.Salt)

    t           := time.Now().UTC()
    now         := time.Date(t.Year(), t.Month(), 15, 0, 0, 0, 0, time.UTC)
    current     := now.Format("200601")
    ts          := timeslot.Default()

    return hash(hash(current + auth) + ts.String() + global.Salt)
}


var strictPolicy *bluemonday.Policy // for title, name, etc
var normalPolicy *bluemonday.Policy // for article content


// Sanitize html codes and return clean and harmless string.
func Sanitize(s string, normal ...bool) string {
    if len(normal) > 0 && normal[0] == true {
        return normalPolicy.Sanitize(s)
    } else {
        return strictPolicy.Sanitize(s)
    }
}


// Sanitize html codes directly.
func SanitizeSelf(s *string, normal ...bool) {
    if len(normal) > 0 && normal[0] == true {
        *s = normalPolicy.Sanitize(*s)
    } else {
        *s = strictPolicy.Sanitize(*s)
    }
}


func init() {
    strictPolicy = bluemonday.StrictPolicy()
    normalPolicy = bluemonday.UGCPolicy()
}
