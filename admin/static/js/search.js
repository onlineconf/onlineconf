$(function() {
    $('#search-text').val('');
    $('#search-button')
        .button({
            icons: { primary: 'ui-icon-search' },
            text: false
        })
        .click(function () {
            if ($(this).button('option', 'icons').primary == 'ui-icon-close') {
                $('#search-text').val('');
            }
            $('#search-form').submit();
        });
    $('#search-form').submit(function () {
        var val = $('#search-text').val();
        if (val.length) {
            $('#tree').jstree('search', val);
            $('#search-button').button('option', 'icons', { primary: 'ui-icon-close' });
        } else {
            $('#tree').jstree('clear_search');
            $('#search-button').button('option', 'icons', { primary: 'ui-icon-search' });
        }
        return false;
    });
    $.expr[':'].jstree_node_contains = function (a, i, m) {
        var node = $(a).parent('li').data('node');
        return node.name.toLowerCase().indexOf(m[3].toLowerCase()) >= 0
            || node.data != null && node.data.toLowerCase().indexOf(m[3].toLowerCase()) >= 0
            || node.summary != null && node.summary.toLowerCase().indexOf(m[3].toLowerCase()) >= 0
            || node.description != null && node.description.toLowerCase().indexOf(m[3].toLowerCase()) >= 0;
    };
});
