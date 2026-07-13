package db

import "database/sql"

type FileRecord struct {
	ID        string `json:"id"`
	Bytes     int64  `json:"bytes"`
	CreatedAt int64  `json:"createdAt"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
	MimeType  string `json:"mimeType,omitempty"`
	APIKeyID  string `json:"apiKeyId,omitempty"`
	ExpiresAt *int64 `json:"expiresAt"`
}

func ListFiles(db *sql.DB, limit int) ([]FileRecord, error) {
	rows, err := db.Query(`SELECT id,bytes,created_at,filename,purpose,COALESCE(mime_type,''),COALESCE(api_key_id,''),expires_at
		FROM files WHERE deleted_at IS NULL ORDER BY created_at DESC,id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := make([]FileRecord, 0)
	for rows.Next() {
		var file FileRecord
		var expires sql.NullInt64
		if err := rows.Scan(&file.ID, &file.Bytes, &file.CreatedAt, &file.Filename, &file.Purpose, &file.MimeType, &file.APIKeyID, &expires); err != nil {
			return nil, err
		}
		if expires.Valid {
			file.ExpiresAt = &expires.Int64
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func GetFileContent(db *sql.DB, id string) (FileRecord, []byte, error) {
	var file FileRecord
	var content []byte
	var expires sql.NullInt64
	err := db.QueryRow(`SELECT id,bytes,created_at,filename,purpose,COALESCE(mime_type,''),COALESCE(api_key_id,''),expires_at,content
		FROM files WHERE id=? AND deleted_at IS NULL`, id).Scan(&file.ID, &file.Bytes, &file.CreatedAt, &file.Filename, &file.Purpose, &file.MimeType, &file.APIKeyID, &expires, &content)
	if expires.Valid {
		file.ExpiresAt = &expires.Int64
	}
	return file, content, err
}
