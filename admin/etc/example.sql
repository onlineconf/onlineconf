DELETE FROM `my_config_tree_group`;
DELETE FROM `my_config_tree_log`;
DELETE FROM `my_config_tree` ORDER BY `ID` DESC;
DELETE FROM `my_config_group`;

--
-- Dumping data for table `mod_auth`
--

INSERT INTO `mod_auth` VALUES ('admin','hCG7NSubpIe8M');
INSERT INTO `mod_auth` VALUES ('beaver','iJdvMl0D40I6M');
INSERT INTO `mod_auth` VALUES ('hedgehog','6nLwikib1PMAo');
INSERT INTO `mod_auth` VALUES ('meerkat','uGruJjKihJ7oI');
INSERT INTO `mod_auth` VALUES ('rabbit','JhDoIGtWEfyrc');

--
-- Dumping data for table `my_config_group`
--

INSERT INTO `my_config_group` VALUES (1,'root');
INSERT INTO `my_config_group` VALUES (2,'sysadmin');
INSERT INTO `my_config_group` VALUES (3,'gopher-developer');
INSERT INTO `my_config_group` VALUES (4,'squirrel-developer');
INSERT INTO `my_config_group` VALUES (5,'squirrel-manager');

--
-- Dumping data for table `my_config_tree`
--

