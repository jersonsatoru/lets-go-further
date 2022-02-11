package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/jersonsatoru/lets-go-further/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateUser(v *validator.Validator, u *User) {
	v.Check(u.Name != "", "name", "must be greater than 0")
	v.Check(len(u.Name) <= 500, "name", "must have the maximum of 500 characters")
	ValidateEmail(v, u.Email)
	ValidatePasswordPlaintext(v, *u.Password.plaintext)
	if u.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be greater than 0")
	v.Check(validator.Matches(email, validator.EmailRxp), "email", "must be a valid email")
}

func ValidatePasswordPlaintext(v *validator.Validator, plaintextPassword string) {
	v.Check(plaintextPassword != "", "password", "must not be empty")
	v.Check(len(plaintextPassword) >= 8, "password", "must greater than 8")
	v.Check(len(plaintextPassword) <= 72, "password", "must have maximum of 72 characters")
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, version, id
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.CreatedAt, &user.Version, &user.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, name, email, activated, version, created_at, password_hash
		FROM users
		WHERE email = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var user User
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Version,
		&user.CreatedAt,
		&user.Password.hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m *UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, activated = $3, password_hash = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	args := []interface{}{
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Password.hash,
		&user.ID,
		&user.Version,
	}
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(plaintextToken, tokenScope string) (*User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.activated, u.created_at, u.version, u.password_hash
		FROM users u INNER JOIN tokens t ON (u.id = t.user_id)
		WHERE t.hash = $1 AND NOW() < t.expiry AND t.scope = $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	hash := sha256.Sum256([]byte(plaintextToken))
	var user User
	err := m.DB.QueryRowContext(ctx, query, hash[:], tokenScope).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.CreatedAt,
		&user.Version,
		&user.Password.hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
