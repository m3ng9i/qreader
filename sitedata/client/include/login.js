$(document).ready(function() {

    // check api token
    var checkToken = function(token) {
        var result = {};
        result.ok = false;
        result.error = "";

        $.ajax({
            url: "/api/checktoken",
            headers: { "X-QReader-Token": token },
            async: false,
            success: function(data) {
                if (data.success) {
                    result.ok = true;
                } else {
                    result.error = "密码错误，或系统时间与服务器时间时差较大。";
                }
            },
            error: function(data) {
                result.error = "错误：" + data;
            }
        });

        return result;
    };

    $("#container form").submit(function() {
        var password = $(this).find("input[name='password']")[0].value;
        var authToken = QReader.AuthToken(password);

        var result = checkToken(QReader.ApiToken(authToken));
        if (result.ok == true) {
            localStorage.authToken = authToken;
            window.location.href = "/";
        } else {
            $("#error div").text(result.error);
            $("#error").css("display", "block");
        }

        return false;
    });

    // check token after page load
    if (checkToken(QReader.ApiToken).ok == true) {
        window.location.href = "/";
    }

});

