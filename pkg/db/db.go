package db

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofor-little/env"
	"github.com/jmoiron/sqlx"
)

type IUser struct {
	ID   int    `db:"id"`
	USER string `db:"user"`
	TYPE string `db:"type"`
}

type Db struct {
	db       *sqlx.DB
	tx       *sqlx.Tx
	dbName   string
	dbHost   string
	login    string
	password string
	cols     []string
}

func (d *Db) Connect() *sqlx.DB {
	_, err := os.Executable()

	if err != nil {
		panic(err)
	}

	if err := env.Load("./.env"); err != nil {
		fmt.Println("error")
		panic(err)
	}
	d.login = env.Get("DB_LOGIN", "i")
	d.password = env.Get("DB_PASS", "fucked")
	d.dbName = env.Get("DB_NAME", "urmom")
	d.dbHost = env.Get("DB_HOST", "host")
	connectStr := fmt.Sprintf("%s:%s@(127.0.0.1:3306)/%s", d.login, d.password, d.dbName)
	db, err := sqlx.Connect("mysql", connectStr)
	d.db = db

	return db

}

func (d *Db) WriteRegularUser(user string) (int64, error) {
	tx := d.db.MustBegin()

	res, err := tx.NamedExec(`INSERT INTO users (user, type) VALUES (:user, regular)`, user)

	if err != nil {
		return 0.00, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0.00, err
	}

	errN := tx.Commit()

	if errN != nil {
		return 0.00, errN
	}

	return id, nil
}

func (d *Db) GetRegularUsers() []string {
	usersRes, err := d.tx.Queryx("select user from users")
	if err != nil {
		log.Fatal(err)
	}
	defer usersRes.Close()

	var users []string
	for usersRes.Next() {
		var user string
		err := usersRes.Scan(&user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	return users

}
