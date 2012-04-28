drop function `my_config_tree_access`;
drop function `my_config_tree_group_access`;

alter table `my_config_tree` add column `Notification` ENUM('none', 'no-value', 'with-value');

delimiter $$

CREATE FUNCTION `my_config_tree_notification` (`node_id` bigint(20) unsigned) RETURNS enum('none', 'no-value', 'with-value')
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

CREATE FUNCTION `my_config_tree_access` (`node_id` bigint(20) unsigned, `username` varchar(256)) RETURNS boolean
READS SQL DATA
BEGIN
    DECLARE `result` boolean;
    DECLARE `group_id` int(11);
    DECLARE `rw` boolean;
    DECLARE `done` boolean;
    DECLARE cur CURSOR FOR
        SELECT tg.`GroupID`, tg.`RW`
        FROM `my_config_tree_group` tg
        JOIN `my_config_user_group` ug ON ug.`GroupID` = tg.`GroupID`
        WHERE tg.`NodeID` = `node_id` AND ug.`User` = `username`;

    CREATE TEMPORARY TABLE `my_config_tree_access_tmp` (`GroupID` int(11) NOT NULL UNIQUE, `RW` boolean) ENGINE=MEMORY;

    WHILE `node_id` IS NOT NULL DO
        SET done = false;
        OPEN cur;
        cur_loop: LOOP
            BEGIN
                DECLARE EXIT HANDLER FOR NOT FOUND BEGIN SET done = true; END;
                FETCH cur INTO `group_id`, `rw`;
            END;
            IF done THEN
                LEAVE cur_loop;
            END IF;
            BEGIN
                DECLARE EXIT HANDLER FOR SQLSTATE '23000' BEGIN END;
                INSERT INTO `my_config_tree_access_tmp`
                VALUES (`group_id`, `rw`);
            END;
        END LOOP;
        CLOSE cur;

        SELECT `ParentID` INTO `node_id` FROM `my_config_tree` WHERE `ID` = `node_id`;
    END WHILE;

    BEGIN
        DECLARE CONTINUE HANDLER FOR SQLSTATE '02000' BEGIN END;

        SELECT tmp.`RW` INTO `result`
        FROM `my_config_tree_access_tmp` tmp
        ORDER BY tmp.`RW` DESC
        LIMIT 1;
    END;

    DROP TEMPORARY TABLE IF EXISTS `my_config_tree_access_tmp`;

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

delimiter ;
