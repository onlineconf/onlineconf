$(function() {
    $('#access-dialog').node_dialog({
        title: 'Доступ к',
        async: true,
        refresh: function (event, ui) {
            var $access = $(this).empty();
            var node = $access.node_dialog('option', 'node').data('node');
            $.get('/access' + node.path, function (data) {
                $.each(data, function (id, group) {
                    $('<div/>')
                        .append($('<span class="label"/>').text(group.group))
                        .append(
                            $('<span class="value"/>')
                                .append(
                                    $('<input type="checkbox" class="access-overridden" title="Переопределить"/>')
                                        .prop('checked', group.overridden)
                                        .click(function () {
                                            var $slider = $(this).parent('span.value').find('.ui-slider');
                                            if ($(this).prop('checked')) {
                                                $slider
                                                    .slider('option', 'disabled', false)
                                                    .find('a.ui-slider-handle > span').addClass('wait-for-set').text('?');
                                            } else if ($slider.find('a.ui-slider-handle > span').hasClass('wait-for-set')) {
                                                $slider
                                                    .slider('option', 'disabled', true)
                                                    .slider('value', $slider.slider('value'));
                                            } else {
                                                $.ajax({
                                                    type: 'DELETE',
                                                    url: '/access' + node.path,
                                                    data: { group: group.group },
                                                    success: function (data) {
                                                        $slider
                                                            .slider('option', 'disabled', true)
                                                            .slider('value', data.rw == null ? 0 : data.rw ? 2 : 1);
                                                        $('#tree').jstree('refresh', $access.node_dialog('option', 'node'));
                                                    }
                                                });
                                            }
                                        })
                                )
                                .append($('<span class="access-slider"/>').append(
                                    $('<div/>')
                                        .slider({
                                            min: 0,
                                            max: 2,
                                            range: 'min',
                                            value: group.rw == null ? 0 : group.rw ? 2 : 1,
                                            slide: function (event, ui) {
                                                $(this).find('a.ui-slider-handle > span').removeClass('wait-for-set').text(ui.value == 2 ? 'rw' : ui.value == 1 ? 'ro' : 'no');
                                            },
                                            change: function (event, ui) {
                                                $(this).find('a.ui-slider-handle > span').removeClass('wait-for-set').text(ui.value == 2 ? 'rw' : ui.value == 1 ? 'ro' : 'no');
                                                if (!$(this).slider('option', 'disabled')) {
                                                    $.post(
                                                        '/access' + node.path,
                                                        { group: group.group, rw: ui.value == 2 ? true : ui.value == 1 ? false : null },
                                                        function (data) { $('#tree').jstree('refresh', $access.node_dialog('option', 'node')) }
                                                    );
                                                }
                                            }
                                        })
                                        .slider('option', 'disabled', !group.overridden)
                                        .find('a.ui-slider-handle').append($('<span/>').text(group.rw == null ? 'no' : group.rw ? 'rw' : 'ro')).end()
                                ))
                        )
                        .appendTo($access);
                });
                ui.success();
            });
        }
    });
});
