package vmmodels

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidTemplateID  = errors.New("invalid template id")
	ErrInvalidSessionID   = errors.New("invalid session id")
	ErrInvalidExecutionID = errors.New("invalid execution id")
)

type TemplateID string

type SessionID string

type ExecutionID string

func ParseTemplateID(raw string) (TemplateID, error) {
	normalized, err := parseUUIDString(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidTemplateID, err)
	}
	return TemplateID(normalized), nil
}

func ParseSessionID(raw string) (SessionID, error) {
	normalized, err := parseUUIDString(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidSessionID, err)
	}
	return SessionID(normalized), nil
}

func ParseExecutionID(raw string) (ExecutionID, error) {
	normalized, err := parseUUIDString(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidExecutionID, err)
	}
	return ExecutionID(normalized), nil
}

func (id TemplateID) String() string {
	return string(id)
}

func (id SessionID) String() string {
	return string(id)
}

func (id ExecutionID) String() string {
	return string(id)
}

func parseUUIDString(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("empty")
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}
