$(function() {
    $('#group-dialog').dialog({
        modal: true,
        autoOpen: false,
        buttons: {
            'Cоздать группу': function() {
                $('#create-group-dialog').dialog('open');
            },
            'Закрыть': function() {
                $(this).dialog('close');
            }
        },
        open: function () {
            if ($(this).hasClass('jstree')) {
                $(this).jstree('refresh');
                return;
            }
            $(this).jstree({
                plugins: ['json_data', 'types', 'ui', 'contextmenu', 'sort', 'themes'],
                core: { animation: 100 },
                json_data: {
                    data: function ($node, success) {
                        var tree = this;
                        $.get(
                            '/group/' + ($node != -1 ? $node.data('group') : ''),
                            {},
                            $node == -1
                                ? function (data, text, jqxhr) {
                                    var result = [];
                                    $.each(data || [], function(id, group) {
                                        result.push({
                                            data: group,
                                            attr: { id: 'group-' + group, rel: 'group' },
                                            metadata: { group: group },
                                            state: 'closed'
                                        });
                                    });
                                    success.call(tree, result);
                                }
                                : function (data, text, jqxhr) {
                                    var result = [];
                                    $.each(data || [], function(id, user) {
                                        result.push({
                                            data: user,
                                            attr: { rel: 'user' },
                                            metadata: { user: user },
                                            state: null
                                        });
                                    });
                                    success.call(tree, result);
                                }
                        );
                    }
                },
                types: {
                    valid_children: ['group'],
                    max_depth: 2,
                    types: {
                        group: {
                            icon: { image: 'css/type/group.png' },
                            valid_children: ['user']
                        },
                        user: {
                            icon: { image: 'css/type/user.png' },
                            valid_children: 'none'
                        }
                    }
                },
                contextmenu: {
                    select_node: true,
                    items: function (node) {
                        if (node.attr('rel') == "group") {
                            return {
                                refresh: {
                                    label: 'Обновить',
                                    separator_after: true,
                                    action: function(node) {
                                        this.refresh(node);
                                    }
                                },
                                remove: {
                                    label: 'Удалить группу',
                                    _disabled: !this.is_leaf(node),
                                    action: function (node) {
                                        this.open_node(node);
                                        $('#delete-group-dialog').data('node', node).dialog('open');
                                    }
                                },
                                add_user: {
                                    label: 'Добавить пользователя',
                                    action: function (node) {
                                        this.open_node(node);
                                        $('#add-user-dialog')
                                            .data('node', node)
                                            .dialog('option', 'title', 'Добавить в группу ' + node.data('group'))
                                            .dialog('open');
                                    }
                                },
                            };
                        } else {
                            return {
                                remove: {
                                    label: 'Удалить пользователя',
                                    action: function (node) {
                                        $('#delete-user-dialog').data('node', node).dialog('open');
                                    }
                                }
                            }
                        }
                    }
                },
                ui: { select_limit: 1 },
                themes: { theme: 'default' }
            });
        }
    });
    $('#access-group').click(function() { $('#group-dialog').dialog('open'); return false });

    $('#create-group-dialog').dialog({
        modal: true,
        autoOpen: false,
        minHeight: 132,
        open: function () {
            $('#create-group-name').val('');
        },
        buttons: {
            'Добавить': function () {
                var group = $('#create-group-name').val();
                $.post(
                    '/group/' + group,
                    {},
                    function (data) {
                        $('#group-dialog').jstree('create_node', -1, 'inside', {
                            data: group,
                            attr: { id: 'group-' + group, rel: 'group' },
                            metadata: { group: group }
                        });
                    }
                );
                $(this).dialog('close');
            },
            'Отменить': function () {
                $(this).dialog('close');
            }
        }
    });

    $('#delete-group-dialog').dialog({
        modal: true,
        autoOpen: false,
        resizable: false,
        open: function() {
            var node = $(this).data('node');
            var group = $('#group-dialog').jstree('get_text', node);
            $(this).dialog('option', 'title', 'Удалить ' + group + '?');
            $('#delete-group-name').text(group);
        },
        buttons: {
            Удалить: function() {
                var node = $(this).data('node');
                var group = $('#group-dialog').jstree('get_text', node);
                $.ajax({
                    type: 'DELETE',
                    url: '/group/' + group,
                    success: function() {
                        $('#group-dialog').jstree('delete_node', node);
                    }
                });
                $(this).dialog('close');
            },
            Отменить: function() {
                $(this).dialog('close');
            }
        }
    });

    $('#add-user-dialog').dialog({
        modal: true,
        autoOpen: false,
        minHeight: 132,
        open: function () {
            $('#add-user-name').val('');
        },
        buttons: {
            'Добавить': function () {
                var node = $(this).data('node');
                var group = node.data('group');
                var user = $('#add-user-name').val();
                $.post(
                    '/group/' + group + '/' + user,
                    {},
                    function (data) {
                        $('#group-dialog').jstree('create_node', node, 'inside', {
                            data: user,
                            attr: { rel: 'user' },
                            metadata: { user: user },
                            state: null
                        });
                    }
                );
                $(this).dialog('close');
            },
            'Отменить': function () {
                $(this).dialog('close');
            }
        }
    });
    $('#add-user-name').autocomplete({ source: '/user', minLength: 2 });

    $('#delete-user-dialog').dialog({
        modal: true,
        autoOpen: false,
        resizable: false,
        open: function() {
            var node = $(this).data('node');
            var path = $('#group-dialog').jstree('get_path', node);
            $(this).dialog('option', 'title', 'Удалить ' + path[1] + '?');
            $('#delete-user-group').text(path[0]);
            $('#delete-user-name').text(path[1]);
        },
        buttons: {
            Удалить: function() {
                var node = $(this).data('node');
                var path = $('#group-dialog').jstree('get_path', node);
                $.ajax({
                    type: 'DELETE',
                    url: '/group/' + path[0] + '/' + path[1],
                    success: function() {
                        $('#group-dialog').jstree('delete_node', node);
                    }
                });
                $(this).dialog('close');
            },
            Отменить: function() {
                $(this).dialog('close');
            }
        }
    });
});
