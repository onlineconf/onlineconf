ALTER TABLE `my_change_notification`
    MODIFY `Origin` varchar(255) CHARACTER SET utf8mb4 NOT NULL,
    MODIFY `Message` mediumtext CHARACTER SET utf8mb4 NOT NULL,
    DEFAUlT CHARSET=utf8mb4;
