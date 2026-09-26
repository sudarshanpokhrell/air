package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// App roles. Global admins (users.is_admin) act as RoleAdmin in every app.
const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
)

var appRoleRank = map[string]int{
	RoleDeveloper: 1,
	RoleAdmin:     2,
}

// RoleAtLeast reports whether role grants at least min.
func RoleAtLeast(role, min string) bool {
	return appRoleRank[role] >= appRoleRank[min] && appRoleRank[role] > 0
}

type Member struct {
	AppID     string    `json:"-"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type MemberStore struct {
	db *sql.DB
}

// GetMemberRole returns the user's role in the app, or ErrNotFound if they're not a member.
func (s *MemberStore) GetMemberRole(ctx context.Context, appID, userID string) (string, error) {
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM app_members WHERE app_id = $1 AND user_id = $2`, appID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (s *MemberStore) ListMembers(ctx context.Context, appID string) ([]*Member, error) {
	const q = `
		SELECT m.app_id, m.user_id, u.email, u.name, m.role, m.created_at
		FROM app_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.app_id = $1
		ORDER BY m.created_at`

	rows, err := s.db.QueryContext(ctx, q, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []*Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.AppID, &m.UserID, &m.Email, &m.Name, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

// AddMember returns ErrConflict if the user is already a member.
func (s *MemberStore) AddMember(ctx context.Context, appID, userID, role string) error {
	const q = `INSERT INTO app_members (app_id, user_id, role) VALUES ($1, $2, $3::app_role)`

	_, err := s.db.ExecContext(ctx, q, appID, userID, role)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func (s *MemberStore) UpdateMemberRole(ctx context.Context, appID, userID, role string) error {
	const q = `UPDATE app_members SET role = $3::app_role WHERE app_id = $1 AND user_id = $2`

	res, err := s.db.ExecContext(ctx, q, appID, userID, role)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MemberStore) RemoveMember(ctx context.Context, appID, userID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM app_members WHERE app_id = $1 AND user_id = $2`, appID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
