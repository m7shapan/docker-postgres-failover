package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"golang.yandex/hasql"
	"golang.yandex/hasql/checkers"
)

type Person struct {
	ID        int    `db:"personid"`
	FirstName string `db:"firstname"`
	LastName  string `db:"lastname"`
	Address   string `db:"address"`
	City      string `db:"city"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	testConnectToDBThroughHAP()
	// testConnectToMultipleHosts()
	// hasqlFun()
}

func testConnectToDBThroughHAP() {
	db, err := sqlx.Connect("postgres", "host=localhost port=5433 user=postgres password=password dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		people := []Person{}
		err := db.Select(&people, "SELECT * FROM persons ORDER BY firstname ASC")
		fmt.Println(time.Now().Minute(), time.Now().Second(), people, err)
		time.Sleep(time.Millisecond * 100)
	}
}

func testConnectToMultipleHosts() {
	DATABASE_URL := "postgres://postgres:password@localhost:6432,localhost:6433/postgres"
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	for {
		person := Person{}
		err = conn.QueryRow(context.Background(), "SELECT personid, firstname, lastname FROM persons ORDER BY firstname ASC").Scan(&person.ID, &person.FirstName, &person.LastName)
		fmt.Println(time.Now().Second(), person, err)
		time.Sleep(time.Millisecond * 100)
	}
}

func hasqlFun() {
	master, err := sql.Open("postgres", "host=localhost port=6432 user=postgres password=password dbname=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	slave, err := sql.Open("postgres", "host=localhost port=6433 user=postgres password=password dbname=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	cl, err := hasql.NewCluster(
		[]hasql.Node{hasql.NewNode("master", master), hasql.NewNode("slave", slave)},
		checkers.PostgreSQL,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer cl.Close()

	for {
		node := cl.PrimaryPreferred()

		if node == nil {
			panic("no primary nor secondary")
		}

		fmt.Println("Node address", node.Addr())

		person := Person{}
		err = node.DB().QueryRow("SELECT personid, firstname, lastname FROM persons ORDER BY firstname ASC").Scan(&person.ID, &person.FirstName, &person.LastName)
		fmt.Println(time.Now().Second(), person, err)
		time.Sleep(time.Millisecond * 100)
	}
}
