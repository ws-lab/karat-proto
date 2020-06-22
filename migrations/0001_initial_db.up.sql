CREATE TABLE resource (
	id	 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	name	 TEXT NOT NULL
);
CREATE TABLE `rvariable` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	`name`	TEXT NOT NULL,
	`alias`	TEXT NOT NULL,
	`mask_type`	TEXT NOT NULL
);
CREATE TABLE `rvalue` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	`name`	TEXT NOT NULL,
	`resource_id`	INTEGER NOT NULL,
	`note`	TEXT,
	`rvariable_id`	INTEGER,
	`unit_id`	INTEGER,
FOREIGN KEY (rvariable_id)
    REFERENCES rvariable (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (unit_id)
    REFERENCES unit (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (resource_id)
    REFERENCES resource (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
);

PRAGMA foreign_keys=on;
CREATE TABLE `unit` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	`rvariable_id`	INTEGER NOT NULL,
	`name`	text NOT NULL,
	`multiplier`	numeric NOT NULL,
FOREIGN KEY (rvariable_id)
    REFERENCES rvariable (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
);

PRAGMA foreign_keys=on;
CREATE TABLE `dconf_ext` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	`code`	INTEGER NOT NULL,
	`index`	INTEGER NOT NULL,
	`resource_id`	INTEGER NOT NULL,
	`rvariable_id`	INTEGER NOT NULL,
	`rvalue_id`	INTEGER,
	`len`	INTEGER NOT NULL,
	`type_data`	INTEGER NOT NULL,
	`little_endian`	INTEGER NOT NULL,
	`is_sensor`	INTEGER NOT NULL,
	`unit_id`	INTEGER,
FOREIGN KEY (code)
    REFERENCES dconf (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (resource_id)
    REFERENCES resource (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (rvariable_id)
    REFERENCES rvariable (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (rvalue_id)
    REFERENCES rvalue (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
FOREIGN KEY (unit_id)
    REFERENCES unit (id) 
       ON UPDATE CASCADE
       ON DELETE RESTRICT
);

CREATE UNIQUE INDEX dconf_ext_code_index_uniq ON dconf_ext(code, `index`);

CREATE TABLE `dconf` (
	`id`	INTEGER NOT NULL UNIQUE,
	`vr`	TEXT NOT NULL,
	`dmodels`	TEXT,
	`packet_type`	INTEGER NOT NULL,
	PRIMARY KEY(id)
);

CREATE TABLE `flags` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	`vr`	TEXT,
	`note`	TEXT NOT NULL,
	`byte`	TEXT NOT NULL,
	`device`	TEXT,
	`type`	INTEGER NOT NULL,
	`nbit`	INTEGER,
	`ext_id`	INTEGER
);

CREATE TABLE `packet_error` (
	`id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	`vr`	TEXT NOT NULL,
	`note`	TEXT NOT NULL,
	`byte`	TEXT NOT NULL
);


