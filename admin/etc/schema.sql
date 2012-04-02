CREATE TABLE `my_config_tree` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `Name` varchar(256) character set ascii collate ascii_bin NOT NULL,
    `ParentID` bigint(20) unsigned default NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL,
    `Value` text collate utf8_bin,
    `ContentType` varchar(256) NOT NULL default 'application/x-null',
    `Summary` varchar(256) NOT NULL default '',
    `Description` varchar(8192) NOT NULL default '',
    `Version` int(11) NOT NULL default '1',
    `MTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Deleted` tinyint(1) NOT NULL default '0',
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
    `Value` text collate utf8_bin,
    `ContentType` varchar(256) NOT NULL default 'application/x-null',
    `Author` varchar(128) character set ascii NOT NULL,
    `MTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Comment` varchar(256) default NULL,
    `Deleted` tinyint(1) NOT NULL default '0',
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `NodeID` (`NodeID`,`Version`),
    KEY `MTime` (`MTime`),
    CONSTRAINT `my_config_tree_log_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `my_config_tree_group` (
    `NodeID` bigint(20) unsigned NOT NULL,
    `GroupID` int(11) NOT NULL,
    `RW` tinyint(1) default '0',
    PRIMARY KEY  (`NodeID`,`GroupID`),
    KEY `GroupID` (`GroupID`),
    CONSTRAINT `my_config_tree_group_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`),
    CONSTRAINT `my_config_tree_group_ibfk_2` FOREIGN KEY (`GroupID`) REFERENCES `my_config_group` (`ID`)
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
    DECLARE `overridden` boolean;

    WHILE `node_id` IS NOT NULL DO
        SELECT `RW`, true INTO `result`, `overridden`
        FROM `my_config_tree_group` tg
        JOIN `my_config_user_group` ug ON ug.`GroupID` = tg.`GroupID`
        WHERE tg.`NodeID` = `node_id` AND ug.`User` = `username`
        ORDER BY `RW` DESC
        LIMIT 1;

        IF `overridden` THEN
            RETURN `result`;
        END IF;

        SELECT `ParentID` INTO `node_id` FROM `my_config_tree` WHERE `ID` = `node_id`;
    END WHILE;

    RETURN `result`;
END;$$

CREATE FUNCTION `my_config_tree_group_access` (`node_id` bigint(20) unsigned, `group_id` int(11)) RETURNS boolean
READS SQL DATA
BEGIN
    DECLARE `result` boolean;
    DECLARE `overridden` boolean;

    WHILE `node_id` IS NOT NULL DO
        SELECT `RW`, true INTO `result`, `overridden`
        FROM `my_config_tree_group` tg
        WHERE tg.`NodeID` = `node_id` AND tg.`GroupID` = `group_id`
        ORDER BY `RW` DESC
        LIMIT 1;

        IF `overridden` THEN
            RETURN `result`;
        END IF;

        SELECT `ParentID` INTO `node_id` FROM `my_config_tree` WHERE `ID` = `node_id`;
    END WHILE;

    RETURN `result`;
END;$$

delimiter ;

INSERT INTO `my_config_tree` (`Name`) VALUES ('');
INSERT INTO `my_config_tree_log` (`NodeID`, `Version`, `Value`, `ContentType`, `Author`, `MTime`, `Comment`, `Deleted`)
SELECT `ID`, `Version`, `Value`, `ContentType`, 'onlineconf', `MTime`, NULL, `Deleted` FROM `my_config_tree` WHERE `Name` = '';

ALTER TABLE `my_config_activity` ADD COLUMN `Online` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `my_config_activity` ADD COLUMN `Package` varchar(32);
