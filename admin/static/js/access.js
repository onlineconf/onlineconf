$(function() {
    $('#access-dialog').node_dialog({
        title: 'Доступ к',
        async: true,
        refresh: function (event, ui) {
            var $access = $(this).empty();
            var node = $access.node_dialog('option', 'node').data('node');
            $.get('/access' + node.path, function (data) {
                $access.empty();
                $.each(data, function (id, group) {
                    $('<div/>')
                        .append($('<span class="label"/>').text(group.group))
                        .append(
                            $('<span class="value"/>')
                                .overridable_slider({
                                    values: [null, false, true],
                                    labels: ['no', 'ro', 'rw'],
                                    titles: ['Нет доступа', 'Только чтение', 'Чтение и запись'],
                                    overridden: group.overridden,
                                    value: group.rw,
                                    change: function (event, ui) {
                                        if (ui.overridden) {
                                            $.post(
                                                '/access' + node.path,
                                                { group: group.group, rw: ui.value },
                                                function (data) {
                                                    ui.success({ overridden: true, value: data.rw })
                                                    $('#tree').jstree('refresh', $access.node_dialog('option', 'node'));
                                                }
                                            );
                                        } else {
                                            $.ajax({
                                                type: 'DELETE',
                                                url: '/access' + node.path,
                                                data: { group: group.group },
                                                success: function (data) {
                                                    ui.success({ overridden: false, value: data.rw });
                                                    $('#tree').jstree('refresh', $access.node_dialog('option', 'node'));
                                                }
                                            });
                                        }
                                    }
                                })
                        )
                        .appendTo($access);
                });
                ui.success();
            });
        }
    });
});