INSERT INTO `my_config_tree` VALUES (1,'',NULL,'/',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,'none');
INSERT INTO `my_config_tree` VALUES (2,'onlineconf',1,'/onlineconf',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,NULL);
INSERT INTO `my_config_tree` VALUES (3,'module',2,'/onlineconf/module','[{\"mime\":\"application/x-symlink\",\"value\":\"/onlineconf/chroot/gopher\",\"service\":\"gopher\"},{\"mime\":\"application/x-symlink\",\"value\":\"/onlineconf/chroot/squirrel\",\"service\":\"squirrel\"}]','application/x-case','','',2,'2019-08-05 17:39:43',0,NULL);
INSERT INTO `my_config_tree` VALUES (4,'service',2,'/onlineconf/service',NULL,'application/x-null','','',1,'2019-08-05 17:01:16',0,NULL);
INSERT INTO `my_config_tree` VALUES (5,'gopher',4,'/onlineconf/service/gopher','9cc1ee455a3363ffc504f40006f70d0c8276648a5d3eb3f9524e94d1b7a83aef','text/plain','','',1,'2019-08-05 17:13:00',0,NULL);
INSERT INTO `my_config_tree` VALUES (6,'squirrel',4,'/onlineconf/service/squirrel','960a38ace75a2fe9426aa5d48b536cc6db18a2023b9bdd698e562fc3023585a6','text/plain','','',1,'2019-08-05 17:13:35',0,NULL);
INSERT INTO `my_config_tree` VALUES (7,'gopher',1,'/gopher',NULL,'application/x-null','Gopher project','',1,'2019-08-05 17:17:01',0,NULL);
INSERT INTO `my_config_tree` VALUES (8,'squirrel',1,'/squirrel',NULL,'application/x-null','Squirrel project','',1,'2019-08-05 17:17:16',0,NULL);
INSERT INTO `my_config_tree` VALUES (9,'chroot',2,'/onlineconf/chroot',NULL,'application/x-null','','',1,'2019-08-05 17:17:50',0,NULL);
INSERT INTO `my_config_tree` VALUES (10,'gopher',9,'/onlineconf/chroot/gopher','delimiter: /','application/x-yaml','','',1,'2019-08-05 17:18:12',0,NULL);
INSERT INTO `my_config_tree` VALUES (11,'TREE',10,'/onlineconf/chroot/gopher/TREE',NULL,'application/x-null','','',1,'2019-08-05 17:19:38',0,NULL);
INSERT INTO `my_config_tree` VALUES (12,'gopher',11,'/onlineconf/chroot/gopher/TREE/gopher','/gopher','application/x-symlink','','',1,'2019-08-05 17:20:15',0,NULL);
INSERT INTO `my_config_tree` VALUES (13,'infrastructure',1,'/infrastructure',NULL,'application/x-null','Sysadmin stuff','All IP addresses, hosts, credentials and offer infrastructure information should live here.\nParameters from this hierarchy should not be used directly but through symlinks in project subtrees only.',1,'2019-08-05 17:24:45',0,NULL);
INSERT INTO `my_config_tree` VALUES (14,'postgresql',13,'/infrastructure/postgresql',NULL,'application/x-null','','',1,'2019-08-05 17:25:54',0,NULL);
INSERT INTO `my_config_tree` VALUES (15,'tarantool',13,'/infrastructure/tarantool',NULL,'application/x-null','','',1,'2019-08-05 17:26:04',0,NULL);
INSERT INTO `my_config_tree` VALUES (16,'gopher-main',14,'/infrastructure/postgresql/gopher-main',NULL,'application/x-null','Main gopher database','',1,'2019-08-05 17:27:13',0,NULL);
INSERT INTO `my_config_tree` VALUES (17,'host',16,'/infrastructure/postgresql/gopher-main/host','db1.gopher.example.com','text/plain','','',1,'2019-08-05 17:27:53',0,NULL);
INSERT INTO `my_config_tree` VALUES (18,'user',16,'/infrastructure/postgresql/gopher-main/user','gopher','text/plain','','',1,'2019-08-05 17:28:13',0,NULL);
INSERT INTO `my_config_tree` VALUES (19,'password',16,'/infrastructure/postgresql/gopher-main/password','gopher-gopher-gopher-gopher','text/plain','','',1,'2019-08-05 17:28:41',0,NULL);
INSERT INTO `my_config_tree` VALUES (20,'base',16,'/infrastructure/postgresql/gopher-main/base','gopher','text/plain','','',1,'2019-08-05 17:29:01',0,NULL);
INSERT INTO `my_config_tree` VALUES (21,'gopher-stat',14,'/infrastructure/postgresql/gopher-stat',NULL,'application/x-null','Gopher statistics','',1,'2019-08-05 17:30:18',0,NULL);
INSERT INTO `my_config_tree` VALUES (22,'base',21,'/infrastructure/postgresql/gopher-stat/base','gopher_stat','text/plain','','',1,'2019-08-05 17:30:51',0,NULL);
INSERT INTO `my_config_tree` VALUES (23,'host',21,'/infrastructure/postgresql/gopher-stat/host','statdb1.gopher.example.com','text/plain','','',1,'2019-08-05 17:31:15',0,NULL);
INSERT INTO `my_config_tree` VALUES (24,'user',21,'/infrastructure/postgresql/gopher-stat/user','gopher','text/plain','','',1,'2019-08-05 17:31:30',0,NULL);
INSERT INTO `my_config_tree` VALUES (25,'password',21,'/infrastructure/postgresql/gopher-stat/password','gopher-gopher-gopher-gopher','text/plain','','',1,'2019-08-05 17:32:03',0,NULL);
INSERT INTO `my_config_tree` VALUES (26,'gopher-user',15,'/infrastructure/tarantool/gopher-user','userkvs1.gopher.example.com','text/plain','','',1,'2019-08-05 17:33:21',0,NULL);
INSERT INTO `my_config_tree` VALUES (27,'replica',26,'/infrastructure/tarantool/gopher-user/replica','userkvs2.gopher.example.com','text/plain','','',1,'2019-08-05 17:33:56',0,NULL);
INSERT INTO `my_config_tree` VALUES (28,'user',7,'/gopher/user',NULL,'application/x-null','','',1,'2019-08-05 17:34:39',0,NULL);
INSERT INTO `my_config_tree` VALUES (29,'box',28,'/gopher/user/box','/infrastructure/tarantool/gopher-user','application/x-symlink','','',1,'2019-08-05 17:35:34',0,NULL);
INSERT INTO `my_config_tree` VALUES (30,'statistics',7,'/gopher/statistics',NULL,'application/x-null','','',1,'2019-08-05 17:36:02',0,NULL);
INSERT INTO `my_config_tree` VALUES (31,'database',30,'/gopher/statistics/database','/infrastructure/postgresql/gopher-stat','application/x-symlink','','',1,'2019-08-05 17:36:24',0,NULL);
INSERT INTO `my_config_tree` VALUES (32,'squirrel',9,'/onlineconf/chroot/squirrel','delimiter: /','application/x-yaml','','',1,'2019-08-05 17:38:49',0,NULL);
INSERT INTO `my_config_tree` VALUES (33,'graphite',13,'/infrastructure/graphite',NULL,'application/x-null','','',1,'2019-08-05 17:50:37',0,NULL);
INSERT INTO `my_config_tree` VALUES (34,'host',33,'/infrastructure/graphite/host','graphite1.example.com','text/plain','','',1,'2019-08-05 17:51:06',0,NULL);
INSERT INTO `my_config_tree` VALUES (35,'port',33,'/infrastructure/graphite/port',NULL,'application/x-null','','',1,'2019-08-05 17:51:19',0,NULL);
INSERT INTO `my_config_tree` VALUES (36,'carbon',35,'/infrastructure/graphite/port/carbon','2003','text/plain','','',1,'2019-08-05 17:51:53',0,NULL);
INSERT INTO `my_config_tree` VALUES (37,'carbide',35,'/infrastructure/graphite/port/carbide','2004','text/plain','','',1,'2019-08-05 17:52:07',0,NULL);
INSERT INTO `my_config_tree` VALUES (38,'graphite',30,'/gopher/statistics/graphite','${/infrastructure/graphite/host}:${/infrastructure/graphite/port/carbide}','application/x-template','','',1,'2019-08-05 17:53:12',0,NULL);
INSERT INTO `my_config_tree` VALUES (39,'sms-verification',28,'/gopher/user/sms-verification','[{\"mime\":\"text/plain\",\"value\":\"1\"},{\"mime\":\"text/plain\",\"value\":\"0\",\"group\":\"gopher-alpha\"}]','application/x-case','Require SMS verification','',2,'2019-08-05 18:06:54',0,NULL);
INSERT INTO `my_config_tree` VALUES (40,'group',2,'/onlineconf/group',NULL,'application/x-null','','',1,'2019-08-05 18:05:39',0,NULL);
INSERT INTO `my_config_tree` VALUES (41,'gopher-alpha',40,'/onlineconf/group/gopher-alpha','alpha*.gopher.example.com','text/plain','Gopher Alpha','',1,'2019-08-05 18:06:18',0,NULL);
INSERT INTO `my_config_tree` VALUES (42,'content',7,'/gopher/content',NULL,'application/x-null','','',1,'2019-08-05 18:08:26',0,NULL);
INSERT INTO `my_config_tree` VALUES (43,'database',42,'/gopher/content/database','/infrastructure/postgresql/gopher-main','application/x-symlink','','',1,'2019-08-05 18:08:44',0,NULL);
INSERT INTO `my_config_tree` VALUES (44,'gopherizator-percent',42,'/gopher/content/gopherizator-percent','42','text/plain','','',1,'2019-08-05 18:10:44',0,NULL);
INSERT INTO `my_config_tree` VALUES (45,'squirrel-user',15,'/infrastructure/tarantool/squirrel-user','userkvs1.squirrel.example.com','text/plain','','',1,'2019-08-05 18:15:40',0,NULL);
INSERT INTO `my_config_tree` VALUES (46,'replica',45,'/infrastructure/tarantool/squirrel-user/replica','userkvs2.squirrel.example.com','text/plain','','',1,'2019-08-05 18:16:04',0,NULL);
INSERT INTO `my_config_tree` VALUES (47,'user',8,'/squirrel/user',NULL,'application/x-null','','',1,'2019-08-05 18:16:50',0,NULL);
INSERT INTO `my_config_tree` VALUES (48,'box',47,'/squirrel/user/box','/infrastructure/tarantool/squirrel-user','application/x-symlink','','',1,'2019-08-05 18:17:09',0,NULL);
INSERT INTO `my_config_tree` VALUES (49,'statistics',8,'/squirrel/statistics',NULL,'application/x-null','','',1,'2019-08-05 18:18:14',0,NULL);
INSERT INTO `my_config_tree` VALUES (50,'graphite',49,'/squirrel/statistics/graphite','${/infrastructure/graphite/host}:${/infrastructure/graphite/port/carbide}','application/x-template','','',2,'2019-08-05 18:30:28',0,NULL);
INSERT INTO `my_config_tree` VALUES (51,'nut',8,'/squirrel/nut',NULL,'application/x-null','','',1,'2019-08-05 18:19:37',0,NULL);
INSERT INTO `my_config_tree` VALUES (52,'mashroom',8,'/squirrel/mashroom',NULL,'application/x-null','','',1,'2019-08-05 18:19:53',0,NULL);
INSERT INTO `my_config_tree` VALUES (54,'collect-huzelnut',51,'/squirrel/nut/collect-huzelnut','1','text/plain','','',1,'2019-08-05 18:21:32',0,NULL);
INSERT INTO `my_config_tree` VALUES (55,'collect-amanita',52,'/squirrel/mashroom/collect-amanita','0','text/plain','','',1,'2019-08-05 18:22:22',0,NULL);
INSERT INTO `my_config_tree` VALUES (56,'collect-boletus',52,'/squirrel/mashroom/collect-boletus','1','text/plain','','',1,'2019-08-05 18:23:36',0,NULL);
INSERT INTO `my_config_tree` VALUES (57,'red',32,'/onlineconf/chroot/squirrel/red',NULL,'application/x-null','','',1,'2019-08-05 18:24:37',0,NULL);
INSERT INTO `my_config_tree` VALUES (58,'gray',32,'/onlineconf/chroot/squirrel/gray',NULL,'application/x-null','','',1,'2019-08-05 18:26:29',0,NULL);
INSERT INTO `my_config_tree` VALUES (59,'user',58,'/onlineconf/chroot/squirrel/gray/user','/squirrel/user','application/x-symlink','','',1,'2019-08-05 18:26:50',0,NULL);
INSERT INTO `my_config_tree` VALUES (60,'mashroom',58,'/onlineconf/chroot/squirrel/gray/mashroom','/squirrel/mashroom','application/x-symlink','','',1,'2019-08-05 18:27:18',0,NULL);
INSERT INTO `my_config_tree` VALUES (61,'user',57,'/onlineconf/chroot/squirrel/red/user','/squirrel/user','application/x-symlink','','',1,'2019-08-05 18:27:34',0,NULL);
INSERT INTO `my_config_tree` VALUES (62,'statistics',57,'/onlineconf/chroot/squirrel/red/statistics','/squirrel/statistics','application/x-symlink','','',1,'2019-08-05 18:27:55',0,NULL);
INSERT INTO `my_config_tree` VALUES (63,'nut',57,'/onlineconf/chroot/squirrel/red/nut','/squirrel/nut','application/x-symlink','','',1,'2019-08-05 18:28:11',0,NULL);
INSERT INTO `my_config_tree` VALUES (64,'promocode',8,'/squirrel/promocode','{\n  \"BELKA5\": {\n    \"action\": \"discount\",\n    \"percent\": 5,\n    \"from\": \"2019-09-01\",\n    \"till\": \"2019-12-01\"\n  },\n  \"BELKA15\": {\n    \"premiumCard\": true,\n    \"action\": \"discount\",\n    \"percent\": 15,\n    \"from\": \"2019-09-01\",\n    \"till\": \"2019-12-01\"\n  },\n  \"STRELKA200\": {\n    \"action\": \"bonus\",\n    \"value\": 200,\n    \"from\": \"2019-01-01\",\n    \"till\": \"2020-01-01\"\n  }\n}','application/json','Active promocodes','',1,'2019-09-16 16:23:11',0,NULL);
INSERT INTO `my_config_tree` VALUES (65,'promocode',58,'/onlineconf/chroot/squirrel/gray/promocode','/squirrel/promocode','application/x-symlink','','',1,'2019-09-16 16:26:42',0,NULL);
INSERT INTO `my_config_tree` VALUES (66,'ephemeral-ip',2,'/onlineconf/ephemeral-ip','172.0.0.0/8','text/plain','','',1,'2022-10-24 18:15:21',0,NULL);

