
var appFactories = angular.module("appFactories", []);

appFactories.factory("AccessManager", [
    '$location', 'Authenticator',
    function($location, Authenticator) {

    var AccessManager = function(ctrlPath) {

        var t = this;
        t.redirectTo = ctrlPath;

        function initialize() {
            if(!Authenticator.sessionIsAuthenticated())
                $location.url("/login?redirect="+t.redirectTo);
        }

        initialize();
    };

    return (AccessManager);
}]);