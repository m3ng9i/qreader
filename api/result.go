package api

import "net/http"
import "fmt"
import "encoding/json"
import httphelper "github.com/m3ng9i/go-utils/http"
import "github.com/m3ng9i/qreader/global"


// Result will be converted to json string and write to http.ResponseWriter.
type Result struct {
    RequestId   httphelper.RequestId    `json:"request_id"`
    Success     bool                    `json:"success"`
    Error       ApiError                `json:"error"`      // error to show to user
    Result      interface{}             `json:"result"`
    IntError    error                   `json:"-"`          // internal error, used to log
}


// Write json to ResponseWriter.
func (this *Result) Response(w http.ResponseWriter, rid ...httphelper.RequestId) {

    const errMarshal = `{"success":false,"error":"Cannot marshal json data.","result":null}`

    if len(rid) > 0 {
        this.RequestId = rid[0]
    }

    b, err := json.Marshal(this)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, errMarshal)
        global.Logger.Errorf("[API] Cannot marshal json data: %v", this)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK) // always return status 200 even if an error occurs
    w.Write(b)

    if this.IntError != nil {
        global.Logger.Errorf("[API Response] [#%s] [internal error: %s] %s", this.RequestId, this.IntError, string(b))
    } else {
        global.Logger.Debugf("[API Response] [#%s] %s", this.RequestId, string(b))
    }
}

