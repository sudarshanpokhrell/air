package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sudarshanpokhrell/air/internal/validator"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  password  `json:"-"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.text = &plaintext
	p.hash = hash
	return nil
}

func (p *password) Compare(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//Validations

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.EmailRX.MatchString(email), "email", "must be a valid email address")
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.Check(len(name) <= 100, "name", "must not be more than 100 bytes long")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	ValidatePasswordField(v, "password", password)
}

func ValidatePasswordField(v *validator.Validator, key, password string) {
	v.Check(password != "", key, "must be provided")
	v.Check(len(password) >= 8, key, "must be at least 8 bytes long")
	v.Check(len(password) <= 72, key, "must not be more than 72 bytes long") // bcrypt limit
}

func ValidateUser(v *validator.Validator, u *User, plaintextPassword string) {
	ValidateEmail(v, u.Email)
	ValidateName(v, u.Name)
	ValidatePasswordPlaintext(v, plaintextPassword)
}

// Store

type UserStore struct {
	db *sql.DB
}

// CreateUser inserts u (after u.Password.Set) and fills in ID, IsActive,
// CreatedAt and UpdatedAt. Returns ErrConflict if the email is taken.
func (s *UserStore) CreateUser(ctx context.Context, u *User) error {
	const q = `
		INSERT INTO users (email, name, password_hash, is_admin)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, q,
		u.Email,
		u.Name,
		u.Password.hash,
		u.IsAdmin,
	).Scan(
		&u.ID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func (s *UserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	const q = `
		SELECT id, email, name, password_hash, is_admin, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u User
	err := s.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&u.Password.hash,
		&u.IsAdmin,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, email, name, password_hash, is_admin, is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	var u User
	err := s.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&u.Password.hash,
		&u.IsAdmin,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) ListUsers(ctx context.Context) ([]*User, error) {
	const q = `
		SELECT id, email, name, password_hash, is_admin, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Name,
			&u.Password.hash,
			&u.IsAdmin,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// UpdateUser saves name, is_admin and is_active.
func (s *UserStore) UpdateUser(ctx context.Context, u *User) error {
	const q = `
		UPDATE users
		SET name = $2, is_admin = $3, is_active = $4, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, q,
		u.ID,
		u.Name,
		u.IsAdmin,
		u.IsActive,
	).Scan(&u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// UpdatePassword saves u.Password (after u.Password.Set).
func (s *UserStore) UpdatePassword(ctx context.Context, u *User) error {
	const q = `
		UPDATE users
		SET password_hash = $2, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, q, u.ID, u.Password.hash).Scan(&u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
