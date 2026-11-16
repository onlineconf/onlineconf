-- Value encryption at rest, the audit log, and the schema version.
--
-- Encrypted marks a value stored as an AES-256-GCM envelope written by
-- onlineconf-admin with a key from its key file. Old binaries never
-- reference the new columns and tables (every read and insert uses explicit
-- column lists), so this migration may be applied while they are still
-- running. Applying it alone changes no behavior: the feature stays off until
-- /onlineconf/value-encryption is created with the value 1 (fresh
-- installations get that parameter from schema.sql) and a key file is
-- configured, so an upgrade performed without reading the release notes
-- cannot end with old instances serving envelopes as values.
--
-- The column additions are ALGORITHM=INSTANT on MySQL >= 8.0.29 (older 8.0
-- rebuilds the two tables in place); the indexes, which the rotation sweep
-- and the startup checks select by, are built in place, online.

ALTER TABLE `my_config_tree`
    ADD COLUMN `Encrypted` tinyint(1) NOT NULL default '0' AFTER `ContentType`;
ALTER TABLE `my_config_tree`
    ADD KEY `Encrypted` (`Encrypted`);

ALTER TABLE `my_config_tree_log`
    ADD COLUMN `Encrypted` tinyint(1) NOT NULL default '0' AFTER `ContentType`;
ALTER TABLE `my_config_tree_log`
    ADD KEY `Encrypted` (`Encrypted`);

-- Registry of the value encryption keys ever loaded: no key material, only a
-- verifier (the fixed registry constant sealed under the key) that lets an
-- instance prove its key file holds the material registered under an id, and
-- the Active mark the stored data is encrypted under once a rotation sweep
-- completed: the key in use.
CREATE TABLE `my_config_encryption_key` (
    `KeyID` varchar(32) character set ascii NOT NULL,
    `Verifier` varbinary(64) NOT NULL,
    `Active` tinyint(1) NOT NULL default '0',
    `CreatedAt` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`KeyID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- The audit log: the reveals of encrypted values — reveal_view: the current
-- value was shown, to view or to edit it; reveal_history: the parameter's
-- history was shown — written before any decryption happens and served to
-- bots by the botapi audit feed.
CREATE TABLE `my_config_audit` (
    `ID` bigint(20) unsigned NOT NULL auto_increment,
    `NodeID` bigint(20) unsigned NOT NULL,
    `Path` varchar(512) character set ascii collate ascii_bin NOT NULL default '',
    `User` varchar(128) character set ascii NOT NULL,
    `Kind` enum('reveal_view','reveal_history') character set ascii NOT NULL,
    `Time` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`ID`),
    KEY `NodeID` (`NodeID`),
    KEY `Time` (`Time`),
    CONSTRAINT `my_config_audit_ibfk_1` FOREIGN KEY (`NodeID`) REFERENCES `my_config_tree` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- The schema version. Every admin/etc/update-N.sql ends with
-- INSERT INTO my_config_schema (Version) VALUES (N), and schema.sql records
-- the version it creates, so that onlineconf-admin refuses to start against
-- a schema behind the one it is written for, naming the files to apply. A
-- database without this table is at version 7 or older. Applying an update
-- twice fails on the primary key, as it should.
CREATE TABLE `my_config_schema` (
    `Version` int(11) unsigned NOT NULL,
    `AppliedAt` timestamp NOT NULL default CURRENT_TIMESTAMP,
    PRIMARY KEY  (`Version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `my_config_schema` (`Version`) VALUES (8);
