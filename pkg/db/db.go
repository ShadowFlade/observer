package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

func (d *Db) GetRegularUsers() ([]string, []int) {
	usersRes, err := d.tx.Queryx("select * from users")
	if err != nil {
		log.Fatal(err)
	}
	defer usersRes.Close()

	var users []string
	var ids []int
	for usersRes.Next() {
		var user IUser
		err := usersRes.Scan(&user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user.USER)
		ids = append(ids, user.ID)
	}

	return users, ids

}

type UserStatDB struct {
	mem_usage         float32   `db:"mem_usage"`
	mem_usage_percent float32   `db:"mem_usage_percent"`
	date_inserted     time.Time `db:"date_inserted"`
	user_id           int       `db:"user_id"`
	day_active_users  int       `db:"day_active_users"`
}

func (d *Db) WriteStats(totalMemUsage float32, totalMemUsagePercent float32, userId int, activeUsers int) bool {
	useStatDB := UserStatDB{
		mem_usage:         float32(totalMemUsage),
		mem_usage_percent: float32(totalMemUsagePercent),
		user_id:           userId,
		day_active_users:  activeUsers,
		date_inserted:     time.Now(),
	}
	res, err := d.tx.NamedExec(`insert into stats (mem_usage,mem_usage_percent,user_id,day_active_users,date_inserted) values (:mem_usage, :mem_usage_percent, :user_id, :day_active_users, :date_inserted`, useStatDB)

	if err != nil {
		log.Fatalf("Could not insert into stats with user_id: %d", userId)
	}
	id, err := res.LastInsertId()

	if err != nil {
		panic("I dont know what i do anymore have to quit")
	}

	if id > 0 {
		return true
	}
	return false
}

func (d *Db) IsDbPresent() bool {
	rows, err := d.db.Queryx("SELECT table_name FROM information_schema.tables WHERE table_schema = 'observer'")

	if err != nil {
		log.Fatal("Could not check for existing of a database")
	}
	defer rows.Close()
	tables := []string{}

	// Iterate over the result set and retrieve table names
	tableCount := 0
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}
		tables = append(tables, tableName)
		tableCount++
	}

	if tableCount == 0 {
		return false
	} else {
		return true
	}

}

// TODO refactor: return []string of successfully created table/db messages
func (d *Db) CreateSchema() error {
	sqlDbCreate := "create database observser;"
	res, err := d.db.Exec(sqlDbCreate)
	log.Println(res, ": create database observer result")

	if err != nil {
		return errors.New("Could not create database observer")
	}
	sqlQueryCreate := "create table observer.stats (id int auto_increment not null, mem_usage float not null, mem_usage_percent float not null ,date_inserted datetime not null, primary key (`id`), user_id int not null, day_active_users int not null);"
	res, err = d.db.Exec(sqlQueryCreate)
	if err != nil {
		return errors.New("Could not create table stats")
	}
	fmt.Println(res, ": created table observer")

	sqlQueryUsers := "create table observer.users (id int auto_increment not null, user varchar(255), type varchar(20), primary(`id`)"
	res, err = d.db.Exec(sqlQueryUsers)
	if err != nil {
		return errors.New("Could now create table users")
	}
	return nil
}

