var QReader = QReader || {};

QReader.apiroot                 = "/api/";

QReader.api                     = {}
QReader.api.status              = QReader.apiroot;
QReader.api.subscription        = QReader.apiroot + "feed/subscription";
QReader.api.feedlist            = QReader.apiroot + "feed/list";
QReader.api.feed                = QReader.apiroot + "feed/id/";
QReader.api.articlesRandom      = QReader.apiroot + "articles/random";
QReader.api.articlesUnread      = QReader.apiroot + "articles/unread/";
QReader.api.articlesFid         = QReader.apiroot + "articles/fid/";
QReader.api.articlesTag         = QReader.apiroot + "articles/tag/";
QReader.api.articlesStarred     = QReader.apiroot + "articles/starred/";
QReader.api.articlesSearch      = QReader.apiroot + "articles/search/";
QReader.api.article             = QReader.apiroot + "article/content/";
QReader.api.markArticleRead     = QReader.apiroot + "article/read/";
QReader.api.markArticleUnread   = QReader.apiroot + "article/unread/";
QReader.api.markArticlesRead    = QReader.apiroot + "articles/read";
QReader.api.markArticlesStarred = QReader.apiroot + "articles/starred";
QReader.api.tagsList            = QReader.apiroot + "tags/list";
QReader.api.settings            = QReader.apiroot + "system/settings";
QReader.api.shutdown            = QReader.apiroot + "system/shutdown";

QReader.app                     = angular.module("QReader", ["ngRoute", "ngSanitize"]);


// Application initialization.
// Keyboard shortcuts are defined in this section.
QReader.app.run(function($route, $location) {

    // r for reload
    Mousetrap.bind("r", function() {
        $route.reload();
    });

    // 0 or h for home
    Mousetrap.bind(["0", "h"], function() {
        $location.url("/articles/unread");
        $route.reload();
    });

    // t for go top
    Mousetrap.bind("t", function() {
        document.body.scrollTop=0;
    });

    // b for go bottom
    Mousetrap.bind("b", function() {
        document.body.scrollTop=document.body.scrollHeight;
    });
});


QReader.app.filter("ReadableFilesize", function() {
    var func = NAMESPACE.ReadableFilesize;
    func.$stateful = true;
    return func;
});


