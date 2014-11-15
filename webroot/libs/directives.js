
var appDirectives = angular.module('appDirectives', [])

appDirectives.directive('handlerEditor', [function() {
	return {
		restrict: 'A',
		require: '?ngModel',
		link: function(scope, elem, attrs, ctrl) {
			if(!ctrl) return;

			var editor;
            var _aceScroller, _aceGutter;
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

            function setReadOnly(val) {
                editor.setReadOnly(val);
                //aceScroller = $(elem).find('.ace_scroller');
                //aceGutter = $(elem).find('.ace_gutter');
                if(val) {
                    _aceScroller.addClass('readonly');
                    _aceGutter.addClass('readonly');
                } else {
                    _aceScroller.removeClass('readonly');
                    _aceGutter.removeClass('readonly');
                }
            }

			function initEditor() {
				editor = ace.edit(elem[0].id);
				editor.setTheme("ace/theme/ambiance");

                editor.addEventListener("blur", function(obj) {
                    scope.$apply(function() {
                        console.info("Write back");
                        scope.selectedHandler.content = editor.getValue();
                    });
                });

                _aceScroller = $(elem).find('.ace_scroller');
                _aceGutter = $(elem).find('.ace_gutter');
			}

            function newBlankEditor() {

                editor.setValue(ctrl.$modelValue);
                editor.gotoLine(editor.session.getLength());
                setTimeout(function() { setReadOnly(false); }, 400);
            }
			function loadEditor(blkEditor) {
                //if !(blkEditor)
				scope.selectedHandler.language = langMap.Language(scope.selectedHandler.name);
				if (scope.selectedHandler.language !== null) {

                    editor.setValue(ctrl.$modelValue);
					editor.gotoLine(editor.session.getLength());
                    setTimeout(function() { setReadOnly(true); }, 400);
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
                            case "init-blank":
                                initEditor();
                                newBlankEditor();
                                break;
            				default:
            					break;
            			}
            		}
            	});
                scope.$watch('selectedHandler.language', function(newVal, oldVal) {
                    if(newVal && newVal !== null && newVal !== "")

                        if(editor) editor.getSession().setMode("ace/mode/" + newVal);
                });
                scope.$watch('editing', function(newVal, oldVal) {
                    if(newVal !== null) {
                        if(newVal === true) {
                            //if(editor) editor.setReadOnly(false);
                            if(editor) setReadOnly(false);
                        } else {
                            //if(editor) editor.setReadOnly(true);
                            if(editor) setReadOnly(true);
                        }
                    }
                });
			}

			init();
		}
	}
}]);