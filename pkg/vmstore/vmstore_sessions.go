package vmstore

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// CreateSession creates a new VM session.
func (s *VMStore) CreateSession(session *vmmodels.VMSession) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_session (id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.VMID, session.WorkspaceID, session.BaseCommitOID, session.WorktreePath, session.Status, session.CreatedAt.Unix())
	return err
}

// GetSession retrieves a session by ID.
func (s *VMStore) GetSession(id string) (*vmmodels.VMSession, error) {
	var session vmmodels.VMSession
	var createdAt int64
	var closedAt sql.NullInt64
	var lastError sql.NullString
	var runtimeMeta sql.NullString

	err := s.db.QueryRow(`
		SELECT id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at, closed_at, last_error_json, runtime_meta_json
		FROM vm_session WHERE id = ?
	`, id).Scan(&session.ID, &session.VMID, &session.WorkspaceID, &session.BaseCommitOID, &session.WorktreePath, &session.Status, &createdAt, &closedAt, &lastError, &runtimeMeta)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	session.CreatedAt = time.Unix(createdAt, 0)
	if closedAt.Valid {
		t := time.Unix(closedAt.Int64, 0)
		session.ClosedAt = &t
	}
	if lastError.Valid {
		session.LastError = lastError.String
	}
	if runtimeMeta.Valid {
		session.RuntimeMeta = json.RawMessage(runtimeMeta.String)
	}

	return &session, nil
}

// UpdateSession updates a session.
func (s *VMStore) UpdateSession(session *vmmodels.VMSession) error {
	var closedAt interface{}
	if session.ClosedAt != nil {
		closedAt = session.ClosedAt.Unix()
	}

	_, err := s.db.Exec(`
		UPDATE vm_session SET status = ?, closed_at = ?, last_error_json = ?, runtime_meta_json = ?
		WHERE id = ?
	`, session.Status, closedAt, session.LastError, session.RuntimeMeta, session.ID)
	return err
}

// ListSessions lists sessions with optional status filter.
func (s *VMStore) ListSessions(status string) ([]*vmmodels.VMSession, error) {
	query := "SELECT id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at, closed_at, last_error_json, runtime_meta_json FROM vm_session"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*vmmodels.VMSession
	for rows.Next() {
		var session vmmodels.VMSession
		var createdAt int64
		var closedAt sql.NullInt64
		var lastError sql.NullString
		var runtimeMeta sql.NullString

		if err := rows.Scan(&session.ID, &session.VMID, &session.WorkspaceID, &session.BaseCommitOID, &session.WorktreePath, &session.Status, &createdAt, &closedAt, &lastError, &runtimeMeta); err != nil {
			return nil, err
		}

		session.CreatedAt = time.Unix(createdAt, 0)
		if closedAt.Valid {
			t := time.Unix(closedAt.Int64, 0)
			session.ClosedAt = &t
		}
		if lastError.Valid {
			session.LastError = lastError.String
		}
		if runtimeMeta.Valid {
			session.RuntimeMeta = json.RawMessage(runtimeMeta.String)
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}
