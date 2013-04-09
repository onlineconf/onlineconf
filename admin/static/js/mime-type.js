var mimeType;

$(function() {
    mimeType = {
        "application/x-null": { title: "Null", edit: editNull, preview: previewNull, view: viewNull, validate: validateNone },
        "text/plain": { title: "Text", edit: editText, preview: previewText, view: viewText, validate: validateNone },
        "application/json": { title: "JSON", edit: editText, preview: previewText, view: viewText, validate: validateJSON },
        "application/x-yaml": { title: "YAML", edit: editText, preview: previewText, view: viewText, validate: validateNone },
        "application/x-symlink": { title: "Symbolic link", edit: editSymlink, preview: previewSymlink, view: viewSymlink, validate: validateNone },
        "application/x-case": { title: "Case", edit: editCase, preview: previewCase, view: viewCase, validate: validateNone },
    };

    function previewNull (data) {
        return $('<span class="null"/>');
    }

    function previewText (data) {
        return $('<span/>').text(data);
    }

    function previewSymlink (data) {
        return $('<a class="symlink" onclick="event.stopPropagation()"/>').attr('href', '#' + data).text(data);
    }

    function previewCase (data) {
        var result = $('<span class="case-preview"/>');
        try {
            var cases = $.parseJSON(data);
            $.each(cases, function (id, value) {
                var $key = $('<span class="case-key"/>');
                var def = false;
                if ('datacenter' in value) {
                    $key.text(value.datacenter);
                    resolveList('datacenter', $key);
                } else if ('group' in value) {
                    $key.text(value.group);
                    resolveList('group', $key);
                } else if ('server' in value) {
                    $key.text(value.server);
                } else {
                    def = true;
                }
                var $span = $('<span/>');
                if (!def) $span.append($key).append(': ');
                var $value = mimeType[value.mime].preview(value.value);
                if (value.mime == 'application/x-case') $value.prepend('{ ').append(' }');
                $span.append($('<span class="case-value"/>').append($value)).appendTo(result);
                if (id < cases.length - 1) result.append('; ');
            });
        } catch (e) {
        }
        return result;
    }

    function viewNull (span, mime, data) {
        span.empty().append('<span class="null">NULL</span>');
    }

    function viewText (span, mime, data) {
        span.empty();
        var viewer = $('<span class="view-text"/>').appendTo(span);
        var cm = CodeMirror(viewer[0], { readOnly: true, mode: mime, tabindex: -1 });
        cm.setValue(data != null ? data : '');
        createCodeMirrorToolbar(cm).prependTo(span);
    }

    function viewSymlink (span, mime, data) {
        span.empty().append($('<a class="symlink"/>').attr('href', '#' + data).text(data));
    }

    function viewCase (span, mime, data) {
        span.empty();
        var cspan = $('<div class="case-view"/>').appendTo(span);
        try {
            var cases = $.parseJSON(data);
            $.each(cases, function (id, value) {
                var $key = $('<span class="case-key"/>');
                var kspan = $('<span class="case-key-block"/>')
                    .append($key)
                    .append(': ');
                var vspan = $('<span class="case-value case-value-block"/>');
                if ('datacenter' in value) {
                    $key.text(value.datacenter);
                    resolveList('datacenter', $key);
                } else if ('group' in value) {
                    $key.text(value.group);
                    resolveList('group', $key);
                } else if ('server' in value) {
                    $key.text(value.server);
                } else {
                    $key.text('default').addClass('case-key-default');
                }
                $('<div/>')
                    .append(kspan)
                    .append(vspan)
                    .appendTo(cspan);
                mimeType[value.mime].view(vspan, value.mime, value.value);
            });
        } catch (e) {
        }
    }

    function editNull (span, mime, data) {
        span.empty().parent('div').hide();
        return function () { return null };
    }

    function editText (span, mime, data) {
        span.empty().parent('div').show();
        var editor = $('<span/>').appendTo(span);
        var cm = CodeMirror(editor[0], { mode: mime, tabMode: mime == 'text/plain' ? 'default' : 'shift' });
        cm.setValue(data != null ? data : '');
        editor.find('.CodeMirror').addClass('ui-widget-content ui-corner-all');
        createCodeMirrorToolbar(cm).prependTo(span);
        return function () { return cm.getValue() };
    }

    function editSymlink (span, mime, data) {
        span.empty().parent('div').show();
        var text = $('<input/>').addClass('input-width-fill').val(data).appendTo(span)
            .wrap($('<div class="ui-widget-content ui-corner-all input-width-fill-wrapper"/>'))
            .autocompletePath(true);
        return function () { return text.val() };
    }

    function editCase (span, mime, data) {
        span.empty().parent('div').show();
        var defaultExists = false;
        var container = $('<div/>').appendTo(span);
        var changeKey = function () {
            var $kvSpan = $(this).parent('span.label').find('~ span.value');
            var val = $kvSpan.find('select, input').val();
            $kvSpan.empty();
            if ($(this).val() == 'datacenter') {
                var $dc = $('<select name="datacenter"/>').appendTo($kvSpan);
                $.get('/config/onlineconf/datacenter?symlink=resolve', {}, function (data) {
                    $dc.empty();
                    $.each(data.children, function (id, param) {
                        $('<option/>').val(param.name).text(param.summary || param.name).appendTo($dc);
                    });
                    $dc.val(val);
                });
            } else if ($(this).val() == 'group') {
                var $dc = $('<select name="group"/>').appendTo($kvSpan);
                $.get('/config/onlineconf/group?symlink=resolve', {}, function (data) {
                    $dc.empty();
                    $.each(data.children, function (id, param) {
                        if (param.name == 'priority') return;
                        $('<option/>').val(param.name).text(param.summary || param.name).appendTo($dc);
                    });
                    $dc.val(val);
                });
            } else if ($(this).val() == 'server') {
                $('<input name="server"/>')
                    .val(val)
                    .appendTo($kvSpan)
                    .addClass('input-width-fill')
                    .wrap($('<div class="ui-widget-content ui-corner-all input-width-fill-wrapper"/>'));
            }
            $(this).parents('.value-case:first').siblings().find('> div > span > select > option[value=default]').toggle($(this).val() != 'default');
        };
        var addFunc = function (kt, kv, m, v) {
            var div = $('<div class="value-case nice-form ui-corner-all"/>').appendTo(container);
            var keyType = $('<select name="key"/>')
                .append('<option value="default">Default</option>')
                .append('<option value="server">Сервер</option>')
                .append('<option value="group">Группа</option>')
                .append('<option value="datacenter">Датацентр</option>')
                .change(changeKey)
                .val(kt);
            if (span.find('> div > div > div > span > select > option[value=default]:selected').length) keyType.find('option[value=default]').hide();
            $('<div/>')
                .append($('<span class="label"/>').append(keyType))
                .append($('<span class="value"/>').append($('<input/>').val(kv)))
                .appendTo(div)
                .find('span.label select').change().end();
            var value = $('<span class="value getter"/>').data('getter', function () { return v });
            var type = $('<select name="mime"/>').change(function () {
                value.data('getter', mimeType[$(this).val()].edit(value, $(this).val(), value.data('getter')(true)));
                value.data('mime', $(this).val());
            });
            $.each(mimeType, function (k, v) { $('<option/>').prop('value', k).text(v.title).appendTo(type) });
            var node = span.parents('.ui-dialog-content').data('node');
            if (node && node.data('node').num_children != 0) {
                type.find('option[value="application/x-symlink"]').hide();
            }
            type.val(m);
            $('<div/>')
                .append('<span class="label"><span>Тип:</span></span>')
                .append($('<span class="value"/>').append(type))
                .appendTo(div);
            $('<div/>')
                .append('<span class="label"><span>Значение:</span></span>')
                .append(value)
                .appendTo(div);
            $('<div/>')
                .append($('<button type="button">Удалить</button>').click(function () { div.remove() }))
                .appendTo(div);
            type.change();
            return false;
        };
        $('<div/>')
            .append($('<button type="button">Добавить</button>').click(function () { addFunc('server') }))
            .appendTo(span);
        var ok = false;
        try {
            var cases = $.parseJSON(data);
            if ($.isArray(cases)) {
                $.each(cases, function (id, value) {
                    if (typeof value === "object" && "mime" in value && "value" in value) {
                        var key = "server" in value ? 'server'
                            : "group" in value ? 'group'
                            : "datacenter" in value ? 'datacenter'
                            : "default";
                        ok = true;
                        addFunc(key, value[key], value.mime, value.value);
                    }
                });
            }
        } catch (e) {
        }
        if (!ok) {
            var mime = span.data('mime');
            if (mime == null) mime = 'application/x-null';
            addFunc('default', '', mime, data);
        }
        return function (change) {
            var cases = [];
            container.find('> .value-case').each(function () {
                var data = {
                    mime: $(this).find('select[name=mime]').val(),
                    value: $(this).find('span.value.getter').data('getter')()
                };
                var key = $(this).find('select[name=key]').val();
                if (key == 'datacenter') {
                    data.datacenter = $(this).find('select[name=datacenter]').val();
                } else if (key == 'group') {
                    data.group = $(this).find('select[name=group]').val();
                } else if (key == 'server') {
                    data.server = $(this).find('input[name=server]').val();
                }
                if (key == 'default') cases.unshift(data);
                else cases.push(data);
            });
            if (change) {
                var def = $.map(cases, function (v) { return !("server" in v || "group" in v || "datacenter" in v) ? v.value : null });
                if (def.length) return def[0];
                var star = $.map(cases, function (v) { return v.server == "*" ? v.value : null });
                return star[0];
            } else {
                return $.toJSON(cases);
            }
        };
    }

    function validateNone (span, data) {
        span.validationEngine('hidePrompt');
        return true;
    }

    function validateJSON (span, data) {
        try {
            $.parseJSON(data);
        } catch (e) {
            span.validationEngine('showPrompt', '* Некорректный JSON<br/>* ' + e, false, false, true);
            return false;
        }
        span.validationEngine('hidePrompt');
        return true;
    }

    var resolveStash = {};
    function resolveList (type, span) {
        if (!(type in resolveStash)) {
            resolveStash[type] = {
                time: new Date(0),
                queue: [],
                map: {}
            };
        }
        var stash = resolveStash[type];
        if (stash.time < new Date - new Date(60000)) {
            if (stash.queue.length == 0) {
                $.get('/config/onlineconf/' + type + '?symlink=resolve', {}, function (data) {
                    stash.map = {};
                    $.each(data.children, function (id, param) {
                        stash.map[param.name] = param.summary;
                    });
                    $.each(stash.queue, function (id, elem) {
                        var summary = stash.map[elem.name]
                        if (summary) elem.span.text(summary);
                    });
                    stash.queue = [];
                    stash.time = new Date;
                });
            }
            stash.queue.push({ span: span, name: span.text() });
        }
        var summary = stash.map[span.text()]
        if (summary) span.text(summary);
    }

    function createCodeMirrorToolbar (cm) {
        var rand = Math.random();
        var fullscreen = $('<input type="checkbox" id="cm-fullscreen-' + rand + '"/>');
        var wrap = $('<input type="checkbox" id="cm-wrap-' + rand + '"/>');
        var toolbar = $('<span class="cm-toolbar"/>')
            .append('<label for="cm-fullscreen-' + rand + '">Fullscreen</label>')
            .append(fullscreen)
            .append('<label for="cm-wrap-' + rand + '">Переносить</label>')
            .append(wrap);
        fullscreen.button({ text: false, label: 'Fullscreen', icons: { primary: 'ui-icon-arrow-4-diag' } })
            .click(function (e) {
                toolbar.stop().fadeTo(0, 1);
                if ($(this).prop('checked')) setFullScreen(cm, true);
                cm.focus();
            });
        wrap.button({ text: false, icons: { primary: 'ui-icon-arrowreturn-1-w' } })
            .click(function (e) {
                toolbar.stop().fadeTo(0, 1);
                var value = $(this).prop('checked') ? true : false;
                cm.setOption('lineWrapping', value);
                cm.setOption('lineNumbers', value);
                cm.focus();
            });
        toolbar.buttonset();
        cm.setOption('extraKeys', { "Esc": function (cm) { setFullScreen(cm, false); fullscreen.prop('checked', false).button('refresh') } });
        cm.on('focus', function (cm) { toolbar.fadeIn() });
        cm.on('blur', function (cm) { toolbar.fadeOut() });
        return toolbar;
    }

    function setFullScreen(cm, full) {
        var $wrap = $(cm.getWrapperElement());
        if ($wrap.hasClass('CodeMirror-fullscreen') == full) return;
        if (full) {
            document.documentElement.style.overflow = "hidden";
            $wrap.addClass('CodeMirror-fullscreen').height($(window).height());
        } else {
            document.documentElement.style.overflow = "";
            $wrap.removeClass('CodeMirror-fullscreen').height('');
        }
        cm.refresh();
    }

    CodeMirror.on(window, "resize", function() {
        $('.CodeMirror-fullscreen').height($(window).height());
    });
});
