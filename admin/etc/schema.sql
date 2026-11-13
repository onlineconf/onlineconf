CREATE TABLE `my_config_tree` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `Name` varchar(256) character set ascii collate ascii_bin NOT NULL,
    `ParentID` bigint(20) unsigned default NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL,
    `Value` mediumtext character set utf8mb4 collate utf8mb4_bin,
    `ContentType` varchar(255) character set ascii NOT NULL default 'application/x-null',
    `Encrypted` tinyint(1) NOT NULL default '0',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `my_config_tree_log` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `NodeID` bigint(20) unsigned NOT NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL default '',
    `Version` int(11) NOT NULL,
    `Value` mediumtext character set utf8mb4 collate utf8mb4_bin,
    `ContentType` varchar(255) character set ascii NOT NULL default 'application/x-null',
    `Encrypted` tinyint(1) NOT NULL default '0',
    `Author` varchar(128) character set ascii NOT NULL,
    `MTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Comment` varchar(512) default NULL,
    `Deleted` tinyint(1) NOT NULL default '0',
    `Silent` tinyint(1) NOT NULL default '0',
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `NodeID` (`NodeID`,`Version`),
    KEY `MTime` (`MTime`),
    KEY `Path` (`Path`),
    CONSTRAINT `my_config_tree_log_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `my_config_group` (
    `ID` int(11) NOT NULL auto_increment,
    `Name` varchar(128) NOT NULL,
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `Name` (`Name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `my_config_user_group` (
    `User` varchar(128) character set ascii NOT NULL,
    `GroupID` int(11) NOT NULL,
    UNIQUE KEY `User_Group` (`User`,`GroupID`),
    KEY `User` (`User`),
    KEY `GroupID` (`GroupID`),
    CONSTRAINT `my_config_user_group_ibfk_1` FOREIGN KEY (`GroupID`) REFERENCES `my_config_group` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `my_config_tree_group` (
    `NodeID` bigint(20) unsigned NOT NULL,
    `GroupID` int(11) NOT NULL,
    `RW` tinyint(1) default '0',
    PRIMARY KEY  (`NodeID`,`GroupID`),
    KEY `GroupID` (`GroupID`),
    CONSTRAINT `my_config_tree_group_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`) ON DELETE CASCADE,
    CONSTRAINT `my_config_tree_group_ibfk_2` FOREIGN KEY (`GroupID`) REFERENCES `my_config_group` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `my_config_activity` (
    `Host` varchar(255) character set ascii NOT NULL,
    `Time` timestamp NULL default NULL,
    `Online` timestamp NOT NULL default CURRENT_TIMESTAMP,
    `Package` varchar(32) NOT NULL,
    PRIMARY KEY  (`Host`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Registry of the value encryption keys ever loaded: no key material, only a
-- verifier (the fixed registry constant sealed under the key) that lets an
-- instance prove its key file holds the material registered under an id, and
-- the Active mark the stored data is encrypted under once a rotation sweep
-- completed. Seq orders keys by registration.
CREATE TABLE `my_config_encryption_key` (
    `KeyID` varchar(32) character set ascii NOT NULL,
    `Seq` int(11) unsigned NOT NULL auto_increment,
    `Verifier` varbinary(64) NOT NULL,
    `Active` tinyint(1) NOT NULL default '0',
    `CreatedAt` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`KeyID`),
    UNIQUE KEY `Seq` (`Seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- The audit log: acts worth telling a parameter's owners about that are not
-- changes of the parameter itself. Today these are the reveals of encrypted
-- values — reveal_view: the current value was shown, to view or to edit it;
-- reveal_history: the parameter's history was shown — written before any
-- decryption happens and served to bots by the botapi audit feed. A later
-- kind is appended at the END of the Kind enum by its own update-N.sql:
-- MySQL applies that as a metadata-only change, whereas reordering or
-- removing members rebuilds the table. NodeID is NULL for an event that is
-- not bound to a parameter.
CREATE TABLE `my_config_audit` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `NodeID` bigint(20) unsigned default NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL default '',
    `User` varchar(128) character set ascii NOT NULL,
    `Kind` enum('reveal_view','reveal_history') character set ascii NOT NULL,
    `Time` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`ID`),
    KEY `NodeID` (`NodeID`),
    KEY `Time` (`Time`),
    CONSTRAINT `my_config_audit_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- The schema version: the number of the last admin/etc/update-N.sql applied.
-- Every update file ends with INSERT INTO my_config_schema (Version) VALUES
-- (N), so that onlineconf-admin refuses to start against a schema behind the
-- one it is written for, naming the files to apply.
CREATE TABLE `my_config_schema` (
    `Version` int(11) unsigned NOT NULL,
    `AppliedAt` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `my_config_schema` (`Version`) VALUES (8);

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

-- /onlineconf/value-encryption is seeded enabled: a fresh installation may
-- store encrypted values out of the box (once a key file is configured). On
-- upgraded installations the parameter does not exist and the feature stays
-- off — applying the schema migration alone changes nothing — until the
-- operator creates it with "1" after every instance runs a version that
-- understands the Encrypted marker and has the key file.
INSERT INTO `my_config_tree` (`ID`, `Name`, `ParentID`, `Value`, `ContentType`, `Notification`) VALUES
    (1, '', NULL, NULL, 'application/x-null', 'none'),
    (2, 'onlineconf', 1, NULL, 'application/x-null', NULL),
    (3, 'module', 2, NULL, 'application/x-null', NULL),
    (4, 'service', 2, NULL, 'application/x-null', NULL),
    (5, 'value-encryption', 2, '1', 'text/plain', NULL);

INSERT INTO `my_config_tree_log` (`NodeID`, `Path`, `Version`, `Value`, `ContentType`, `Encrypted`, `Author`, `MTime`, `Comment`, `Deleted`)
SELECT `ID`, `Path`, `Version`, `Value`, `ContentType`, `Encrypted`, 'onlineconf', `MTime`, 'Init onlineconf', `Deleted` FROM `my_config_tree` ORDER BY `ID`;

INSERT INTO `my_config_group` (`Name`) VALUES ('root');

INSERT INTO `my_config_tree_group` (`NodeID`, `GroupID`, `RW`)
VALUES (
    (SELECT `ID` from `my_config_tree` WHERE `Path` = '/'),
    (SELECT `ID` from `my_config_group` WHERE `Name` = 'root'),
    true
);
