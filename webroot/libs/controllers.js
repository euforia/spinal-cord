
var appControllers = angular.module('appControllers', []);

appControllers.controller('defaultController', [ '$window', '$location', '$scope',
	function($window, $location, $scope) {

		$scope.pageHeaderHtml = "/partials/page-header.html";

		$scope.logOut = function() {

	        console.log("De-authing...");
	        var sStor = $window.sessionStorage;
	        if(sStor['credentials']) {
	            delete sStor['credentials'];
	        }

	        /*
	        var lStor = $window.localStorage;
	        for(var k in lStor) {
	            if(/^token\-/.test(k)) delete lStor[k];
	        }
			*/
	        $location.url("/login");
	    }
	}
]);

appControllers.controller('loginController', [
	'$scope', '$window', '$routeParams', '$location', 'Authenticator',
	function($scope, $window, $routeParams, $location, Authenticator) {

		$scope.credentials = { username: "", password: "" };

		var defaultPage = "/ns";

		$scope.attemptLogin = function() {

			if(Authenticator.login($scope.credentials)) {

				if($routeParams.redirect) $location.url($routeParams.redirect);
				else $location.url(defaultPage);
			} else {

				$("#login-window-header").html("<span>Auth failed!</span>");
			}
		}

		function _initialize() {
			if($window.sessionStorage['credentials']) {

				var creds = JSON.parse($window.sessionStorage['credentials']);
				if(creds.username && creds.username !== "" && creds.password && creds.password !== "") {

					$scope.credentials = creds;
					$scope.attemptLogin();
				}
			}
		}

		_initialize();
	}
]);

appControllers.controller('namespacesController', [ '$window', '$location', '$scope', 'SpinalCord', 'AccessManager',
	function($window, $location, $scope, SpinalCord, AccessManager) {

		var accessMgr = new AccessManager("/ns");

		$scope.pageHeaderHtml = "/partials/page-header.html";

		$scope.Namespaces = [];
		$scope.namespaceSearch = "";

		SpinalCord.Namespaces(function(namespaces) {
			$scope.Namespaces = namespaces;
		});
	}
]);

appControllers.controller('namespaceDetailsController', [ '$window', '$location', '$routeParams', '$scope', 'SpinalCord', 'AccessManager',
	function($window, $location, $routeParams, $scope, SpinalCord, AccessManager) {

		var accessMgr = new AccessManager("/ns/"+$routeParams.Namespace);

		$scope.pageHeaderHtml = "/partials/page-header.html";

		$scope.Namespace = $routeParams.Namespace;
		$scope.EventTypes = [];
		$scope.eventTypeSearch = "";

		SpinalCord.EventTypes($routeParams.Namespace, function(eventTypes) {
			$scope.EventTypes = eventTypes;
		});
	}
]);

appControllers.controller('eventTypeDetailsController', [ '$window', '$location', '$routeParams', '$scope', 'SpinalCord', 'AccessManager',
	function($window, $location, $routeParams, $scope, SpinalCord, AccessManager) {

		var accessMgr;
		if($routeParams.Handler && $routeParams.Handler !== "")
			accessMgr = new AccessManager("/ns/"+$routeParams.Namespace+"/"+$routeParams.EventType+"/"+$routeParams.Handler);
		else
			accessMgr = new AccessManager("/ns/"+$routeParams.Namespace+"/"+$routeParams.EventType);

		$scope.pageHeaderHtml = "/partials/page-header.html";

		$scope.Namespace = $routeParams.Namespace;
		$scope.EventType = $routeParams.EventType;
		$scope.Details = {};

		$scope.handlerSearch = "";
		$scope.editorStatus = "";

		$scope.selectedHandler = {
			language: "",
			content: "",
			path: "",
			sha1: "",
			name: ""
		};

		function fetchHandlerData(handler) {
			SpinalCord.HandlerContents($scope.Namespace, $scope.EventType, handler.name,
				function(data) {
					$scope.selectedHandler = {
						content:  data.data,
						sha1:     data.sha1,
						name:     handler.name,
						path:     handler.path
					};
					handler.sha1 = data.sha1;
					if ($scope.editorStatus === "") $scope.editorStatus = "init";
					else $scope.editorStatus = "load";
				}
			);
		}

		SpinalCord.EventTypeDetails($scope.Namespace, $scope.EventType,
			function(rslt) {
				$scope.Details = rslt;

				if($routeParams.Handler && $routeParams.Handler !== "") {
					fetchHandlerData({
						name: $routeParams.Handler,
						path: $scope.Namespace+"/"+$scope.EventType+"/"+$routeParams.Handler
					});
				}
			}
		);

		$scope.fetchHandlerData = function(handler) {
			$location.path("/ns/"+$scope.Namespace+"/"+$scope.EventType+"/"+handler.name);
		}
	}
]);