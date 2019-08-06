CREATE TABLE `my_change_notification` (
    `ID` bigint(20) unsigned NOT NULL,
    `CTime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `Origin` varchar(255) NOT NULL,
    `Message` mediumtext NOT NULL,
    PRIMARY KEY (`ID`),
    UNIQUE KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
