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
                this.element.css('visibility', 'hidden');
            }
            var node = this.options.node.data('node');
            this.element.dialog('option', 'title', (this.options.title == null ? '' : this.options.title + ' ') + node.path);
            var self = this;
            var afterRefresh = function () {
                self._refreshing = false;
                self.element.css('visibility', 'inherit').dialog('open');
            };
            this._trigger('refresh', null, { success: afterRefresh });
            if (!this.options.async) afterRefresh();
        },
        open: function () {
            if (!this.element.dialog('isOpen') && !this._refreshing) {
                this._refresh();
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

$(function() {
    $.widget('onlineconf.overridable_slider', {
        options: {
            values: [],
            labels: [],
            titles: [],
            disabledValues: [],
            overridden: false,
            value: null,
            change: function (event, ui) { ui.success(ui) },
            _wait_for_set: false,
        },
        _create: function () {
            var self = this;
            var index = this._index(this.options.value);
            this.element
                .append(
                    $('<input type="checkbox" class="inherit-overridden" title="Переопределить"/>')
                        .prop('checked', this.options.overridden)
                        .click(function () {
                            var $slider = self.element.find('.ui-slider');
                            if ($(this).prop('checked')) {
                                self._setOption('overridden', true);
                                self._setOption('_wait_for_set', true);
                            } else if (self.options._wait_for_set) {
                                self._setOption('overridden', false);
                                self._setOption('_wait_for_set', false);
                            } else {
                                self._trigger('change', null, {
                                    overridden: false,
                                    success: function (data) { self._setOptions(data) }
                                });
                            }
                        })
                )
                .append($('<span class="enum-slider"/>').append(
                    $('<div/>')
                        .slider({
                            min: 0,
                            max: this.options.values.length - 1,
                            range: 'min',
                            value: index,
                            slide: function (event, ui) {
                                if (self.options.disabledValues[ui.value]) return false;
                                $(this).find('a.ui-slider-handle')
                                    .attr('title', self.options.titles[ui.value])
                                    .find('> span').removeClass('wait-for-set').text(self.options.labels[ui.value]);
                            },
                            change: function (event, ui) {
                                $(this).find('a.ui-slider-handle')
                                    .attr('title', self.options.titles[ui.value])
                                    .find('> span').removeClass('wait-for-set').text(self.options.labels[ui.value]);
                                var value = self.options.values[ui.value];
                                if (!$(this).slider('option', 'disabled')) {
                                    self._trigger('change', null, {
                                        overridden: true,
                                        value: value,
                                        success: function (data) { self._setOptions(data) }
                                    });
                                }
                            }
                        })
                        .slider('option', 'disabled', !this.options.overridden)
                        .find('a.ui-slider-handle')
                            .attr('title', this.options.titles[index])
                            .append($('<span/>').text(this.options.labels[index]))
                            .prop('tabIndex', this.options.overridden ? null : -1)
                            .end()
                    )
                );
        },
        destroy: function () {
            this.element.empty();
            $.Widget.prototype.destroy.apply(this, arguments);
        },
        _setOption: function (key, value) {
            $.Widget.prototype._setOption.apply(this, arguments);
            if (key === "overridden") {
                this.element
                    .find('input.inherit-overridden').prop('checked', value).end()
                    .find('.ui-slider').slider('option', 'disabled', !value)
                        .find('a.ui-slider-handle').prop('tabIndex', value ? null : -1).end();
            } else if (key === "value") {
                var index = this._index(value);
                var $slider = this.element.find('.ui-slider')
                if ($slider.slider('option', 'value') != index) {
                    $slider.slider('option', 'value', index);
                }
                this._setOption('_wait_for_set', false);
            } else if (key === "_wait_for_set") {
                if (value) {
                    this.element.find('a.ui-slider-handle > span').addClass('wait-for-set').text('?');
                } else if (this.element.find('a.ui-slider-handle > span').hasClass('wait-for-set')) {
                    this.element.find('.ui-slider').slider('option', 'value', this._index(this.options.value));
                }
            }
            return this;
        },
        _index: function (value) {
            for (var i = 0; i < this.options.values.length; i++) {
                if (this.options.values[i] == value) {
                    return i;
                }
            }
            return 0;
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
    $.fn.autocompletePath = function (resolveSymlink) {
        var last;
        this.autocomplete({
            source: function(request, response) {
                var path = request.term.replace(/[^\/]+$/, '');
                if (last != null && path == last.path) {
                    response($.grep(last.resp, function (v, i) { return v.value.indexOf(request.term) == 0 }));
                } else {
                    var convPath = path == "/" ? path : path.replace(/\/$/, '')
                    $.get('/config' + convPath + (resolveSymlink ? '?symlink=resolve' : ''), function (data) {
                        var resp = [];
                        $.each(data.children || [], function (id, child) {
                            resp.push({ label: child.name, value: child.path.replace(/^.*\//, path) });
                        });
                        response($.grep(resp, function (v, i) { return v.value.indexOf(request.term) == 0 }));
                        last = { path: path, resp: resp };
                    });
                }
            }
        });
        return this;
    }
});

// Monitoring
$(function() {
    function showMonitoring (event) {
        $('#monitoring-list').empty();
        var params = {};
        var matches = $(this).attr('id').match(/^monitoring-(\w+)$/);
        if (matches) params.sort = matches[1];
        $.get('/monitoring', params, function (data) {
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
        return false;
    }

    $('#monitoring, #monitoring-dialog th a').click(showMonitoring);
    $('#monitoring-dialog').dialog({
        autoOpen: false,
        width: 600
    });
});
