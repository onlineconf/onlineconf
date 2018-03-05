/*global $, CodeMirror, document, window */
var mimeType;

$(function () {
    'use strict';
    function setFullScreen(cm, full) {
        var $wrap = $(cm.getWrapperElement());
        if ($wrap.hasClass('CodeMirror-fullscreen') === full) {
            return;
        }
        if (full) {
            document.documentElement.style.overflow = "hidden";
            $wrap.addClass('CodeMirror-fullscreen').height($(window).height());
        } else {
            document.documentElement.style.overflow = "";
            $wrap.removeClass('CodeMirror-fullscreen').height('');
        }
        cm.refresh();
    }

    function createCodeMirrorToolbar(cm) {
        var rand = Math.random(),
            fullscreen = $('<input type="checkbox" id="cm-fullscreen-' + rand + '"/>'),
            wrap = $('<input type="checkbox" id="cm-wrap-' + rand + '"/>'),
            toolbar = $('<span class="cm-toolbar"/>')
                .append('<label for="cm-fullscreen-' + rand + '">Fullscreen</label>')
                .append(fullscreen)
                .append('<label for="cm-wrap-' + rand + '">Переносить</label>')
                .append(wrap);

        fullscreen.button({ text: false, label: 'Fullscreen', icons: { primary: 'ui-icon-arrow-4-diag' } })
            .click(function () {
                toolbar.stop().fadeTo(0, 1);
                if ($(this).prop('checked')) {
                    setFullScreen(cm, true);
                }
                cm.focus();
            });
        wrap.button({ text: false, icons: { primary: 'ui-icon-arrowreturn-1-w' } })
            .click(function () {
                toolbar.stop().fadeTo(0, 1);
                var value = $(this).prop('checked') ? true : false;
                cm.setOption('lineWrapping', value);
                cm.setOption('lineNumbers', value);
                cm.focus();
            });
        toolbar.buttonset();
        cm.setOption('extraKeys', {
            "Esc": function (cm) {
                setFullScreen(cm, false);
                fullscreen.prop('checked', false).button('refresh');
            }
        });
        cm.on('focus', function () { toolbar.fadeIn(); });
        cm.on('blur', function () { toolbar.fadeOut(); });
        return toolbar;
    }

    var cache = {};
    function getCached(url, callback) {
        if (url in cache) {
            if (cache[url].expire < Date.now()) {
                delete cache[url];
            } else if ("data" in cache[url]) {
                callback(cache[url].data);
                return;
            } else {
                cache[url].callbacks.push(callback);
                return;
            }
        }
        var c = {
            expire: Date.now() + 60000,
            callbacks: [callback]
        };
        cache[url] = c;
        $.get(url, function (data) {
            c.data = data;
            c.callbacks.forEach(function (cb) { cb(data) });
        });
    }

    function resolveList(type, span) {
        getCached('/config/onlineconf/' + type + '?symlink=resolve', function (data) {
            var map = {};
            $.each(data.children, function (id, param) {
                map[param.name] = param.summary;
            });
            var summary = map[span.text()];
            if (summary) {
                span.text(summary);
            }
        });
    }

    /* new server */
    function parseServerData(data, mime) {
        var dataHash = {};

        if (mime === 'application/x-list') {
            dataHash = data.split(',');
        } else if (mime === 'application/x-server') {
            data = data.split(',');
            data.forEach(function (item) {
                item = item.split(':');
                if (dataHash[item[0]]) {
                    dataHash[item[0]].push(item[1]);
                } else {
                    dataHash[item[0]] = [item[1]];
                }
            });
        } else if (mime === 'application/x-server2') {
            data = data.split(';');
            data.forEach(function (item) {
                item = item.split(':');
                dataHash[item[0]] = item[1].split(',');
            });
        }

        return dataHash;
    }
    function editServer(span, mime, data) {
        var addcontainer = $('<div class="add-server-container" />'),
            ip = $('<span class="add-server-host-container" />').append(
                $('<input type="text" data-type="ip" placeholder="127.0.0.1" class="add-server-host" />'),
                $('<span class="add-server-host-delete-button ui-button-icon-primary ui-icon ui-icon-closethick" title="удалить хост" />')
            ),
            port = $('<span class="add-server-port-container" />').append(
                $('<input type="text" data-type="port" placeholder="8080" class="add-server-port" />'),
                $('<span class="add-server-port-delete-button ui-button-icon-primary ui-icon ui-icon-closethick" title="удалить порт" />')
            ),
            addServerButton = $('<div class="add-server-host-button ui-button ui-widget ui-state-default ui-button-text-only ui-corner-all">Добавить значение</div>'),
            addPortButton = $('<span class="add-server-port-button ui-button ui-widget ui-state-default ui-button-text-only ui-corner-all">Добавить порт</span>'),
            onAddServerCallback = function (event) {
                var target = $(event.target),
                    e,
                    container = $('<div data-type="server" class="add-server-row" />');

                if (target.hasClass('simple')) {
                    e = ip.clone();
                    e.find('[data-type="ip"]').addClass('add-server-simple');
                    container.append(e);
                } else {
                    container.append(ip.clone())
                        .append(port.clone())
                        .append(addPortButton.clone());
                }

//                container.find('input[data-type="ip"]').mask('099.099.099.099');
                target.before(container);
            },
            onAddPortCallback = function (event) {
                var target = $(event.target),
                    portClone = port.clone();

                target.before(portClone);
                portClone.find('[data-type="port"]').focus();
            },
            onDeletePortCallback = function (event) {
                var target = $(event.target),
                    parent = target.parent();

                parent.remove();
            },
            onDeleteHostCallback = function (event) {
                var target = $(event.target),
                    parent = target.parents('[data-type="server"]');

                if (window.confirm('А туда ли ты нажал, пацанчик?')) {
                    parent.remove();
                }
            },
            dataHash = {};

        if (mime === 'application/x-list') {
            addServerButton.addClass('simple');
        }
        if (data) {
            dataHash = parseServerData(data, mime);
            $.each(dataHash, function (key, value) {
                var node = $('<div data-type="server" class="add-server-row" />'),
                    ipClone;

                ipClone = ip.clone();
                if (mime === 'application/x-list') {
                    ipClone.find('[data-type="ip"]').val(value).addClass('add-server-simple');
                    node.append(ipClone);
                } else {
                    ipClone.find('[data-type="ip"]').val(key);
                    node.append(ipClone);
                    value.forEach(function (item) {
                        var portClone = port.clone();
                        portClone.find('[data-type="port"]').val(item);
                        node.append(portClone);
                    });
                    node.append(addPortButton.clone());
                }
                addcontainer.append(node);
            });
            addcontainer.append(addServerButton);
        } else {
            if (mime === 'application/x-list') {
                addcontainer.append(
                    $('<div data-type="server" class="add-server-row" />')
                        .append(ip.clone()),
                    addServerButton
                ).find('[data-type="ip"]').addClass('add-server-simple');
            } else {
                addcontainer.append(
                    $('<div data-type="server" class="add-server-row" />')
                        .append(ip.clone())
                        .append(port.clone())
                        .append(addPortButton),
                    addServerButton
                );
            }
        }
        addcontainer.on('click', '.add-server-port-button', onAddPortCallback);
        addcontainer.on('click', '.add-server-host-button', onAddServerCallback);
        addcontainer.on('click', '.add-server-port-delete-button', onDeletePortCallback);
        addcontainer.on('click', '.add-server-host-delete-button', onDeleteHostCallback);
//        addcontainer.find('input[data-type="ip"]').mask('099.099.099.099');

        span.html(addcontainer);
        span.parent().css('display', 'block');

        return function () {
            var result = [],
                nodes = span.find('[data-type="server"]'),
                host,
                ports,
                portValue,
                j,
                splitter = ',',
                i;

            for (i = 0; i < nodes.length; i += 1) {
                host = $(nodes[i]).find('[data-type="ip"]').val();
                ports = $(nodes[i]).find('[data-type="port"]');
                if (mime === 'application/x-list') {
                    result.push(host);
                } else if (mime === 'application/x-server') {
                    for (j = 0; j < ports.length; j += 1) {
                        portValue = $(ports[j]).val();
                        if (portValue && portValue !== '') {
                            result.push(host + ':' + portValue);
                        } else {
                            result.push(host);
                        }
                    }
                } else if (mime === 'application/x-server2') {
                    portValue = [];
                    for (j = 0; j < ports.length; j += 1) {
                        portValue.push($(ports[j]).val());
                    }
                    result.push(host + ':' + portValue.join(','));
                    splitter = ';';
                }
            }
            addcontainer.off('click', '.add-server-port-button', onAddPortCallback);
            addcontainer.off('click', '.add-server-host-button', onAddServerCallback);
            addcontainer.off('click', '.add-server-port-delete-button', onDeletePortCallback);
            addcontainer.off('click', '.add-server-host-delete-button', onDeleteHostCallback);
            addcontainer.remove();
            return result.join(splitter);
        };
    }
    function viewServer(span, mime, data) {
        var view = $('<div class="view-server" />'),
            dataHash = {},
            ip = $('<span class="view-server-host" />'),
            port = $('<span class="view-server-port" />');

        view.append('<div>mime type: ' + mime + '</div>');
        if (data) {
            dataHash = parseServerData(data, mime);
            $.each(dataHash, function (key, value) {
                var node = $('<div class="add-server-row" />'),
                    ipClone;

                ipClone = ip.clone();
                if (value && $.isArray(value)) {
                    ipClone.html(key);
                    node.append(ipClone);
                    value.forEach(function (item) {
                        var portClone = port.clone();
                        portClone.html(item);
                        node.append(portClone);
                    });
                } else {
                    ipClone.html(value);
                    node.append(ipClone);
                }
                view.append(node);
            });
        }

        span.html(view);
    }
    /* other */

    function previewNull() {
        return $('<span class="null"/>');
    }

    function previewText(data) {
        return $('<span/>').text(data);
    }

    function previewTemplate(data) {
        var tokens = data.split(/(\$\{.*?\})/g),
            $span = $('<span/>'),
            i,
            m;
        for (i = 0; i < tokens.length; i++) {
            if (i % 2 == 0) {
                $span.append($('<span/>').text(tokens[i]));
            } else if (m = tokens[i].match(/^\$\{(\/.*)\}$/)) {
                $span.append($('<a class="template" onclick="event.stopPropagation()"/>').attr('href', '#' + m[1]).text(m[0]));
            } else {
                $span.append($('<span class="template"/>').text(tokens[i]));
            }
        }
        return $span;
    }

    function previewSymlink(data) {
        return $('<a class="symlink" onclick="event.stopPropagation()"/>').attr('href', '#' + data).text(data);
    }

    function previewCase(data) {
        var result = $('<span class="case-preview"/>'),
            cases;
        try {
            cases = $.parseJSON(data);
            $.each(cases, function (id, value) {
                var $key = $('<span class="case-key"/>'),
                    def = false,
                    $span,
                    $value;

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
                $span = $('<span/>');
                if (!def) {
                    $span.append($key).append(': ');
                }
                $value = mimeType[value.mime].preview(value.value);
                if (value.mime === 'application/x-case') {
                    $value.prepend('{ ').append(' }');
                }
                $span.append($('<span class="case-value"/>').append($value)).appendTo(result);
                if (id < cases.length - 1) {
                    result.append('; ');
                }
            });
        } catch (ignore) {
        }
        return result;
    }

    function viewNull(span) {
        span.empty().append('<span class="null">NULL</span>');
    }

    function viewText(span, mime, data) {
        span.empty();
        var viewer = $('<span class="view-text"/>').appendTo(span),
            cm = new CodeMirror(viewer[0], { readOnly: true, mode: mime, tabindex: -1 });

        cm.setValue(data !== null ? data : '');
        createCodeMirrorToolbar(cm).prependTo(span);
    }

    function viewSymlink(span, mime, data) {
        span.empty().append($('<a class="symlink"/>').attr('href', '#' + data).text(data));
    }

    function viewCase(span, mime, data) {
        span.empty();
        var cspan = $('<div class="case-view"/>').appendTo(span);
        try {
            var cases = $.parseJSON(data);
            $.each(cases, function (id, value) {
                var $key = $('<span class="case-key"/>'),
                    kspan = $('<span class="case-key-block"/>')
                        .append($key)
                        .append(': '),
                    vspan = $('<span class="case-value case-value-block"/>');

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
        } catch (ignore) {
        }
    }

    function editNull(span) {
        span.empty().parent('div').hide();
        return function () { return null; };
    }

    function editText(span, mime, data) {
        span.empty().parent('div').show();
        var editor = $('<span/>').appendTo(span),
            cm = new CodeMirror(editor[0], { mode: mime, tabMode: mime === 'text/plain' ? 'default' : 'shift' });

        cm.setValue(data !== null ? data : '');
        editor.find('.CodeMirror').addClass('ui-widget-content ui-corner-all');
        createCodeMirrorToolbar(cm).prependTo(span);
        return function () { return cm.getValue(); };
    }

    function editTemplate(span, mime, data) {
        var result = editText(span, mime, data);
        $('<ul class="note"/>')
            .append('<li>${hostname} - полное имя хоста</li>')
            .append('<li>${short_hostname} - сокращенное имя хоста</li>')
            .append('<li>${ip} - ip-адрес, соответствующий хосту</li>')
            .append('<li>${/path/of/parameter} - значение параметра</li>')
            .appendTo(span);
        return result;
    }

    function editSymlink(span, mime, data) {
        span.empty().parent('div').show();
        var text = $('<input/>').addClass('input-width-fill').val(data).appendTo(span)
            .wrap($('<div class="ui-widget-content ui-corner-all input-width-fill-wrapper"/>'))
            .autocompletePath(true);
        return function () { return text.val(); };
    }

    function editCase(span, mime, data) {
        span.empty().parent('div').show();
        var defaultExists = false,
            container = $('<div/>').appendTo(span),
            changeKey = function () {
                var $kvSpan = $(this).parent('span.label').find('~ span.value'),
                    val = $kvSpan.find('select, input').val(),
                    $dc;

                $kvSpan.empty();
                if ($(this).val() === 'datacenter') {
                    $dc = $('<select name="datacenter"/>').appendTo($kvSpan);
                    getCached('/config/onlineconf/datacenter?symlink=resolve', function (data) {
                        $dc.empty();
                        $.each(data.children, function (id, param) {
                            $('<option/>').val(param.name).text(param.summary || param.name).appendTo($dc);
                        });
                        $dc.val(val);
                    });
                } else if ($(this).val() === 'group') {
                    $dc = $('<select name="group"/>').appendTo($kvSpan);
                    getCached('/config/onlineconf/group?symlink=resolve', function (data) {
                        $dc.empty();
                        $.each(data.children, function (id, param) {
                            if (param.name === 'priority') {
                                return;
                            }
                            $('<option/>').val(param.name).text(param.summary || param.name).appendTo($dc);
                        });
                        $dc.val(val);
                    });
                } else if ($(this).val() === 'server') {
                    $('<input name="server"/>')
                        .val(val)
                        .appendTo($kvSpan)
                        .addClass('input-width-fill')
                        .wrap($('<div class="ui-widget-content ui-corner-all input-width-fill-wrapper"/>'));
                }
                $(this).parents('.value-case:first').siblings().find('> div > span > select > option[value=default]').toggle($(this).val() !== 'default');
            },
            addFunc = function (kt, kv, m, v) {
                var div = $('<div class="value-case nice-form ui-corner-all"/>').appendTo(container),
                    keyType = $('<select name="key"/>')
                        .append('<option value="default">Default</option>')
                        .append('<option value="server">Сервер</option>')
                        .append('<option value="group">Группа</option>')
                        .append('<option value="datacenter">Датацентр</option>')
                        .change(changeKey)
                        .val(kt);
                if (span.find('> div > div > div > span > select > option[value=default]:selected').length) {
                    keyType.find('option[value=default]').hide();
                }
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

    mimeType = {
        "application/x-null": { title: "Null", edit: editNull, preview: previewNull, view: viewNull, validate: validateNone },
        "text/plain": { title: "Text", edit: editText, preview: previewText, view: viewText, validate: validateNone },
        "application/x-template": { title: "Template", edit: editTemplate, preview: previewTemplate, view: viewText, validate: validateNone },
        "application/json": { title: "JSON", edit: editText, preview: previewText, view: viewText, validate: validateJSON },
        "application/x-yaml": { title: "YAML", edit: editText, preview: previewText, view: viewText, validate: validateNone },
        "application/x-symlink": { title: "Symbolic link", edit: editSymlink, preview: previewSymlink, view: viewSymlink, validate: validateNone },
        "application/x-case": { title: "Case", edit: editCase, preview: previewCase, view: viewCase, validate: validateNone },
        "application/x-list": { title: "Список", edit: editServer, preview: previewText, view: viewServer, validate: validateNone },//OK
        "application/x-server": { title: "Список пар ip:port", edit: editServer, preview: previewText, view: viewServer, validate: validateNone },//OK
        "application/x-server2": { title: "Список портов для ip", edit: editServer, preview: previewText, view: viewServer, validate: validateNone }
    };

    CodeMirror.defineMode("template", function () {
        return {
            startState: function () {
                return {
                    variable: false,
                };
            },
            token: function (stream, state) {
                if (state.variable && stream.skipTo('}')) {
                    stream.next();
                    state.variable = false;
                    return 'variable-2';
                } else if (!state.variable && stream.skipTo('${')) {
                    state.variable = true;
                    return null;
                } else {
                    stream.skipToEnd();
                    return null;
                }
            }
        };
    });
    CodeMirror.defineMIME("application/x-template", "template");

    CodeMirror.on(window, "resize", function() {
        $('.CodeMirror-fullscreen').height($(window).height());
    });
});
