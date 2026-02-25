package db

import "database/sql"

// InitDataSchema creates the data DB tables if they do not exist.
// Data DB is the source of truth — append-only, never rebuilt.
func InitDataSchema(d *sql.DB) error {
	_, err := d.Exec(dataDDL)
	return err
}

// InitIndexSchema creates the index DB tables if they do not exist.
// Index DB is derived — can be dropped and rebuilt from data DB.
func InitIndexSchema(d *sql.DB) error {
	_, err := d.Exec(indexDDL)
	return err
}

const dataDDL = `
CREATE TABLE IF NOT EXISTS sessions (
	id              VARCHAR PRIMARY KEY,
	session_hash    VARCHAR NOT NULL,
	payload         VARCHAR,
	captured_at     TIMESTAMP NOT NULL,
	actor_type      VARCHAR NOT NULL DEFAULT 'human',
	agent_id        VARCHAR,
	user_email      VARCHAR
);

CREATE TABLE IF NOT EXISTS checkpoints (
	id              VARCHAR PRIMARY KEY,
	git_sha         VARCHAR NOT NULL,
	git_branch      VARCHAR NOT NULL,
	user_email      VARCHAR NOT NULL,
	ts              TIMESTAMP NOT NULL,
	actor_type      VARCHAR NOT NULL DEFAULT 'human',
	agent_id        VARCHAR
);

CREATE TABLE IF NOT EXISTS files_touched (
	id              VARCHAR PRIMARY KEY,
	checkpoint_id   VARCHAR NOT NULL REFERENCES checkpoints(id),
	file_path       VARCHAR NOT NULL,
	change_type     VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS checkpoint_sessions (
	checkpoint_id   VARCHAR NOT NULL REFERENCES checkpoints(id),
	session_id      VARCHAR NOT NULL REFERENCES sessions(id),
	PRIMARY KEY (checkpoint_id, session_id)
);
`

const indexDDL = `
CREATE TABLE IF NOT EXISTS turns_ft (
	session_id      VARCHAR NOT NULL,
	turn_index      INTEGER NOT NULL,
	role            VARCHAR NOT NULL,
	content         VARCHAR NOT NULL,
	PRIMARY KEY (session_id, turn_index)
);

CREATE TABLE IF NOT EXISTS tool_calls_index (
	session_id      VARCHAR NOT NULL,
	call_order      INTEGER NOT NULL,
	tool            VARCHAR NOT NULL,
	path            VARCHAR,
	cmd_prefix      VARCHAR,
	PRIMARY KEY (session_id, call_order)
);

CREATE TABLE IF NOT EXISTS files_index (
	checkpoint_id   VARCHAR NOT NULL,
	file_path       VARCHAR NOT NULL,
	PRIMARY KEY (checkpoint_id, file_path)
);

CREATE TABLE IF NOT EXISTS session_facets (
	session_id      VARCHAR PRIMARY KEY,
	user_email      VARCHAR,
	git_branch      VARCHAR,
	turn_count      INTEGER,
	actor_type      VARCHAR,
	agent_id        VARCHAR
);

CREATE TABLE IF NOT EXISTS file_cooccurrence (
	file_a          VARCHAR NOT NULL,
	file_b          VARCHAR NOT NULL,
	count           INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY (file_a, file_b)
);
`
