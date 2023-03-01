package postgres

import (
	"context"
	"fmt"
	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/22Fariz22/loyal/pkg/postgres"
	"log"
)

type User struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRepository struct {
	*postgres.Postgres
}

func NewUserRepository(db *postgres.Postgres) *UserRepository {
	return &UserRepository{db}
}

func (u *UserRepository) CreateUser(ctx context.Context, l logger.Interface, user *entity.User) error {
	var login string

	// проверяем чтобы не было пустого логина
	if len(user.Login) == 0 {
		l.Error("login length equal 0")
		return auth.ErrBadRequest
	}

	// проверяем чтобы не было пустого пароля
	if len(user.Password) == 0 {
		l.Error("password length equal 0")
		return auth.ErrBadRequest
	}

	// проверяем есть ли такой логин
	_ = u.Pool.QueryRow(ctx, `SELECT login FROM users where login = $1;`, user.Login).Scan(&login)
	if len(login) != 0 {
		l.Error("Login is already taken.")
		return auth.ErrLoginIsAlreadyTaken
	}

	// вставляем новый логин и пароль
	log.Println("auth-db-CreateUser()-user.Login, user.Password: ", user.Login, user.Password)
	_, err := u.Pool.Exec(ctx, "INSERT INTO users(login, password) values($1, $2);", user.Login, user.Password)
	if err != nil {
		l.Error("error in pool.Exec - INSERT:", err)
		return err
	}

	return nil
}

func (u *UserRepository) GetUser(ctx context.Context, l logger.Interface, login, password string) (*entity.User, error) {
	row, err := u.Pool.Query(ctx, "select user_id,login,password from users where login = $1 and password = $2", login, password)
	if err != nil {
		l.Error("error in pool.Query SELECT.")
		return nil, err
	}
	defer row.Close()

	rows := make([]User, 1)

	for row.Next() {
		var u User
		fmt.Println(row.Values())
		err := row.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			l.Error("Error in row.Scan().")
			return nil, auth.ErrUserNotFound
		}

		rows = append(rows, u)
	}
	if len(rows) < 2 {
		return nil, auth.ErrUserNotFound
	}

	fmt.Println("db-GetUser()-rows", rows)
	fmt.Println("db-GetUser()-len(rows)", len(rows))
	fmt.Println("db-GetUser()-rows[0]", rows[0])
	fmt.Println("db-GetUser()-rows[1]", rows[1])
	return toEntity(&rows[1]), nil
}

func toEntity(u *User) *entity.User {
	return &entity.User{
		ID:       u.ID,
		Login:    u.Login,
		Password: u.Password,
	}
}
