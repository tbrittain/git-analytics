package store

// SchemaSQL contains the DDL for the analytics database.
const SchemaSQL = `
CREATE TABLE IF NOT EXISTS commits (
	hash         VARCHAR PRIMARY KEY,
	author_name  VARCHAR NOT NULL,
	author_email VARCHAR NOT NULL,
	committed_at TIMESTAMP NOT NULL,
	message      VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS file_stats (
	commit_hash VARCHAR NOT NULL,
	file_path   VARCHAR NOT NULL,
	additions   INTEGER NOT NULL,
	deletions   INTEGER NOT NULL,
	PRIMARY KEY (commit_hash, file_path)
);

CREATE TABLE IF NOT EXISTS index_state (
	key   VARCHAR PRIMARY KEY,
	value VARCHAR NOT NULL
);
`
