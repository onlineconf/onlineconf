$(function() {
    $.widget('onlineconf.node_dialog', {
        options: {
            node: null,
            refresh: null,
            title: null,
            async: false,
            dialog: {}
        },
        _refreshing: false,
        _create: function () {
            this.element.dialog($.extend({ autoOpen: false }, this.options.dialog));
        },
        _setOptions: function () {
            $.Widget.prototype._setOptions.apply(this, arguments);
            this.refresh();
        },
        _refresh: function () {
            this._refreshing = true;
            if (this.options.async && this.element.dialog('isOpen')) {
                this.element.dialog('close');
            }
            var node = this.options.node.data('node');
            this.element.dialog('option', 'title', (this.options.title == null ? '' : this.options.title + ' ') + node.path);
            var self = this;
            this._trigger('refresh', null, { success: function () { self._refreshing = false; self.element.dialog('open') } });
            if (!this.options.async) this._refreshing = false;
        },
        open: function () {
            if (!this.element.dialog('isOpen') && !this._refreshing) {
                this._refresh();
                if (!this.options.async) {
                    this.element.dialog('open');
                }
            }
        },
        close: function () {
            this.element.dialog('close');
        },
        refresh: function () {
            if (this.element.dialog('isOpen')) {
                this._refresh();
            }
        }
    });
});

var can_edit_groups = false;

$(function() {
    $('#error-dialog')
        .dialog({
            modal: true,
            autoOpen:false,
            minHeight: 70,
            width: 600,
            buttons: { OK: function() { $(this).dialog('close') } }
        })
        .ajaxError(function(event, request, settings, exception) {
            var message;
            if (request.getResponseHeader('Content-Type') == 'application/json') {
                try {
                    var json = settings.converters['text json'](request.responseText);
                    message = json.message;
                } catch(e) {
                    message = 'Не удалось извлечь текст ошибки';
                }
            } else {
                message = exception;
            }
            $(this).text(message).dialog('option', 'title', exception).dialog('open');
        });

    $.get('/whoami', {}, function (data) {
        can_edit_groups = data.can_edit_groups;
        if (data.can_edit_groups) {
            $('#access-group').show();
        }
    });
});

$(function() {
    $.fn.autocompletePath = function () {
        var lastBranchAc;
        this.autocomplete({
            source: function(request, response) {
                if (/\/$/.test(request.term)) {
                    $.get('/config' + request.term, function(data) {
                        var resp = [];
                        $.each(data.children || [], function(id, child) {
                            resp.push({ label: child.name, value: child.path });
                        });
                        response(resp);
                        lastBranchAc = { term: request.term, resp: resp };
                    });
                } else if (lastBranchAc != null && request.term.indexOf(lastBranchAc.term) == 0) {
                    response($.grep(lastBranchAc.resp, function (v, i) { return v.value.indexOf(request.term) == 0 }));
                } else {
                    response([]);
                }
            }
        });
        return this;
    }
});

// Monitoring
$(function() {
    $('#monitoring').click(function () {
        $('#monitoring-list').empty();
        $.get('/monitoring', function (data) {
            $.each(data, function (id, host) {
                $('<tr/>')
                    .append($('<td/>').text(host.host))
                    .append($('<td/>').text(host.mtime).addClass(host.mtime_alert ? 'monitoring-alert' : ''))
                    .append($('<td/>').text(host.online).addClass(host.online_alert ? 'monitoring-alert' : ''))
                    .append($('<td/>').text(host.package))
                    .appendTo('#monitoring-list');
            });
            $('#monitoring-dialog').dialog('open');
        });
    });
    $('#monitoring-dialog').dialog({
        autoOpen: false,
        width: 600
    });
});
