package migrations

const SchemaV11 = `
ALTER TABLE batches RENAME TO batches_legacy;
CREATE TABLE batches (
 id TEXT PRIMARY KEY, endpoint TEXT NOT NULL DEFAULT '', completion_window TEXT NOT NULL DEFAULT '24h',
 status TEXT NOT NULL, input_file_id TEXT NOT NULL DEFAULT '', output_file_id TEXT, error_file_id TEXT,
 created_at INTEGER NOT NULL, in_progress_at INTEGER, expires_at INTEGER, finalizing_at INTEGER,
 completed_at INTEGER, failed_at INTEGER, expired_at INTEGER, cancelling_at INTEGER, cancelled_at INTEGER,
 request_counts_total INTEGER DEFAULT 0, request_counts_completed INTEGER DEFAULT 0, request_counts_failed INTEGER DEFAULT 0,
 metadata TEXT, api_key_id TEXT, errors TEXT, model TEXT, usage TEXT,
 output_expires_after_seconds INTEGER, output_expires_after_anchor TEXT
);
INSERT INTO batches(id,endpoint,status,created_at,expires_at,request_counts_total,request_counts_completed,request_counts_failed)
 SELECT id,endpoint,status,COALESCE(CAST(strftime('%s',created_at) AS INTEGER),0),CAST(strftime('%s',expires_at) AS INTEGER),request_count,completed_count,failed_count FROM batches_legacy;
CREATE INDEX idx_batches_api_key ON batches(api_key_id);
CREATE INDEX idx_batches_status ON batches(status);
`
