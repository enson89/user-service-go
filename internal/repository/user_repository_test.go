// internal/repository/user_repository_test.go
package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"user-service/internal/model"
	"user-service/internal/repository"
)

func TestCreate_UserRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	u := &model.User{Email: "a@b.com", PasswordHash: "h", Role: "user"}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
	)).
		WithArgs(u.Email, u.PasswordHash, u.Role).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectCommit()

	err = repo.Create(context.Background(), u)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), u.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, email, password_hash, role FROM users WHERE email = $1`,
	)).
		WithArgs("no@one.com").
		WillReturnError(sql.ErrNoRows)

	u, err := repo.GetByEmail(context.Background(), "no@one.com")
	assert.NoError(t, err)
	assert.Nil(t, u)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_Found(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
		AddRow(7, "x@y.com", "hash", "admin")
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, email, password_hash, role FROM users WHERE email = $1`,
	)).
		WithArgs("x@y.com").
		WillReturnRows(rows)

	u, err := repo.GetByEmail(context.Background(), "x@y.com")
	assert.NoError(t, err)
	assert.Equal(t, int64(7), u.ID)
	assert.Equal(t, "admin", u.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
		AddRow(3, "u@v.com", "pwh", "user")
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, email, password_hash, role FROM users WHERE id = $1`,
	)).
		WithArgs(int64(3)).
		WillReturnRows(rows)

	u, err := repo.GetByID(context.Background(), 3)
	assert.NoError(t, err)
	assert.Equal(t, "u@v.com", u.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM users WHERE id = $1`,
	)).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 5)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM users WHERE id = $1`,
	)).
		WithArgs(int64(6)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 6)
	assert.EqualError(t, err, "no user found to delete")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	u := &model.User{ID: 5, Name: "Bob"}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2`,
	)).
		WithArgs("Bob", int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.Update(context.Background(), u)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	u := &model.User{ID: 5, Name: "Bob"}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2`,
	)).
		WithArgs("Bob", int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.Update(context.Background(), u)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
