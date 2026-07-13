package db

import (
	"database/sql"
	"encoding/json"
)

type BatchRecord struct {
	ID                     string      `json:"id"`
	Endpoint               string      `json:"endpoint"`
	CompletionWindow       string      `json:"completionWindow"`
	Status                 string      `json:"status"`
	InputFileID            string      `json:"inputFileId"`
	OutputFileID           *string     `json:"outputFileId"`
	ErrorFileID            *string     `json:"errorFileId"`
	CreatedAt              int64       `json:"createdAt"`
	ExpiresAt              *int64      `json:"expiresAt"`
	RequestCountsTotal     int         `json:"requestCountsTotal"`
	RequestCountsCompleted int         `json:"requestCountsCompleted"`
	RequestCountsFailed    int         `json:"requestCountsFailed"`
	Metadata               interface{} `json:"metadata"`
	Errors                 interface{} `json:"errors"`
	Model                  *string     `json:"model"`
	Usage                  interface{} `json:"usage"`
}

const batchColumns = `id,endpoint,completion_window,status,input_file_id,output_file_id,error_file_id,created_at,expires_at,
 request_counts_total,request_counts_completed,request_counts_failed,metadata,errors,model,usage`

func ListBatches(db *sql.DB, limit int) ([]BatchRecord, error) {
	rows, err := db.Query("SELECT "+batchColumns+" FROM batches ORDER BY created_at DESC,id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	batches := make([]BatchRecord, 0)
	for rows.Next() {
		batch, err := scanBatch(rows.Scan)
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	return batches, rows.Err()
}

func GetBatch(db *sql.DB, id string) (*BatchRecord, error) {
	batch, err := scanBatch(db.QueryRow("SELECT "+batchColumns+" FROM batches WHERE id=?", id).Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &batch, err
}

func scanBatch(scan func(...interface{}) error) (BatchRecord, error) {
	var batch BatchRecord
	var output, errorFile, model sql.NullString
	var expires sql.NullInt64
	var metadata, errors, usage sql.NullString
	err := scan(&batch.ID, &batch.Endpoint, &batch.CompletionWindow, &batch.Status, &batch.InputFileID,
		&output, &errorFile, &batch.CreatedAt, &expires, &batch.RequestCountsTotal,
		&batch.RequestCountsCompleted, &batch.RequestCountsFailed, &metadata, &errors, &model, &usage)
	if err != nil {
		return BatchRecord{}, err
	}
	if output.Valid {
		batch.OutputFileID = &output.String
	}
	if errorFile.Valid {
		batch.ErrorFileID = &errorFile.String
	}
	if model.Valid {
		batch.Model = &model.String
	}
	if expires.Valid {
		batch.ExpiresAt = &expires.Int64
	}
	batch.Metadata = parseNullableJSON(metadata)
	batch.Errors = parseNullableJSON(errors)
	batch.Usage = parseNullableJSON(usage)
	return batch, nil
}

func parseNullableJSON(value sql.NullString) interface{} {
	if !value.Valid {
		return nil
	}
	var parsed interface{}
	if json.Unmarshal([]byte(value.String), &parsed) != nil {
		return nil
	}
	return parsed
}
