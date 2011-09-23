CREATE TABLE `my_config` (
    `Version` int(11) NOT NULL,
    `Module` int(11) unsigned NOT NULL,
    `Key` varchar(256) NOT NULL,
    `Value` text,
    `Flags` int(11) default '0',
    `Access` int(11) default '1',
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
    `Access` int(11) default '1',
    `Comment` varchar(256) default NULL,
    KEY `Module` (`Module`,`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;

CREATE TABLE `my_config_activity` (
    `Host` varchar(256) NOT NULL,
    `Time` datetime NOT NULL,
    PRIMARY KEY `Host` (`Host`)
) ENGINE=InnoDB DEFAULT CHARSET=cp1251;


insert into my_config_module (Name,Version,Comment) VALUES ('@OVERLOAD',0,'Overloaded config params');
update my_config_module set ID=2147483647 where Name='@OVERLOAD';

insert into my_config_module (Name,Version,Comment) VALUES ('@SELFTEST',0,'OnlineConf Sefl Test');
update my_config_module set ID=2147483646 where Name='@SELFTEST';
