CREATE TABLE `my_config_tree` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `Name` varchar(256) character set ascii collate ascii_bin NOT NULL,
    `ParentID` bigint(20) unsigned default NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL,
    `Value` mediumtext character set utf8mb4 collate utf8mb4_bin,
    `ContentType` varchar(255) character set ascii NOT NULL default 'application/x-null',
    `Summary` varchar(255) NOT NULL default '',
    `Description` varchar(8192) NOT NULL default '',
    `Version` int(11) NOT NULL default '1',
    `MTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Deleted` tinyint(1) NOT NULL default '0',
    `Notification` enum('none','no-value','with-value') character set ascii default NULL,
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `ID` (`ID`),
    UNIQUE KEY `Path` (`Path`),
    UNIQUE KEY `Name` (`ParentID`,`Name`),
    CONSTRAINT `my_config_tree_ibfk_1` FOREIGN KEY (`ParentID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_tree_log` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `NodeID` bigint(20) unsigned NOT NULL,
    `Version` int(11) NOT NULL,
    `Value` mediumtext character set utf8mb4 collate utf8mb4_bin,
    `ContentType` varchar(255) character set ascii NOT NULL default 'application/x-null',
    `Author` varchar(128) character set ascii NOT NULL,
    `MTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Comment` varchar(512) default NULL,
    `Deleted` tinyint(1) NOT NULL default '0',
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `NodeID` (`NodeID`,`Version`),
    KEY `MTime` (`MTime`),
    CONSTRAINT `my_config_tree_log_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_group` (
    `ID` int(11) NOT NULL auto_increment,
    `Name` varchar(128) NOT NULL,
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `Name` (`Name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_user_group` (
    `User` varchar(128) character set ascii NOT NULL,
    `GroupID` int(11) NOT NULL,
    UNIQUE KEY `User_Group` (`User`,`GroupID`),
    KEY `User` (`User`),
    KEY `GroupID` (`GroupID`),
    CONSTRAINT `my_config_user_group_ibfk_1` FOREIGN KEY (`GroupID`) REFERENCES `my_config_group` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_tree_group` (
    `NodeID` bigint(20) unsigned NOT NULL,
    `GroupID` int(11) NOT NULL,
    `RW` tinyint(1) default '0',
    PRIMARY KEY  (`NodeID`,`GroupID`),
    KEY `GroupID` (`GroupID`),
    CONSTRAINT `my_config_tree_group_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`) ON DELETE CASCADE,
    CONSTRAINT `my_config_tree_group_ibfk_2` FOREIGN KEY (`GroupID`) REFERENCES `my_config_group` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_activity` (
    `Host` varchar(255) character set ascii NOT NULL,
    `Time` timestamp NULL default NULL,
    `Online` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Package` varchar(32) NOT NULL,
    PRIMARY KEY  (`Host`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

delimiter $$

CREATE TRIGGER `my_config_tree_path` BEFORE INSERT ON `my_config_tree` FOR EACH ROW
BEGIN
    DECLARE cnt int;
    DECLARE parent_path varchar(512);
    IF NEW.ParentID IS NULL THEN
        SELECT count(*) INTO cnt FROM my_config_tree WHERE ParentID IS NULL;
        IF cnt = 0 THEN
            SET NEW.Name = '';
            SET NEW.Path = '/';
        ELSE
            SET NEW.Name = NULL;
        END IF;
    ELSEIF NEW.Name LIKE '%/%' THEN
        SET NEW.Name = NULL;
    ELSE
        SELECT `Path` INTO parent_path FROM `my_config_tree` where `ID` = NEW.`ParentID`;
        IF parent_path = '/' THEN
            SET NEW.Path = CONCAT('/', NEW.Name);
        ELSE
            SET NEW.Path = CONCAT(parent_path, '/', NEW.Name);
        END IF;
    END IF;
END;$$

CREATE TRIGGER `my_config_tree_move` BEFORE UPDATE ON `my_config_tree` FOR EACH ROW
BEGIN
    DECLARE parent_path varchar(512);
    IF NEW.Path <> OLD.Path OR NEW.Name <> OLD.Name OR NEW.ParentID <> OLD.ParentID THEN
        SELECT `Path` INTO parent_path FROM `my_config_tree` where `ID` = NEW.`ParentID`;
        IF parent_path = '/' THEN
            SET NEW.Path = CONCAT('/', NEW.Name);
        ELSE
            SET NEW.Path = CONCAT(parent_path, '/', NEW.Name);
        END IF;
    END IF;
END;$$

CREATE FUNCTION `my_config_tree_access` (`node_id` bigint(20) unsigned, `username` varchar(256)) RETURNS boolean
READS SQL DATA
BEGIN
    DECLARE `result` boolean;
    DECLARE `current_node_id` bigint(20) unsigned;
    DECLARE `group_id` int(11);
    DECLARE `rw` boolean;
    DECLARE `done` boolean DEFAULT false;
    DECLARE `group_cur` CURSOR FOR
        SELECT ug.`GroupID`
        FROM `my_config_user_group` ug
        WHERE ug.`User` = `username`;

    OPEN `group_cur`;
    group_loop: LOOP
        BEGIN
            DECLARE EXIT HANDLER FOR NOT FOUND BEGIN SET `done` = true; END;
            FETCH `group_cur` INTO `group_id`;
        END;
        IF `done` THEN
            LEAVE group_loop;
        END IF;

        SET `current_node_id` = `node_id`;
        node_loop: WHILE `current_node_id` IS NOT NULL DO
            BEGIN
                DECLARE `current_result` boolean;
                DECLARE EXIT HANDLER FOR NOT FOUND BEGIN END;

                SELECT tg.`RW` INTO `current_result`
                FROM `my_config_tree_group` tg
                WHERE tg.`NodeID` = `current_node_id`
                AND tg.`GroupID` = `group_id`;

                IF `current_result` IS NOT NULL AND (`result` IS NULL OR `current_result` > `result`) THEN
                    SET `result` = `current_result`;
                    IF `result` THEN
                        LEAVE group_loop;
                    END IF;
                END IF;
                LEAVE node_loop;
            END;
            SELECT `ParentID` INTO `current_node_id` FROM `my_config_tree` WHERE `ID` = `current_node_id`;
        END WHILE;
    END LOOP;
    CLOSE `group_cur`;

    RETURN `result`;
END;$$

CREATE FUNCTION `my_config_tree_group_access` (`node_id` bigint(20) unsigned, `group_id` int(11)) RETURNS boolean
READS SQL DATA
BEGIN
    DECLARE `result` boolean;

    WHILE `node_id` IS NOT NULL DO
        BEGIN
            DECLARE EXIT HANDLER FOR SQLSTATE '02000' BEGIN END;

            SELECT `RW` INTO `result`
            FROM `my_config_tree_group` tg
            WHERE tg.`NodeID` = `node_id` AND tg.`GroupID` = `group_id`
            ORDER BY `RW` DESC
            LIMIT 1;

            RETURN `result`;
        END;

        SELECT `ParentID` INTO `node_id` FROM `my_config_tree` WHERE `ID` = `node_id`;
    END WHILE;

    RETURN `result`;
END;$$

CREATE FUNCTION `my_config_tree_notification` (`node_id` bigint(20) unsigned) RETURNS enum('none','no-value','with-value')
READS SQL DATA
BEGIN
    DECLARE `notification` enum('none', 'no-value', 'with-value');

    WHILE `node_id` IS NOT NULL DO
        SELECT t.`ParentID`, t.`Notification` INTO `node_id`, `notification`
        FROM `my_config_tree` t
        WHERE t.`ID` = `node_id`;

        IF `notification` IS NOT NULL THEN
            RETURN `notification`;
        END IF;
    END WHILE;

    RETURN NULL;
END;$$

delimiter ;

INSERT INTO `my_config_tree` (`ID`, `Name`, `ParentID`, `Notification`) VALUES
    (1, '', NULL, 'none'),
    (2, 'onlineconf', 1, NULL),
    (3, 'module', 2, NULL),
    (4, 'service', 2, NULL);

INSERT INTO `my_config_tree_log` (`NodeID`, `Version`, `Value`, `ContentType`, `Author`, `MTime`, `Comment`, `Deleted`)
SELECT `ID`, `Version`, `Value`, `ContentType`, 'onlineconf', `MTime`, 'Init onlineconf', `Deleted` FROM `my_config_tree` ORDER BY `ID`;

INSERT INTO `my_config_group` (`Name`) VALUES ('root');

INSERT INTO `my_config_tree_group` (`NodeID`, `GroupID`, `RW`)
VALUES (
    (SELECT `ID` from `my_config_tree` WHERE `Path` = '/'),
    (SELECT `ID` from `my_config_group` WHERE `Name` = 'root'),
    true
);