--
-- Dumping data for table `my_config_tree_group`
--

INSERT INTO `my_config_tree_group` VALUES (1,1,1);
INSERT INTO `my_config_tree_group` VALUES (3,3,0);
INSERT INTO `my_config_tree_group` VALUES (3,4,0);
INSERT INTO `my_config_tree_group` VALUES (7,3,1);
INSERT INTO `my_config_tree_group` VALUES (8,4,1);
INSERT INTO `my_config_tree_group` VALUES (10,3,1);
INSERT INTO `my_config_tree_group` VALUES (13,2,1);
INSERT INTO `my_config_tree_group` VALUES (32,4,1);
INSERT INTO `my_config_tree_group` VALUES (51,5,1);
INSERT INTO `my_config_tree_group` VALUES (52,5,1);
INSERT INTO `my_config_tree_group` VALUES (64,5,1);

--
-- Dumping data for table `my_config_tree_log`
--

INSERT INTO `my_config_tree_log` VALUES (1,1,1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` VALUES (2,2,1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` VALUES (3,3,1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` VALUES (4,4,1,NULL,'application/x-null','onlineconf','2019-08-05 17:01:16',NULL,0);
INSERT INTO `my_config_tree_log` VALUES (5,5,1,'9cc1ee455a3363ffc504f40006f70d0c8276648a5d3eb3f9524e94d1b7a83aef','text/plain','admin','2019-08-05 17:13:00','Add account for gopher service',0);
INSERT INTO `my_config_tree_log` VALUES (6,6,1,'960a38ace75a2fe9426aa5d48b536cc6db18a2023b9bdd698e562fc3023585a6','text/plain','admin','2019-08-05 17:13:35','Add account for squirrel service',0);
INSERT INTO `my_config_tree_log` VALUES (7,7,1,NULL,'application/x-null','admin','2019-08-05 17:17:01','Initialize gopher',0);
INSERT INTO `my_config_tree_log` VALUES (8,8,1,NULL,'application/x-null','admin','2019-08-05 17:17:16','Squirrel project root',0);
INSERT INTO `my_config_tree_log` VALUES (9,9,1,NULL,'application/x-null','admin','2019-08-05 17:17:50','Initialize chroot',0);
INSERT INTO `my_config_tree_log` VALUES (10,10,1,'delimiter: /','application/x-yaml','admin','2019-08-05 17:18:12','Initialize gopher',0);
INSERT INTO `my_config_tree_log` VALUES (11,11,1,NULL,'application/x-null','admin','2019-08-05 17:19:38','Initialize gopher',0);
INSERT INTO `my_config_tree_log` VALUES (12,12,1,'/gopher','application/x-symlink','admin','2019-08-05 17:20:15','Initialize gopher',0);
INSERT INTO `my_config_tree_log` VALUES (13,13,1,NULL,'application/x-null','admin','2019-08-05 17:24:45','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (14,14,1,NULL,'application/x-null','admin','2019-08-05 17:25:54','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (15,15,1,NULL,'application/x-null','admin','2019-08-05 17:26:04','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (16,16,1,NULL,'application/x-null','admin','2019-08-05 17:27:13','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (17,17,1,'db1.gopher.example.com','text/plain','admin','2019-08-05 17:27:53','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (18,18,1,'gopher','text/plain','admin','2019-08-05 17:28:13','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (19,19,1,'gopher-gopher-gopher-gopher','text/plain','admin','2019-08-05 17:28:41','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (20,20,1,'gopher','text/plain','admin','2019-08-05 17:29:01','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (21,21,1,NULL,'application/x-null','admin','2019-08-05 17:30:18','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (22,22,1,'gopher_stat','text/plain','admin','2019-08-05 17:30:51','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (23,23,1,'statdb1.gopher.example.com','text/plain','admin','2019-08-05 17:31:15','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (24,24,1,'gopher','text/plain','admin','2019-08-05 17:31:30','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (25,25,1,'gopher-gopher-gopher-gopher','text/plain','admin','2019-08-05 17:32:03','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (26,26,1,'userkvs1.gopher.example.com','text/plain','admin','2019-08-05 17:33:21','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (27,27,1,'userkvs2.gopher.example.com','text/plain','admin','2019-08-05 17:33:56','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (28,28,1,NULL,'application/x-null','admin','2019-08-05 17:34:39','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (29,29,1,'/infrastructure/tarantool/gopher-user','application/x-symlink','admin','2019-08-05 17:35:34','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (30,30,1,NULL,'application/x-null','admin','2019-08-05 17:36:02','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (31,31,1,'/infrastructure/postgresql/gopher-stat','application/x-symlink','admin','2019-08-05 17:36:24','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (32,32,1,'delimiter: /','application/x-yaml','admin','2019-08-05 17:38:49','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (33,3,2,'[{\"mime\":\"application/x-symlink\",\"value\":\"/onlineconf/chroot/gopher\",\"service\":\"gopher\"},{\"mime\":\"application/x-symlink\",\"value\":\"/onlineconf/chroot/squirrel\",\"service\":\"squirrel\"}]','application/x-case','admin','2019-08-05 17:39:43','Init chrooted module',0);
INSERT INTO `my_config_tree_log` VALUES (34,33,1,NULL,'application/x-null','admin','2019-08-05 17:50:37','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (35,34,1,'graphite1.example.com','text/plain','admin','2019-08-05 17:51:06','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (36,35,1,NULL,'application/x-null','admin','2019-08-05 17:51:19','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (37,36,1,'2003','text/plain','admin','2019-08-05 17:51:53','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (38,37,1,'2004','text/plain','admin','2019-08-05 17:52:07','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (39,38,1,'${/infrastructure/graphite/host}:${/infrastructure/graphite/port/carbide}','application/x-template','admin','2019-08-05 17:53:12','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (40,39,1,'1','text/plain','admin','2019-08-05 18:02:17','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (41,40,1,NULL,'application/x-null','admin','2019-08-05 18:05:39','Init infra',0);
INSERT INTO `my_config_tree_log` VALUES (42,41,1,'alpha*.gopher.example.com','text/plain','admin','2019-08-05 18:06:18','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (43,39,2,'[{\"mime\":\"text/plain\",\"value\":\"1\"},{\"mime\":\"text/plain\",\"value\":\"0\",\"group\":\"gopher-alpha\"}]','application/x-case','admin','2019-08-05 18:06:54','Disable SMS verification on alpha',0);
INSERT INTO `my_config_tree_log` VALUES (44,42,1,NULL,'application/x-null','admin','2019-08-05 18:08:26','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (45,43,1,'/infrastructure/postgresql/gopher-main','application/x-symlink','admin','2019-08-05 18:08:44','Init gopher',0);
INSERT INTO `my_config_tree_log` VALUES (46,44,1,'42','text/plain','admin','2019-08-05 18:10:44','Enable gopherizator for 42% of users',0);
INSERT INTO `my_config_tree_log` VALUES (47,45,1,'userkvs1.squirrel.example.com','text/plain','admin','2019-08-05 18:15:40','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (48,46,1,'userkvs2.squirrel.example.com','text/plain','admin','2019-08-05 18:16:04','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (49,47,1,NULL,'application/x-null','admin','2019-08-05 18:16:50','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (50,48,1,'/infrastructure/tarantool/squirrel-user','application/x-symlink','admin','2019-08-05 18:17:09','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (51,49,1,NULL,'application/x-null','admin','2019-08-05 18:18:14','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (52,50,1,'${/infrasructure/graphite/host}:${/infrastucture/graphite/port/carbide}','application/x-template','admin','2019-08-05 18:18:53','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (53,51,1,NULL,'application/x-null','admin','2019-08-05 18:19:37','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (54,52,1,NULL,'application/x-null','admin','2019-08-05 18:19:53','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (55,54,1,'1','text/plain','admin','2019-08-05 18:21:32','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (56,55,1,'0','text/plain','admin','2019-08-05 18:22:22','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (57,56,1,'1','text/plain','admin','2019-08-05 18:23:36','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (58,57,1,NULL,'application/x-null','admin','2019-08-05 18:24:37','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (59,58,1,NULL,'application/x-null','admin','2019-08-05 18:26:29','Init squrrel',0);
INSERT INTO `my_config_tree_log` VALUES (60,59,1,'/squirrel/user','application/x-symlink','admin','2019-08-05 18:26:50','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (61,60,1,'/squirrel/mashroom','application/x-symlink','admin','2019-08-05 18:27:18','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (62,61,1,'/squirrel/user','application/x-symlink','admin','2019-08-05 18:27:34','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (63,62,1,'/squirrel/statistics','application/x-symlink','admin','2019-08-05 18:27:55','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (64,63,1,'/squirrel/nut','application/x-symlink','admin','2019-08-05 18:28:11','Init squirrel',0);
INSERT INTO `my_config_tree_log` VALUES (65,50,2,'${/infrastructure/graphite/host}:${/infrastructure/graphite/port/carbide}','application/x-template','admin','2019-08-05 18:30:28','Fix misprint',0);
INSERT INTO `my_config_tree_log` VALUES (66,64,1,'{\n  \"BELKA5\": {\n    \"action\": \"discount\",\n    \"percent\": 5,\n    \"from\": \"2019-09-01\",\n    \"till\": \"2019-12-01\"\n  },\n  \"BELKA15\": {\n    \"premiumCard\": true,\n    \"action\": \"discount\",\n    \"percent\": 15,\n    \"from\": \"2019-09-01\",\n    \"till\": \"2019-12-01\"\n  },\n  \"STRELKA200\": {\n    \"action\": \"bonus\",\n    \"value\": 200,\n    \"from\": \"2019-01-01\",\n    \"till\": \"2020-01-01\"\n  }\n}','application/json','admin','2019-09-16 16:23:11','json example',0);
INSERT INTO `my_config_tree_log` VALUES (67,65,1,'/squirrel/promocode','application/x-symlink','admin','2019-09-16 16:26:42','json example',0);

--
-- Dumping data for table `my_config_user_group`
--

INSERT INTO `my_config_user_group` VALUES ('admin',1);
INSERT INTO `my_config_user_group` VALUES ('hedgehog',2);
INSERT INTO `my_config_user_group` VALUES ('meerkat',3);
INSERT INTO `my_config_user_group` VALUES ('rabbit',4);
INSERT INTO `my_config_user_group` VALUES ('beaver',5);
