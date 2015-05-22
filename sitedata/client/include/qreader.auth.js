var QReader = QReader || {};


QReader.salt = "34682084954d47239577b53caad5baf4";


QReader.Hash = function(str) {
    var h = forge.md.sha1.create();
    h.update(str);
    return h.digest().toHex();
};


// Auth token will be saved in localStorage
QReader.AuthToken = function(password) {
    return this.Hash(password + this.salt);
};


// Auth token will be encoded to api token and send to api server.
// If no auth token found, this function will return empty string.
QReader.ApiToken = function(authToken) {

    var auth = "";
    if (arguments.length > 0) {
        auth = authToken;
    } else {
        var auth = localStorage.authToken || "";
    }
    if (auth == "") {
        return "";
    }

    var d = new Date();
    var month = d.getUTCMonth() + 1;
    if (month < 10) {
        month = "0" + String(month);
    } else {
        month = String(month);
    }
    var current = String(d.getUTCFullYear()) + month;

    var ts = NAMESPACE.CurrentTimeSlot();

    return this.Hash(this.Hash(current + auth) + ts + this.salt);
};

