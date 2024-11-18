package db

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofor-little/env"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"time"
)

type IUser struct {
	ID         int    `db:"id"`
	USER       string `db:"user"`
	TYPE       string `db:"type"`
	LN_USER_ID int    `db:"ln_user_id"`
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

type DbIsNotPresent error

func (d *Db) Init() {
	_, err := os.Executable()

	if err != nil {
		panic(err)
	}

	if err := env.Load("./.env"); err != nil {
		fmt.Println("error")
		panic(err)
	}

	db, err := d.Connect(false)

	if err != nil {
		db, err = d.ConnectAndCreateSchema(true)

		if err != nil {
			log.Fatal(err)
		}

	} else {
		d.db = db
	}

}

func (d *Db) Connect(isRetryWithoutDB bool) (*sqlx.DB, error) {
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
	connectStr := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", d.login, d.password, d.dbName)
	db, err := sqlx.Connect("mysql", connectStr)

	if err != nil { //here i guess we can only make an assumption that db is not created yet, mb figure out later what type of error this is
		return db, err
	}

	d.db = db

	return db, nil

}
func (d *Db) ConnectAndCreateSchema(isRetryWithoutDB bool) (*sqlx.DB, error) {
	d.login = env.Get("DB_LOGIN", "i")
	d.password = env.Get("DB_PASS", "fucked")
	d.dbName = env.Get("DB_NAME", "urmom")
	d.dbHost = env.Get("DB_HOST", "host")
	connectStr := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/", d.login, d.password)
	db, err := sqlx.Connect("mysql", connectStr)

	if err != nil { //here i guess we can only make an assumption that db is not created yet, mb figure out later what type of error this is
		log.Fatal(err)
	}

	d.db = db

	err = d.CreateSchema()
	if err != nil {
		log.Fatal(err)
	}

	db, err = d.Connect(false)

	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func (d *Db) WriteRegularUser(user string, id int32) (int64, error) {
	tx := d.db.MustBegin()
	res := tx.MustExec("INSERT INTO users (user, type, ln_user_id) VALUES (?, 'regular', ?)", user, id)

	insertedId, err := res.LastInsertId()

	if err != nil {
		return 0.00, err
	}

	errN := tx.Commit()

	if errN != nil {
		return 0.00, errN
	}

	return insertedId, nil
}

type T interface{}

func (d *Db) GetRegularUsers() ([]string, []int) {
	usersRes, err := d.db.Queryx("select * from users")
	if err != nil {
		log.Fatal(err)
	}
	defer usersRes.Close()

	var users []string
	var ids []int

	for usersRes.Next() {
		var user IUser
		err := usersRes.StructScan(&user)

		if err != nil {
			log.Fatal(err)
		}

		users = append(users, user.USER)
		ids = append(ids, user.ID)
	}

	return users, ids
}

type UserStatDB struct {
	MemUsage        float32   `db:"mem_usage"`
	MemUsagePercent float32   `db:"mem_usage_percent"`
	DateInserted    time.Time `db:"date_inserted"`
	UserId          int       `db:"user_id"`
	DayActiveUsers  int       `db:"day_active_users"`
}

func (d *Db) WriteStats(
	totalMemUsage float32,
	totalMemUsagePercent float32,
	userId int,
	activeUsers int,
) bool {
	// mode := env.Get("MODE", "dev")
	useStatDB := UserStatDB{
		MemUsage:        float32(totalMemUsage),
		MemUsagePercent: float32(totalMemUsagePercent),
		UserId:          userId,
		DayActiveUsers:  activeUsers,
		DateInserted:    time.Now(),
	}
	fmt.Printf("%+v\n", useStatDB)

	tx := d.db.MustBegin()

	// if mode == "dev" {
	// 	fmt.Println("inserting into stats", totalMemUsage, totalMemUsagePercent, userId, activeUsers)
	// 	return true
	// }

	res, err := tx.NamedExec(`insert into stats (mem_usage,mem_usage_percent,user_id,day_active_users,date_inserted) values (:mem_usage, :mem_usage_percent, :user_id, :day_active_users, :date_inserted)`, useStatDB)

	if err != nil {
		log.Fatalf(err.Error())
	}
	id, err := res.LastInsertId()

	if err != nil {
		panic("I dont know what i do anymore have to quit")
	}

	if id > 0 {
		errN := tx.Commit()
		if errN != nil {
			log.Fatal(errN)
		}
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
	sqlDbCreate := "create database observer;"
	res, err := d.db.Exec(sqlDbCreate)

	if err != nil {
		return errors.New("Could not create database observer")
	}

	sqlQueryCreate := "create table observer.stats (id int auto_increment not null, mem_usage float not null, mem_usage_percent float not null ,date_inserted datetime not null, primary key (`id`), user_id int not null, day_active_users int not null);"
	res, err = d.db.Exec(sqlQueryCreate)

	if err != nil {
		return errors.New("Could not create table stats")
	}

	log.Println(res, ": created table stats")

	//ln shows that this values is from unix system
	sqlQueryUsers := "create table observer.users (id int auto_increment not null, user varchar(255), type varchar(20), ln_user_id int, primary key (`id`));"
	res, err = d.db.Exec(sqlQueryUsers)

	if err != nil {
		return err
	}

	log.Println("Created table users")

	return nil
}
