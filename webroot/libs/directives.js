
var appDirectives = angular.module('appDirectives', [])

appDirectives.directive('handlerEditor', [function() {
	return {
		restrict: 'A',
		require: '?ngModel',
		link: function(scope, elem, attrs, ctrl) {
			if(!ctrl) return;

			var editor;
			var langMap = {
			Language: function(handlerName) {
					switch(true) {
						case /\.sh$/.test(handlerName):
							return "sh";
							break;
						case /.*\.py$/.test(handlerName):
							return "python";
							break;
						case /\.pl$/.test(handlerName):
							return "perl";
							break;
						case /\.rb$/.test(handlerName):
							return "python";
							break;
						default:
							return null;
							break;
					}
				}
			};

			function initEditor() {
				editor = ace.edit(elem[0].id);
				editor.setTheme("ace/theme/ambiance");
			}

			function loadEditor() {

				scope.selectedHandler.language = langMap.Language(scope.selectedHandler.name);
				if (scope.selectedHandler.language !== null) {
            		editor.getSession().setMode("ace/mode/" + scope.selectedHandler.language);
					editor.setValue(scope.selectedHandler.content);
					editor.setReadOnly(true);
					editor.gotoLine(editor.session.getLength());
					$(elem).fadeIn();
				} else {
            		console.log("Could not determind handler language! Not loading editor!");
				}
			}

			function init() {
				if(elem.length > 1) {
					console.warn("can only initialize 1 editor per session");
				}
            	scope.$watch('editorStatus', function(newVal, oldVal) {
            		if(newVal !== oldVal) {
            			switch(newVal) {
            				case "init":
            					initEditor();
            					loadEditor();
            					break;
            				case "load":
            					loadEditor();
            					break;
            				case "reload":
            					initEditor();
            					loadEditor();
            					break;
            				default:
            					break;
            			}
            		}
            	});
			}

			init();
		}
	}
}]);