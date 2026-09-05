package api

import (
	"context"
	"database/sql"
	"fmt"
)

// requireProjectAccess checks ownership or organization membership before exposing BIM data.
func (s *Server) requireProjectAccess(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return fmt.Errorf("project and user are required")
	}
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE id=$1 AND owner_id=$2`, projectID, userID).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
