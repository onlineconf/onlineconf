CREATE TABLE `my_config` (
    `Version` int(11) NOT NULL,
    `Module` int(11) unsigned NOT NULL,
    `Key` varchar(256) NOT NULL,
    `Value` text,
    `Flags` int(11) default '0',
    `Access` int(11) default '1',
    UNIQUE KEY `Module` (`Module`,`Key`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_module` (
    `ID` int(11) NOT NULL auto_increment,
    `Name` varchar(64) NOT NULL,
    `Version` int(11) NOT NULL default '0',
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

REATE TABLE `my_config_log` (
    `Version` int(11) NOT NULL,
    `Module` int(11) unsigned NOT NULL,
    `Key` varchar(256) NOT NULL,
    `Value` text,
    `Flags` int(11) default '0',
    `Access` int(11) default '1',
    KEY `Module` (`Module`,`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;