QReader.app.config(function($routeProvider, $httpProvider) {

    // /articles/unread?page={page}
    $routeProvider.when("/articles/unread", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    // /articles/fid/{fid}?page={page}
    $routeProvider.when("/articles/fid/:fid", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    // /articles/tag/{tag}?page={page}
    $routeProvider.when("/articles/tag/:tag", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    $routeProvider.when("/articles/random", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    // /articles/starred?page={page}
    $routeProvider.when("/articles/starred", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    // /articles/search?q={query}&page={page}
    $routeProvider.when("/articles/search", {
        templateUrl: "/include/article_list.tpl.html",
        controller: "ArticleListController"
    });

    $routeProvider.when("/article/:id", {
        templateUrl: "/include/article.tpl.html",
        controller: "ArticleController"
    });

    $routeProvider.when("/feed/info/:fid", {
        templateUrl: "/include/feed_info.tpl.html",
        controller: "FeedInfoController"
    });

    $routeProvider.when("/feed/list", {
        templateUrl: "/include/feed_list.tpl.html",
        controller: "FeedListController"
    });

    $routeProvider.when("/tags/list", {
        templateUrl: "/include/tags_list.tpl.html",
        controller: "TagsListController"
    });

    $routeProvider.when("/settings", {
        templateUrl: "/include/settings.tpl.html",
        controller: "SettingsController"
    });

    $routeProvider.otherwise({
        redirectTo: '/articles/unread'
    });


    $httpProvider.interceptors.push(function($q, QDoc) {
        return {
            'request': function(config) {
                // Add api token to each request's header.
                config.headers["X-QReader-Token"] = QReader.ApiToken();
                return config;
            },

            'response': function(response) {
                // if not login, or api token is not correct, redirect to login page.
                if (response.data.success == false) {
                    if (response.data.error.errcode == 100) {
                        window.location.href = "/login.html";
                        return;
                    }
                }
                return response;
            },

            'responseError': function(rejection) {
                if (rejection.status == 0) {
                    QDoc.SetError("读取服务器数据出错，请检查网络与 QReader 服务器状态。");
                } else if (rejection.status == 404) {
                    QDoc.SetError("404错误：访问的内容没有找到");
                }
                return $q.reject(rejection);
            },
        };
    });

});


QReader.app.factory("QDoc", function() {

    var d = {}

    // Init a page: clear error, set title
    d.Init = function(title) {
        this.ClearError();
        this.SetPageTitle(title);
        this.SetHtmlTitle(title);
    };

    // Set html title.
    d.SetHtmlTitle = function(title) {
        if (title == "" || title == "QReader") {
            title = "QReader";
        } else {
            title = title + " - QReader";
        }
        angular.element(document.querySelector("html>head>title")).html(title);
    };

    // Set page title.
    d.SetPageTitle = function(title) {
        angular.element(document.querySelector("#title")).html(title);
    };

    // Set error.
    d.SetError = function(error) {
        var icon = "<i class='fa fa-exclamation-triangle'></i>&nbsp;"
        angular.element(document.querySelector("#error > div")).html(icon + error);
        var error = document.getElementById("error");
        error.style.display = "block";
        error.title = "点击关闭";
    };

    // Clear error.
    d.ClearError = function() {
        document.getElementById("error").style.display = "none";
    };

    // Loading animation
    d.Loading = function(on, text) {
        var display = "none";
        if (on) {
            display = "block"
        }
        document.getElementById("loading").style.display = display;

        var t = text || "加载中";
        angular.element(document.querySelector("#loading > div > span:nth-child(2)")).html(t);
    };

    return d;
});


// Mark articles read by ids, feedid or tag.
QReader.app.factory("MarkArticlesRead", function($http, $route, QDoc) {

    return function(tp, value) {
        if (tp != "ids" && tp != "feedid" && tp != "tag") {
            QDoc.SetError("Parameter of 'MarkArticlesRead' error.");
            return;
        }

        var post = { "type": tp, "value": value }

        $http.put(QReader.api.markArticlesRead, post).success(function(data) {
            if (data.success) {
                $route.reload();
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        });
    };

});


QReader.app.controller("ArticleListController", function($location, $http, $routeParams, $scope, $route, QDoc, MarkArticlesRead) {

    var url = $location.url();
    var request_url = "";
    var route = "";
    var page = 0;

    $scope.data = {};
    $scope.data.Number = 0;
    $scope.data.page = 0;   // for article list page
    $scope.data.pagenum = 0; // for article list page

    if (url == "/" || url.match(/^\/articles\/unread/)) {
        QDoc.Init("文章列表");
        route = "unread";

    } else if (url.match(/^\/articles\/random/)) {
        route = "random";
        QDoc.Init("随机文章");

    } else if (url.match(/^\/articles\/fid/)) {
        route = "fid";
        QDoc.ClearError();

    } else if (url.match(/^\/articles\/tag/)) {
        route = "tag";
        QDoc.ClearError();

    } else if (url.match(/^\/articles\/starred/)) {
        route = "starred";
        QDoc.Init("加星文章");

    } else if (url.match(/^\/articles\/search/)) {
        route = "search";
        QDoc.Init("搜索");

    } else {
        QDoc.SetError("路由错误");
        return
    }

    var tag = "";
    var limit = localStorage.pageItemNumber || QReader.defaultStorage.pageItemNumber;

    if (route == "unread" || route == "fid" || route == "tag" || route == "starred" || route == "search") {
        page = parseInt($location.search().page || 1);
        var offset = (page - 1) * limit;

        if (route == "unread") {

            request_url = QReader.api.articlesUnread + limit + "/" + offset;

            $scope.pagelinkPrefix = "/#/articles/unread?";

        } else if (route == "fid") {

            var fid = $routeParams.fid || 0;
            if (fid == 0) {
                QDoc.SetError("路由错误：fid值无效。");
                return
            }

            request_url = QReader.api.articlesFid + fid + "/" + limit + "/" + offset;

            $scope.pagelinkPrefix = "/#/articles/fid/" + fid + "?";

        } else if (route == "tag") {

            tag = $routeParams.tag || "";
            if (tag == "") {
                QDoc.SetError("路由错误：tag值无效。");
                return
            }

            QDoc.Init(tag);

            request_url = QReader.api.articlesTag + tag + "/" + limit + "/" + offset;

            $scope.pagelinkPrefix = "/#/articles/tag/" + tag + "?";

        } else if (route == "starred") {

            request_url = QReader.api.articlesStarred + limit + "/" + offset;
            $scope.pagelinkPrefix = "/#/articles/starred?";

        } else if (route == "search") {

            var query = encodeURIComponent($location.search().q || "");
            if (query == "") {
                QDoc.SetError("路由错误：query 值无效。");
                return
            }

            request_url = QReader.api.articlesSearch + limit + "?q=" + query + "&page=" + page;
            $scope.pagelinkPrefix = "/#/articles/search?q=" + query + "&";
        }
    } else if (route == "random") {
        request_url = QReader.api.articlesRandom + "?n=" + limit;
    }

    var bindKeys = function(pagenum) {
        var url = $location.url();

        // p or ctrl+left, command+left: turn to previous page.
        Mousetrap.bind(["mod+left", "p"], function() {
            if (route == "unread" || route == "fid" || route == "tag" || route == "starred" || route == "search") {
                var page = parseInt($location.search().page || 1) - 1;
                if (page <= 0) {
                    page = pagenum;
                }
                $location.search("page", page);
                $route.reload();
            }
        });

        // n or ctrl+right, command+right: turn to next page.
        Mousetrap.bind(["mod+right", "n"], function() {
            if (route == "unread" || route == "fid" || route == "tag" || route == "starred" || route == "search") {
                var page = parseInt($location.search().page || 1) + 1;
                if (page > pagenum) {
                    page = 1;
                }
                $location.search("page", page);
                $route.reload();
            }
        });
    };


    $http.get(request_url).success(function(data) {
        if (data.success) {
            if (route == "unread" || route == "fid" || route == "tag" || route == "starred" || route == "search") {
                data.result.page = page;
                limit = data.result.limit || limit; // result of api.SearchList() contains a variable named limit
                data.result.pagenum = Math.ceil(data.result.Number / limit);
            }
            $scope.data = data.result;

            bindKeys(data.result.pagenum);

            var feedId = 0;
            var feedName = "";
            if (route == "fid") {
                feedName = data.result.Articles[0].feed_alias || data.result.Articles[0].feed_name;
                feedId = data.result.Articles[0].feed_id;
                QDoc.ClearError();
                QDoc.SetPageTitle("<a href='/#/feed/info/" + feedId + "' title='查看feed详情'>" + feedName + "</a>");
                QDoc.SetHtmlTitle(feedName);
            }

            $scope.route = route;

            // Select all or select nothing in article list.
            var select = function(sel) {
                angular.element(document.querySelectorAll("#article_list input[type='checkbox']")).prop("checked", sel);
            };

            $scope.select_all = function() {
                select(true);
            };

            $scope.select_none = function() {
                select(false);
            };

            $scope.markRead = function(tp) {
                if (tp == "selected" || tp == "page") {
                    var sel = null;
                    var indication = "";
                    if (tp == "selected") {
                        sel = document.querySelectorAll("#article_list input[type='checkbox']:checked");
                        indication = "确定要将已选条目标记为已读吗？";

                    } else {
                        sel = document.querySelectorAll("#article_list input[type='checkbox']");
                        indication = "确定要将本页条目都标记为已读吗？";
                    }

                    var ids = [];
                    for (i in sel) {
                        if (sel[i].value) {
                            ids.push(sel[i].value)
                        }
                    }

                    if (ids.length > 0 && window.confirm(indication)) {
                        MarkArticlesRead("ids", ids);
                    }

                } else if (tp == "feed") {
                    if (window.confirm("你确定要将“" + feedName + "”所含文章全部标记为已读吗？")) {
                        MarkArticlesRead("feedid", feedId);
                    }

                } else if (tp == "tag") {
                    if (window.confirm("你确定要将“" + tag + "”所含文章全部标记为已读吗？")) {
                        MarkArticlesRead("tag", tag);
                    }
                }
            };

        } else {
            if (data.error.errcode == 103) {
                QDoc.SetError("搜索指令语法错误");
            } else if (data.error.errcode == 302) {
                QDoc.SetError("没有找到可供阅读的文章");
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        }
    });

    $scope.markStarred = function(status, id) {
        QDoc.ClearError();

        var starred = function(status) {
            var articles = $scope.data.Articles;
            for (i in articles) {
                if (articles[i].item_id == id) {
                    articles[i].item_starred = status;
                    return;
                }
            }
        };

        starred(status);

        var post = { "status": status, "ids": [parseInt(id)]};
        $http.put(QReader.api.markArticlesStarred, post).success(function(data) {
            if (!data.success) {
                starred(!status);
                if (data.error.errcode == 303) {
                    QDoc.SetError("操作失败，文章不存在或已被删除。");
                } else {
                    QDoc.SetError(data.error.errmsg);
                }
            }
        }).error(function() {
            starred(!status);
        });
    };
});


QReader.app.controller("NavController", function($scope, $route, $location) {
    // 刷新
    $scope.reload = function() {
        $route.reload();
    };

    $scope.search = function() {
        var lastSearchQuery = localStorage.lastSearchQuery || QReader.defaultStorage.lastSearchQuery;
        var query = window.prompt("请输入搜索指令：", lastSearchQuery);
        if (query !== null && query != "") {
            localStorage.lastSearchQuery = query;
            $location.url("/articles/search/?q=" + encodeURIComponent(query));
            $route.reload();
        }
    };

    $scope.logout = function() {
        localStorage.authToken = "";
        window.location.href = "/login.html";
    };
});


QReader.app.controller("FeedListController", function($http, $scope, $route, $location, QDoc, MarkArticlesRead) {

    QDoc.Init("订阅管理");

    var checkSubscriptionInput = function() {
        if ($scope.url == "" || $scope.url == "http://") {
            return false;
        }
        return true;
    }

    var subscriptionRequestUrl = function(feedurl) {
        var url = QReader.api.subscription + "?url=" + encodeURIComponent(feedurl);
        return url;
    }

    // Check if a feed is subscribed.
    $scope.isSubscribed = function() {
        QDoc.ClearError();

        if (checkSubscriptionInput() == false) {
            return;
        }

        $http.get(subscriptionRequestUrl($scope.url)).success(function(data) {
            if (data.success) {
                if (data.result) {
                    alert("feed地址: " + $scope.url + " 已被订阅。");
                } else {
                    alert("feed地址: " + $scope.url + " 尚未被订阅。");
                }
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        });
    };

    $scope.subscribe = function() {
        QDoc.ClearError();

        if (checkSubscriptionInput() == false) {
            return;
        }

        QDoc.Loading(true);

        $http.post(subscriptionRequestUrl($scope.url)).success(function(data) {
            QDoc.Loading(false);
            if (data.success) {
                alert("'" + data.result.name + "' 订阅成功，新增文章：" + data.result.number + "篇");
                $route.reload();
            } else {
                if (data.error.errcode == 301) {
                    QDoc.SetError("此feed已被订阅，无法重复订阅");

                } else {
                    QDoc.SetError("订阅失败：" + data.error.errmsg);
                }
            }
        }).error(function() {
            QDoc.Loading(false);
        });
    };

    var initCols = function() {
        $scope.cols = {};
        $scope.cols.feed_id = "";
        $scope.cols.feed_name = ""
        $scope.cols.amount = "";
        $scope.cols.unread = "";
        $scope.cols.starred = "";
        $scope.cols.feed_last_fetch = "";
        $scope.cols.feed_last_failed = "";
    };

    var init = function() {

        var request_url = QReader.api.feedlist;

        $scope.reverse = localStorage.feedListOrderReverse || QReader.defaultStorage.feedListOrderReverse;
        $scope.reverse = ($scope.reverse == "true" || $scope.reverse == true) ? true : false;
        $scope.reverse = !$scope.reverse;
        $scope.colName = localStorage.feedListOrderBy || QReader.defaultStorage.feedListOrderBy;
        initCols();

        $http.get(request_url).success(function(data) {
            if (data.success) {
                $scope.data = data.result;

                // order by id asc
                $scope.sortIndicatorChagne($scope.colName);

            } else {
                QDoc.SetError(data.error.errmsg);
            }
        });
    };

    init();

    $scope.info = function(fid) {
        $location.url("/feed/info/" + fid)
        $route.reload();
    };

    $scope.update = function(id) {
        QDoc.ClearError();

        QDoc.Loading(true);

        var request_url = QReader.api.feed + id;
        $http.post(request_url).success(function(data) {
            QDoc.Loading(false);
            if (data.success) {
                init(); // reloading
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        }).error(function() {
            QDoc.Loading(false);
        });
    };

    $scope.markread = function(fid, name) {
        QDoc.ClearError();
        if (window.confirm("你确定要将 '" + name + "' 所包含的文章全部标记为已读吗？")) {
            MarkArticlesRead("feedid", fid)
        }
    };

    $scope.sortIndicatorChagne = function(colName) {
        $scope.colName = colName;
        $scope.reverse = !$scope.reverse;

        var icon = "fa fa-sort-up";
        if ($scope.reverse) {
            icon = "fa fa-sort-down";
        }

        initCols();
        $scope.cols[colName] = icon;

        localStorage.feedListOrderBy = $scope.colName;
        localStorage.feedListOrderReverse = $scope.reverse;
    };

    $scope.delete = function(fid, name) {
        QDoc.ClearError();
        if (window.confirm("你确定要将feed '" + name + "' 及所包含的全部文章都删除吗？")) {
            $http.delete(QReader.api.feed + fid).success(function(data) {
                if (data.success) {
                    init();
                } else {
                    if (data.error.errcode == 304) {
                        QDoc.SetError("'" + name + "'下包含部分加星文章，请先将星去除后再删除feed。");
                    } else {
                        QDoc.SetError(data.error.errmsg);
                    }
                }
            });
        }
    };
});


QReader.app.controller("FeedInfoController", function($http, $scope, $route, $routeParams, QDoc) {
    QDoc.ClearError();

    var getRequestUrl = function() {
        return QReader.api.feed + $routeParams.fid;
    };

    var update = function(post) {
        $http.put(getRequestUrl(), post).success(function(data) {
            if (data.success) {
                alert("保存成功");
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        });
    };

    $scope.cancel = function() {
        $route.reload();
    };

    $http.get(getRequestUrl()).success(function(data) {
        if (data.success) {
            var title = data.result.feed_name;
            QDoc.SetPageTitle("<a href='/#/articles/fid/" + data.result.feed_id + "' title='查看feed文章'>" + title + "</a>");
            QDoc.SetHtmlTitle(title + " - feed详情");
            $scope.data = data.result;
            if (data.result.Tags) {
                $scope.data.Tags = data.result.Tags.join(", ");
            }

            $scope.save = function() {
                var post = {}
                post.feed_alias         = $scope.data.feed_alias;
                post.feed_url           = $scope.data.feed_feed_url;
                post.feed_note          = $scope.data.feed_note;
                post.feed_max_keep      = parseInt($scope.data.feed_max_keep);
                post.feed_max_unread    = parseInt($scope.data.feed_max_unread);
                post.feed_interval      = parseInt($scope.data.feed_interval);

                if (post.feed_max_keep < 0 || post.feed_max_unread < 0) {
                    alert("最大已读保留数、最大未读保留数均不能小于0。")
                    return;
                }

                if ($scope.data.Tags) {
                    post.tags = $scope.data.Tags.split(",");
                } else {
                    post.tags = [];
                }
                update(post);
            };

        } else {
            if (data.error.errcode == 302) {
                QDoc.SetError("编号为" + $routeParams.fid + "的feed信息并没有找到");
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        }
    });
});


QReader.app.controller("ArticleController", function($http, $scope, $routeParams, QDoc) {
    QDoc.ClearError();
    QDoc.SetPageTitle("");

    var request_url = QReader.api.article + $routeParams.id;
    var article_title;

    $http.get(request_url).success(function(data) {
        if (data.success) {
            data.result.related = data.result.related.slice(0, 3);
            QDoc.SetHtmlTitle(data.result.article.item_title);
            $scope.data = data.result;
            article_title = data.result.article.item_title;
        } else {
            if (data.error.errcode == 302) {
                QDoc.SetError("编号为" + $routeParams.id + "的文章并没有找到");
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        }
    });

    $scope.markRead = function(markread) {
        QDoc.ClearError();

        var request_url = $routeParams.id;
        if (markread) {
            request_url = QReader.api.markArticleRead + request_url;
        } else {
            request_url = QReader.api.markArticleUnread + request_url;
        }

        $http.put(request_url).success(function(data) {
            if (data.success) {
                $scope.data.article.item_read = markread;

                if (markread == false) {
                    QDoc.SetHtmlTitle("（未读）" + article_title);
                } else {
                    QDoc.SetHtmlTitle(article_title);
                }

            } else {
                if (data.error.errcode == 303) {
                QDoc.SetError("编号为" + $routeParams.id + "的文章并没有找到");
                } else {
                    QDoc.SetError(data.error.errmsg);
                }
            }
        });
    };

    $scope.markStarred = function(status) {
        QDoc.ClearError();

        $scope.data.article.item_starred = status;

        var post = { "status": status, "ids": [parseInt($routeParams.id)]};
        $http.put(QReader.api.markArticlesStarred, post).success(function(data) {
            if (!data.success) {
                $scope.data.item_starred = !status;
                if (data.error.errcode == 303) {
                    QDoc.SetError("操作失败，文章不存在或已被删除。");
                } else {
                    QDoc.SetError(data.error.errmsg);
                }
            }
        }).error(function() {
            $scope.data.article.item_starred = !status;
        });
    };
});


QReader.app.controller("SettingsController", function($http, $scope, QDoc) {
    QDoc.Init("系统信息");

    var initSettings = function() {
        var settings = {};
        settings.page_item_number = parseInt(localStorage.pageItemNumber || QReader.defaultStorage.pageItemNumber);
        $scope.settings = settings;
    };

    $scope.save = function() {
        var post = $scope.settings;
        localStorage.pageItemNumber = post.page_item_number;
        alert("保存成功");
    };

    $scope.cancel = function() {
        initSettings();
    };

    $http.get(QReader.api.settings).success(function(data) {
        if (data.success) {
            $scope.status = "QReader api服务正常运行";
            $scope.data = data.result;
            $scope.token = QReader.ApiToken();
            initSettings();
        } else {
            $scope.status = "QReader api服务存在问题：" + data.error.errmsg;
        }
    }).error(function() {
        $scope.status = "QReader api服务异常";
    });

    $scope.shutdown = function() {
        if (window.confirm("确定要关闭 QReader 服务器吗？")) {
            var request_url = QReader.api.shutdown;
            $http.put(request_url);
        }
    };
});


QReader.app.controller("TagsListController", function($http, $scope, $location, QDoc){
    QDoc.Init("标签列表");

    var getData = function(getall) {

        var request_url = QReader.api.tagsList;
        if (getall) {
            request_url += "?getall";
        }

        $http.get(request_url).success(function(data) {
            if (data.success) {
                // sum unread number of a tag.
                var sum = {};
                for (item in data.result) {
                    for (i in data.result[item]) {
                        sum[item] = sum[item] || 0;
                        sum[item] += data.result[item][i].unread;
                    }
                }
                $scope.data = data.result;
                $scope.sum = sum;

                if (getall) {
                    angular.element(document.querySelectorAll("#tagslist input[type='checkbox']")[0]).prop("checked", true);
                }
            } else {
                QDoc.SetError(data.error.errmsg);
            }
        });
    };

    $scope.showall = function() {
        var getall = ($scope.getall == "true" || $scope.getall == true) ? true : false;
        localStorage.tagsListShowAll = getall;
        getData(getall);
    };

    var getall = localStorage.tagsListShowAll || QReader.defaultStorage.tagsListShowAll;
    getall = (getall == "true" || getall == true) ? true : false;
    getData(getall);
});


QReader.defaultStorage = {
    "pageItemNumber"        : 10,
    "tagsListShowAll"       : false,
    "feedListOrderBy"       : "feed_id",
    "feedListOrderReverse"  : false,
    "authToken"             : "",
    "lastSearchQuery"       : "",
};


window.onscroll = function(){
    var topDiv = document.getElementById("gotop");
    if(document.body.scrollTop >= 300) {
        topDiv.style.display = "inline";
    } else {
        topDiv.style.display = "none";
    }
};


