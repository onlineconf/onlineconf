DELETE FROM `my_config_tree_group`;
DELETE FROM `my_config_tree_log`;
DELETE FROM `my_config_tree` ORDER BY `ID` DESC;
DELETE FROM `my_config_group`;

INSERT INTO `my_config_group` VALUES (1,'root');
INSERT INTO `my_config_group` VALUES (2,'sysadmin');
INSERT INTO `my_config_group` VALUES (3,'head-developer');
INSERT INTO `my_config_group` VALUES (4,'back-developer');
INSERT INTO `my_config_group` VALUES (5,'web-developer');
INSERT INTO `my_config_group` VALUES (6,'mobile-developer');

INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (1,'',NULL,'/',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,'none');
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (2,'onlineconf',1,'/onlineconf',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (3,'module',2,'/onlineconf/module','[]','application/x-case','','',1,'2019-08-05 17:39:43',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (4,'service',2,'/onlineconf/service',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (5,'chroot',2,'/onlineconf/chroot',NULL,'application/x-null','','',1,'2019-08-05 17:17:50',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (6,'infrastructure',1,'/infrastructure',NULL,'application/x-null','Sysadmin stuff','All IP addresses, hosts, credentials and offer infrastructure information should live here.\nParameters from this hierarchy should not be used directly but through symlinks in project subtrees only.',1,'2019-08-05 17:24:45',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (7,'group',2,'/onlineconf/group',NULL,'application/x-null','','',1,'2019-08-05 18:05:39',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (8,'ephemeral-ip',2,'/onlineconf/ephemeral-ip','172.0.0.0/8','text/plain','','',1,'2022-10-24 18:15:21',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (9,'services',1,'/services',NULL,'application/x-null','','',1,'2022-10-24 18:16:21',0,NULL);
INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (10,'secure',1,'/secure',NULL,'application/x-null','','',1,'2022-10-24 18:17:21',0,NULL);

INSERT INTO `my_config_tree_group` VALUES (1,1,1);
INSERT INTO `my_config_tree_group` VALUES (3,3,0);
INSERT INTO `my_config_tree_group` VALUES (6,2,1);
INSERT INTO `my_config_tree_group` VALUES (9,3,1);
INSERT INTO `my_config_tree_group` VALUES (9,2,1);
INSERT INTO `my_config_tree_group` VALUES (9,4,0);
INSERT INTO `my_config_tree_group` VALUES (9,5,0);
INSERT INTO `my_config_tree_group` VALUES (10,2,1);

INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (1,1,'/',1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (2,2,'/onlineconf',1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (3,3,'/onlineconf/module',1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (4,4,'/onlineconf/service',1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (5,5,'/onlineconf/chroot',1,NULL,'application/x-null','admin','2019-08-05 17:17:50','Initialize chroot',0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (6,6,'/infrastructure',1,NULL,'application/x-null','admin','2019-08-05 17:24:45','Init infra',0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (7,7,'/onlineconf/group',1,NULL,'application/x-null','admin','2019-08-05 18:05:39','Init infra',0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (8,8,'/onlineconf/ephemeral-ip',1,'172.0.0.0/8','text/plain','admin','2022-10-24 18:15:21','',0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (9,9,'/services',1,NULL,'application/x-null','admin','2022-10-24 18:16:21','',0);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (10,10,'/secure',1,NULL,'application/x-null','admin','2022-10-24 18:17:21','',0);

--
-- Dumping data for table `my_config_user_group`
--

INSERT INTO `my_config_user_group` VALUES ('admin',1);

--
-- Value encryption is enabled out of the box (see schema.sql); a key file is
-- still required on every onlineconf-admin instance before encrypted values
-- can be written.
--

INSERT INTO `my_config_tree` (`ID`,`Name`,`ParentID`,`Path`,`Value`,`ContentType`,`Summary`,`Description`,`Version`,`MTime`,`Deleted`,`Notification`) VALUES (11,'value-encryption',2,'/onlineconf/value-encryption','1','text/plain','','',1,'2026-11-10 00:00:00',0,NULL);
INSERT INTO `my_config_tree_log` (`ID`,`NodeID`,`Path`,`Version`,`Value`,`ContentType`,`Author`,`MTime`,`Comment`,`Deleted`) VALUES (11,11,'/onlineconf/value-encryption',1,'1','text/plain','onlineconf','2026-11-10 00:00:00','Enable value encryption',0);
