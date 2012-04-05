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
        return '';
    }

    function previewText (data) {
        return data;
    }

    function previewSymlink (data) {
        return $('<a class="symlink" onclick="event.stopPropagation()"/>').attr('href', '#' + data).text(data);
    }

    function previewCase (data) {
        var result = [];
        try {
            var cases = $.parseJSON(data);
            $.each(cases, function (id, value) { result.push(value.server + ': ' + value.value) });
        } catch (e) {
        }
        return result.join('; ');
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
        span.empty().append($('<a/>').attr('href', '#' + data).text(data));
    }

    function viewCase (span, mime, data) {
        span.empty();
        try {
            var cases = $.parseJSON(data);
            $.each(cases, function (id, value) {
                var vspan = $('<span class="value"/>');
                $('<div class="nice-form"/>')
                    .append($('<span class="label"/>').text(value.server))
                    .append(vspan)
                    .appendTo(span);
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
            .autocompletePath();
        return function () { return text.val() };
    }

    function editCase (span, mime, data) {
        span.empty().parent('div').show();
        var container = $('<div/>').appendTo(span);
        var addFunc = function (s, m, v) {
            var div = $('<div class="value-case nice-form ui-corner-all"/>').appendTo(container);
            var server = $('<input name="server"/>').val(s);
            $('<div/>')
                .append('<span class="label">Сервер:</span>')
                .append($('<span class="value"/>').append(server))
                .appendTo(div);
            var value = $('<span class="value getter"/>').data('getter', function () { return v });
            var type = $('<select name="mime"/>').change(function () {
                value.data('getter', mimeType[$(this).val()].edit(value, $(this).val(), value.data('getter')(true)));
                value.data('mime', $(this).val());
            });
            $.each(mimeType, function (k, v) { $('<option/>').prop('value', k).text(v.title).appendTo(type) });
            type.find('option[value="application/x-symlink"]').toggle(span.parents('.ui-dialog-content').data('node').data('node').num_children == 0);
            type.val(m);
            $('<div/>')
                .append('<span class="label">Тип:</span>')
                .append($('<span class="value"/>').append(type))
                .appendTo(div);
            $('<div/>')
                .append('<span class="label">Значение:</span>')
                .append(value)
                .appendTo(div);
            $('<div/>')
                .append($('<button type="button">Удалить</button>').click(function () { div.remove() }))
                .appendTo(div);
            server.addClass('input-width-fill').wrap($('<div class="ui-widget-content ui-corner-all input-width-fill-wrapper"/>'));
            type.change();
            return false;
        };
        $('<div/>')
            .append($('<button type="button">Добавить</button>').click(function () { addFunc() }))
            .appendTo(span);
        var ok = false;
        try {
            var cases = $.parseJSON(data);
            if ($.isArray(cases)) {
                $.each(cases, function (id, value) {
                    if (typeof value === "object" && "server" in value && "mime" in value && "value" in value) {
                        ok = true;
                        addFunc(value.server, value.mime, value.value);
                    }
                });
            }
        } catch (e) {
        }
        if (!ok) {
            var mime = span.data('mime');
            if (mime == null) mime = 'application/x-null';
            addFunc('*', mime, data);
        }
        return function (change) {
            var cases = [];
            container.find('.value-case').each(function () {
                cases.push({
                    server: $(this).find('input[name=server]').val(),
                    mime: $(this).find('select[name=mime]').val(),
                    value: $(this).find('span.value.getter').data('getter')()
                });
            });
            if (change) {
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
});
