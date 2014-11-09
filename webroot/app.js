
var app = angular.module('app', [
	'ngRoute',
	'appDirectives',
	'appFactories',
	'appControllers',
	'appServices'
]);

app.config(['$routeProvider',
	function($routeProvider) {
		$routeProvider.
			when('/login', {
				templateUrl: 'partials/login.html',
				controller: 'loginController'
			})
			.when('/ns', {
				templateUrl: 'partials/namespaces.html',
				controller: 'namespacesController'
			})
			.when('/ns/:Namespace', {
				templateUrl: 'partials/namespaceDetails.html',
				controller: 'namespaceDetailsController'
			})
			.when('/ns/:Namespace/:EventType', {
				templateUrl: 'partials/eventTypeDetails.html',
				controller: 'eventTypeDetailsController'
			})
			.when('/ns/:Namespace/:EventType/:Handler', {
				templateUrl: 'partials/eventTypeDetails.html',
				controller: 'eventTypeDetailsController'
			})
			.otherwise({
				redirectTo: '/login'
			});
	}
]);

app.filter('objectLength', function() {
	return function(obj) {
    	return Object.keys(obj).length;
	};
})
.filter('formattedSha1', function() {
	return function(sha1str) {
		if (sha1str == "") return "";
		else return "sha1 : "+sha1str;
	};
})
.filter('formattedHandlerName', function() {
	return function(name) {
		if (name == "") return "";
		else return "name : "+name;
	};
})
.filter('formattedLanguage', function() {
	return function(name) {
		if (name == "") return "";
		else return "lang : "+name;
	};
});
