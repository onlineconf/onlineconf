// Node log
$(function() {
    $('#log-dialog').node_dialog({
        title: 'Журнал',
        async: true,
        dialog: {
            minWidth: 800,
            minHeight: 70,
            maxHeight: 800
        },
        refresh: function (event, ui) {
            var $dialog = $(this);
            var $log = $('#log');
            var node = $(this).node_dialog('option', 'node').data('node');
            $.get('/log' + node.path, function(data) {
                $log.empty();
                $.each(data, function(id, log) {
                    var $data = $('<td class="log-data"/>');
                    $('<tr/>')
                        .append('<td class="log-version">' + log.version + '</td>')
                        .append($data)
                        .append('<td class="log-mtime">' + log.mtime + '</td>')
                        .append('<td class="log-author">' + log.author + '</td>')
                        .append('<td class="log-comment">' + (log.comment != null ? log.comment : '') + '</td>')
                        .appendTo($log);
                    if (log.deleted) {
                        $data.addClass('log-deleted').text('удален');
                    } else if (log.rw == null) {
                        $data.addClass('log-no-access').text('нет доступа');
                    } else if (log.data != null && log.data != '') {
                        mimeType[log.mime].view($data, log.mime, log.data);
                    }
                });
                ui.success();
                $log.find('.CodeMirror').each(function () {
                    this.CodeMirror.refresh();
                });
            })
        }
    });
});

// Global log
$(function() {
    $('#global-log-dialog').dialog({
        title: 'Журнал',
        minWidth: 800,
        minHeight: 70,
        maxHeight: 800,
        autoOpen: false,
        create: function () {
            $('#global-log-form').find('input:not(:checkbox)').val('');
            $('#global-log-form').find('input:checkbox').prop('checked', false);
            $('#global-log-header').addClass('ui-widget-content').insertBefore($(this));
        },
        open: function () { $('#global-log-form').submit() }
    });
    $('#global-log').click(function() { $('#global-log-dialog').dialog('open'); return false });
    $('#global-log-author').autocomplete({ source: '/user', minLength: 2 });
    $('#global-log-branch').autocompletePath();
    var $dates = $('#global-log-from, #global-log-till').datepicker({
        dateFormat: 'yy-mm-dd',
        onSelect: function (selectedDate) {
            var option = this.id == "global-log-from" ? "minDate" : "maxDate";
            var instance = $(this).data('datepicker');
            var date = $.datepicker.parseDate(instance.settings.dateFormat || $.datepicker._defaults.dateFormat, selectedDate, instance.settings);
            $dates.not(this).datepicker('option', option, date);
        }
    });
    $('#global-log-go').button({ icons: { primary: 'ui-icon-arrowthick-1-e' }, text: false });
    $('#global-log-form').submit(function () {
        var $dialog = $('#global-log-dialog');
        var $log = $('#global-log-table').empty();
        $.get('/global-log', $(this).serialize(), function(data) {
            $log.empty();
            $.each(data, function(id, log) {
                var $data = $('<td class="log-data"/>');
                $('<tr/>')
                    .append('<td class="log-mtime">' + log.mtime + '</td>')
                    .append('<td class="log-path">' + log.path + '</td>')
                    .append($data)
                    .append('<td class="log-author">' + log.author + '</td>')
                    .append('<td class="log-comment">' + (log.comment != null ? log.comment : '') + '</td>')
                    .appendTo($log);
                if (log.deleted) {
                    $data.addClass('log-deleted').text('удален');
                } else if (log.rw == null) {
                    $data.addClass('log-no-access').text('нет доступа');
                } else if (log.data != null && log.data != '') {
                    mimeType[log.mime].view($data, log.mime, log.data);
                }
            });
            $dialog
                .dialog('option', 'width', window.innerWidth - 100 > $log.width() + 40 ? $log.width() + 40 : window.innerWidth - 100)
                .dialog('option', 'height', window.innerHeight - 100 > $log.height() + 110 ? $log.height() + 110 : window.innerHeight - 100)
                .dialog('option', 'position', 'center');
        });
        return false;
    });
});
