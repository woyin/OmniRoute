package migrations

const SchemaV10 = `
ALTER TABLE files RENAME TO files_legacy;
CREATE TABLE files (
 id TEXT PRIMARY KEY,
 bytes INTEGER NOT NULL,
 created_at INTEGER NOT NULL,
 filename TEXT NOT NULL,
 purpose TEXT NOT NULL,
 content BLOB,
 mime_type TEXT,
 api_key_id TEXT,
 deleted_at INTEGER,
 status TEXT DEFAULT 'validating',
 expires_at INTEGER
);
INSERT INTO files(id,bytes,created_at,filename,purpose,status)
 SELECT id,size,COALESCE(CAST(strftime('%s',created_at) AS INTEGER),0),filename,purpose,status FROM files_legacy;
CREATE INDEX idx_files_api_key ON files(api_key_id);
`
