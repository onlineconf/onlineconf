drop table my_config;

alter table my_config_activity
    modify `Host` varchar(255) character set ascii NOT NULL,
    modify `Package` varchar(32) character set utf8mb4 DEFAULT NULL,
    default character set utf8mb4;

alter table my_config_group
    modify `Name` varchar(128) character set utf8mb4 NOT NULL,
    default character set utf8mb4;

drop table my_config_log;

drop table my_config_module;

drop table my_config_module_group;

drop table my_config_transaction;

alter table my_config_tree
    modify `Value` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
    modify `ContentType` varchar(255) CHARACTER SET ascii NOT NULL DEFAULT 'application/x-null',
    modify `Summary` varchar(255) character set utf8mb4 NOT NULL DEFAULT '',
    modify `Description` varchar(8192) character set utf8mb4 NOT NULL DEFAULT '',
    modify `Notification` enum('none','no-value','with-value') CHARACTER SET ascii DEFAULT NULL,
    default character set utf8mb4;

alter table my_config_tree_log
    modify `Value` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
    modify `ContentType` varchar(255) CHARACTER SET ascii NOT NULL DEFAULT 'application/x-null',
    modify `Author` varchar(128) CHARACTER SET ascii NOT NULL,
    modify `Comment` varchar(512) character set utf8mb4 DEFAULT NULL,
    default character set utf8mb4;

alter table my_config_user_group
    modify `User` varchar(128) character set ascii NOT NULL,
    default character set utf8mb4;

alter database onlineconf
    default character set utf8mb4;
