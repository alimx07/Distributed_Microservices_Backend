package main

import (
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(DB *sql.DB) *UserRepo {
	return &UserRepo{
		db: DB,
	}
}

func (repo *UserRepo) CreateUser(user User) error {
	_, err := repo.db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		user.UserName, user.Email, user.Password,
	)
	return err

}

func (repo *UserRepo) GetUserData(email string) (User, error) {
	var user User
	err := repo.db.QueryRow("SELECT userid , password From users Where email= $1", email).Scan(
		&user.UserID,
		&user.Password,
	)
	if err != nil {
		return User{}, err
	}
	return user, err
}
func (repo *UserRepo) GetUserByID(id int) (User, error) {
	var user User
	err := repo.db.QueryRow("SELECT username , email From users Where userid= $1", id).Scan(
		&user.UserName,
		&user.Email,
	)
	if err != nil {
		return User{}, err
	}
	return user, err
}

func (repo *UserRepo) UpdateUser(user User) error {
	_, err := repo.db.Exec("UPDATE users SET username = $1, email = $2, password =$3 WHERE id = $4")
	return err
}

func (repo *UserRepo) DeleteUser(userid int) error {
	_, err := repo.db.Exec("DELETE FROM users WHERE userid = $1", userid)
	return err
}
