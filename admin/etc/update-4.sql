alter table my_config_user_group
    add foreign key (GroupID) references my_config_group(ID) on delete cascade;

alter table my_config_tree_group
    drop foreign key my_config_tree_group_ibfk_1,
    drop foreign key my_config_tree_group_ibfk_2;

alter table my_config_tree_group
    add foreign key (NodeID) references my_config_tree(ID) on delete cascade,
    add foreign key (GroupID) references my_config_group(ID) on delete cascade;
