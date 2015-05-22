var NAMESPACE = NAMESPACE || {};


NAMESPACE.CurrentTimeSlot = function(size) {

    // If size if not available, set it to a default value.
    size = size || 5;

    var ErrSize = "Size of slot must between 1 and 30, and 60 could be divided exactly by it.";

    if (size < 1 || size > 30 || 60 % size != 0) {
        throw ErrSize;
    }

    // Convert a number to string, and pad it with leading zero, make sure the length of return value is 2.
    var padZero = function(num) {
        num = num.toString();
        if (num.length < 2) {
            num = "0" + num;
        }
        return num;
    };

    now = new Date();

    var m = now.getMinutes();
    if (now.getSeconds() > 0) {
        m++;
    }

    var slot = now.getUTCFullYear().toString();
    slot += padZero(now.getUTCMonth() + 1);
    slot += padZero(now.getUTCDate());
    slot += padZero(now.getUTCHours());
    slot += padZero(Math.floor(m / size));

    return slot;
};


// Readable file size
NAMESPACE.ReadableFilesize = function(bytes, precision) {
    if (bytes <= 0 || isNaN(parseFloat(bytes)) || !isFinite(bytes)) {
        return '-';
    }
    if (typeof precision === 'undefined') {
        precision = 2;
    }
    var units = ['bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
    var number = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, Math.floor(number))).toFixed(precision) +  ' ' + units[number];
};

