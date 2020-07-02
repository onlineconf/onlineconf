ALTER TABLE `my_config_activity`
    MODIFY `Package` varchar(32) CHARACTER SET utf8mb4 NOT NULL,
    DEFAULT CHARACTER SET utf8mb4;

ALTER TABLE `my_config_group`
    MODIFY `Name` varchar(128) CHARACTER SET utf8mb4 NOT NULL,
    DEFAULT CHARACTER SET utf8mb4;

ALTER TABLE `my_config_tree`
    MODIFY `Summary` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
    MODIFY `Description` varchar(8192) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
    DEFAULT CHARACTER SET utf8mb4;

ALTER TABLE `my_config_tree_log`
    MODIFY `Comment` varchar(512) CHARACTER SET utf8mb4 DEFAULT NULL,
    DEFAULT CHARACTER SET utf8mb4;

ALTER TABLE `my_config_user_group`
    DEFAULT CHARACTER SET utf8mb4;

ALTER TABLE `my_config_tree_group`
    DEFAULT CHARACTER SET utf8mb4;

ALTER DATABASE `onlineconf`
    DEFAULT CHARACTER SET utf8mb4;
