package api

import "net/http"
import "github.com/go-martini/martini"
import httphelper "github.com/m3ng9i/go-utils/http"


type ApiError struct {
    ErrCode uint    `json:"errcode"`    // 0 for no error
    ErrMsg  string  `json:"errmsg"`
}


func (this *ApiError) Error() string {
    return this.ErrMsg
}

/*
1xx     authentication, query string error
2xx     feed fetch or parse error
3xx     database error
4xx     system error, like io, filesystem, etc.
*/

var ErrTokenInvalid         = ApiError{100, "Client token is invalid. Please make sure you have the permission to use QReader."}
var ErrRequestNotAllowd     = ApiError{101, "The request is not allowed."}
var ErrBadRequest           = ApiError{102, "Request query or post data not correct."}
var ErrFetchError           = ApiError{200, "Error occurs when fetching feed. Please check the internet connection and make sure the feed's url is valid."}
var ErrParseError           = ApiError{201, "Error occurs when parsing feed. Please check if the feed is valid."}
var ErrQueryDB              = ApiError{300, "Error occurs when querying the database."}
var ErrAlreadySubscribed    = ApiError{301, "Feed is already subscribed, cannot be subscribed again."}
var ErrNoResultsFound       = ApiError{302, "No results found."}
var ErrNoDataChanged        = ApiError{303, "No data changed."}
var ErrFeedCannotBeDeleted  = ApiError{304, "Feed has starred items, cannot be deleted."}
var ErrSystemError          = ApiError{400, "System error."}
var ErrUnexpectedError      = ApiError{999, "Unexpected error."}


// Default api handler.
func Default() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid
        result.Success = false
        result.Error = ErrRequestNotAllowd
        result.Response(w)
    }
}


// Indicate the QReader api is up and running.
func Status() martini.Handler {
    return func(w http.ResponseWriter, r *http.Request, rid httphelper.RequestId) {
        var result Result
        result.RequestId = rid
        result.Success = true
        result.Response(w)
    }
}

