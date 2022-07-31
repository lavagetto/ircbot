CREATE TABLE contacts (`name` VARCHAR(256) PRIMARY KEY, `phone` VARCHAR(256), `email` VARCHAR(256));
CREATE TABLE topics (`channel` VARCHAR(256) PRIMARY KEY, `topic` TEXT);
CREATE TABLE acls (`command` VARCHAR(256), `identifier` VARCHAR(256), PRIMARY KEY (`command`, `identifier`));
