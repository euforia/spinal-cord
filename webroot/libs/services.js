
var appServices = angular.module('appServices', []);

appServices.factory('ConfigManager', ['$http', function($http) {

    var _config = null;

    function _fetchConfig(callback, fresh) {

        var dfd = $.Deferred();
        if (_config == null || fresh) {
            $http({
                url: "/conf/config.json",
                method: "GET",
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                }
            }).success(function(data, status, headers, config) {

                _config = data;
                dfd.resolve(data);
            }).error(function(data, status, headers, config) {

                console.error(data);
            });
        } else {
            dfd.resolve(_config);
        }
        if(callback) dfd.done(callback)
    }

    _fetchConfig();

    return {
        getConfig: function(callback, fresh) {
            _fetchConfig(callback, fresh);
        }
    };
}]);

appServices.factory('Authenticator', ['$window', '$http', function($window, $http) {
    var Authenticator = {
        login: function(creds) {
            // do actual auth here //
            if(creds.username === "guest" && creds.password === "guest") {
                $window.sessionStorage['credentials'] = JSON.stringify(creds);
                return true;
            }
            return false;
        },
        sessionIsAuthenticated: function() {
            if($window.sessionStorage['credentials']) {

                var creds = JSON.parse($window.sessionStorage['credentials']);
                if(creds.username && creds.username !== "" && creds.password && creds.password !== "") {
                    // do custom checking here
                    return true
                }
            }
            return false;
        }
    };

    return (Authenticator);
}]);

appServices.factory('SpinalCord', ['$http', 'ConfigManager', function($http, ConfigManager) {

    var API_URL = "/api/ns";

    function httpJsonCall(url, method, callback) {
        $http({
            url: url,
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
        }).success(function(data, status, headers, config) {

            if(callback) callback(data);
        }).error(function(data, status, headers, config) {

            console.error(data);
            if(callback) callback({error: data});
        });
    }

    var SpinalCord = {
        Namespaces: function(cb) {
            httpJsonCall(API_URL+"/", "GET", cb);
        },
        EventTypes: function(namespace, cb) {
            httpJsonCall(API_URL+"/"+namespace, "GET", cb);
        },
        EventTypeDetails: function(namespace, eventType, cb) {
            httpJsonCall(API_URL+"/"+namespace+"/"+eventType, "GET", cb);
        },
        HandlerContents: function(ns, etype, handlerName, cb) {
            httpJsonCall(API_URL+"/"+ns+"/"+etype+"/"+handlerName, "GET", cb);
        }
    };

    return (SpinalCord);
}]);


