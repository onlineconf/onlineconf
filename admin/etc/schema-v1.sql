CREATE TABLE `my_config` (
    `Version` int(11) NOT NULL,
    `Module` int(11) unsigned NOT NULL,
    `Key` varchar(256) NOT NULL,
    `Value` text,
    `Flags` int(11) default '0',
    `Comment` varchar(256) default NULL,
    UNIQUE KEY `Module` (`Module`,`Key`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_module` (
    `ID` int(11) NOT NULL auto_increment,
    `Name` varchar(64) NOT NULL,
    `Version` int(11) NOT NULL default '0',
    `Comment` text default NULL,
    PRIMARY KEY  (`ID`),
    UNIQUE KEY `Name` (`Name`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_transaction` (
    `Version` int(11) NOT NULL,
    `Module` int(11) NOT NULL,
    `Comment` varchar(256) default NULL,
    `ChangedBy` varchar(128) NOT NULL,
    `ChangedTime` timestamp NOT NULL default CURRENT_TIMESTAMP,
    UNIQUE KEY `Module` (`Module`,`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_log` (
    `Version` int(11) NOT NULL,
    `Module` int(11) unsigned NOT NULL,
    `Key` varchar(256) NOT NULL,
    `Value` text,
    `Flags` int(11) default '0',
    `Comment` varchar(256) default NULL,
    KEY `Module` (`Module`,`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_activity` (
    `Host` varchar(256) NOT NULL,
    `Time` datetime NOT NULL,
    PRIMARY KEY `Host` (`Host`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_group` (
    `ID` int(11) NOT NULL auto_increment,
    `Name` varchar(256) NOT NULL,
    PRIMARY KEY `ID` (`ID`),
    UNIQUE KEY `Name` (`Name`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_user_group` (
    `User` varchar(128) NOT NULL,
    `GroupID` int(11) NOT NULL,
    UNIQUE KEY `User_Group` (`User`,`GroupID`),
    KEY `User` (`User`),
    KEY `GroupID` (`GroupID`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_module_group` (
    `ModuleID` int(11) NOT NULL,
    `GroupID` int(11) NOT NULL,
    UNIQUE KEY `Module_Group` (`ModuleID`,`GroupID`),
    KEY `ModuleID` (`ModuleID`),
    KEY `GroupID` (`GroupID`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

INSERT INTO `my_config_module` (`Name,`Version`,`Comment`) VALUES ('@OVERLOAD',0,'Overloaded config params');
UPDATE `my_config_module` SET `ID`=2147483647 WHERE `Name`='@OVERLOAD';

INSERT INTO `my_config_module` (`Name`,`Version`,`Comment`) VALUES ('@SELFTEST',0,'OnlineConf Self Test');
UPDATE `my_config_module` SET `ID`=2147483646 WHERE `Name`='@SELFTEST';

INSERT INTO `my_config_group` (`Name`) VALUES ('root');
UPDATE `my_config_group` SET `ID`=2147483647 WHERE `Name`='root';
