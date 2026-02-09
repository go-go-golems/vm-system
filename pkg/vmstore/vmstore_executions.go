package vmstore

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// CreateExecution creates a new execution.
func (s *VMStore) CreateExecution(exec *vmmodels.Execution) error {
	_, err := s.db.Exec(`
		INSERT INTO execution (id, session_id, kind, input, path, args_json, env_json, status, started_at, metrics_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, exec.SessionID, exec.Kind, exec.Input, exec.Path, exec.Args, exec.Env, exec.Status, exec.StartedAt.Unix(), exec.Metrics)
	return err
}

// UpdateExecution updates an execution.
func (s *VMStore) UpdateExecution(exec *vmmodels.Execution) error {
	var endedAt interface{}
	if exec.EndedAt != nil {
		endedAt = exec.EndedAt.Unix()
	}

	_, err := s.db.Exec(`
		UPDATE execution SET status = ?, ended_at = ?, result_json = ?, error_json = ?, metrics_json = ?
		WHERE id = ?
	`, exec.Status, endedAt, exec.Result, exec.Error, exec.Metrics, exec.ID)
	return err
}

// GetExecution retrieves an execution by ID.
func (s *VMStore) GetExecution(id string) (*vmmodels.Execution, error) {
	var exec vmmodels.Execution
	var startedAt int64
	var endedAt sql.NullInt64
	var input, path sql.NullString
	var result, errorJSON sql.NullString

	err := s.db.QueryRow(`
		SELECT id, session_id, kind, input, path, args_json, env_json, status, started_at, ended_at, result_json, error_json, metrics_json
		FROM execution WHERE id = ?
	`, id).Scan(&exec.ID, &exec.SessionID, &exec.Kind, &input, &path, &exec.Args, &exec.Env, &exec.Status, &startedAt, &endedAt, &result, &errorJSON, &exec.Metrics)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrExecutionNotFound
	}
	if err != nil {
		return nil, err
	}

	exec.StartedAt = time.Unix(startedAt, 0)
	if endedAt.Valid {
		t := time.Unix(endedAt.Int64, 0)
		exec.EndedAt = &t
	}
	if input.Valid {
		exec.Input = input.String
	}
	if path.Valid {
		exec.Path = path.String
	}
	if result.Valid {
		exec.Result = json.RawMessage(result.String)
	}
	if errorJSON.Valid {
		exec.Error = json.RawMessage(errorJSON.String)
	}

	return &exec, nil
}

// ListExecutions lists executions for a session.
func (s *VMStore) ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, kind, input, path, args_json, env_json, status, started_at, ended_at, result_json, error_json, metrics_json
		FROM execution WHERE session_id = ? ORDER BY started_at DESC LIMIT ?
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []*vmmodels.Execution
	for rows.Next() {
		var exec vmmodels.Execution
		var startedAt int64
		var endedAt sql.NullInt64
		var input, path sql.NullString
		var result, errorJSON sql.NullString

		if err := rows.Scan(&exec.ID, &exec.SessionID, &exec.Kind, &input, &path, &exec.Args, &exec.Env, &exec.Status, &startedAt, &endedAt, &result, &errorJSON, &exec.Metrics); err != nil {
			return nil, err
		}

		exec.StartedAt = time.Unix(startedAt, 0)
		if endedAt.Valid {
			t := time.Unix(endedAt.Int64, 0)
			exec.EndedAt = &t
		}
		if input.Valid {
			exec.Input = input.String
		}
		if path.Valid {
			exec.Path = path.String
		}
		if result.Valid {
			exec.Result = json.RawMessage(result.String)
		}
		if errorJSON.Valid {
			exec.Error = json.RawMessage(errorJSON.String)
		}

		execs = append(execs, &exec)
	}

	return execs, rows.Err()
}

// AddEvent adds an event to an execution.
func (s *VMStore) AddEvent(event *vmmodels.ExecutionEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO execution_event (execution_id, seq, ts, type, payload_json)
		VALUES (?, ?, ?, ?, ?)
	`, event.ExecutionID, event.Seq, event.Ts.Unix(), event.Type, event.Payload)
	return err
}

// GetEvents retrieves events for an execution.
func (s *VMStore) GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	rows, err := s.db.Query(`
		SELECT execution_id, seq, ts, type, payload_json
		FROM execution_event WHERE execution_id = ? AND seq > ? ORDER BY seq
	`, executionID, afterSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*vmmodels.ExecutionEvent
	for rows.Next() {
		var event vmmodels.ExecutionEvent
		var ts int64

		if err := rows.Scan(&event.ExecutionID, &event.Seq, &ts, &event.Type, &event.Payload); err != nil {
			return nil, err
		}

		event.Ts = time.Unix(ts, 0)
		events = append(events, &event)
	}

	return events, rows.Err()
}
