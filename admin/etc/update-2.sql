drop function `my_config_tree_access`;

delimiter $$

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

delimiter ;

-- vim:et:sw=4:sts=4:ts=4:ai:si
