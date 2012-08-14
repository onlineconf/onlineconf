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
                    resolveDatacenter($key);
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
        var cm = CodeMirror(span[0], { readOnly: 'nocursor', mode: mime });
        cm.setValue(data != null ? data : '');
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
                    resolveDatacenter($key);
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
        var cm = CodeMirror(span[0], { mode: mime, tabMode: mime == 'text/plain' ? 'default' : 'shift' });
        cm.setValue(data != null ? data : '');
        span.find('.CodeMirror').addClass('ui-widget-content ui-corner-all');
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
                } else if (key == 'server') {
                    data.server = $(this).find('input[name=server]').val();
                }
                if (key == 'default') cases.unshift(data);
                else cases.push(data);
            });
            if (change) {
                var def = $.map(cases, function (v) { return !("server" in v || "datacenter" in v) ? v.value : null });
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

    var resolveDcQueue = [];
    var resolveDcTime = new Date(0);
    var resolveDcMap = {};
    function resolveDatacenter (span) {
        if (resolveDcTime < new Date - new Date(60000)) {
            if (resolveDcQueue.length == 0) {
                $.get('/config/onlineconf/datacenter?symlink=resolve', {}, function (data) {
                    resolveDcMap = {};
                    $.each(data.children, function (id, param) {
                        resolveDcMap[param.name] = param.summary;
                    });
                    $.each(resolveDcQueue, function (id, elem) {
                        var summary = resolveDcMap[elem.name]
                        if (summary) elem.span.text(summary);
                    });
                    resolveDcQueue = [];
                    resolveDcTime = new Date;
                });
            }
            resolveDcQueue.push({ span: span, name: span.text() });
        }
        var summary = resolveDcMap[span.text()]
        if (summary) span.text(summary);
    }
});
