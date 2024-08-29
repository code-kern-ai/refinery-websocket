package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"database/sql"

	_ "github.com/lib/pq"
)

var addr = flag.String("addr", ":8080", "http service address")

var (
	Logger *log.Logger
	host string     = os.Getenv("DB_HOST")
	port string     = os.Getenv("DB_PORT")
	user string     = os.Getenv("DB_USER")
	password string = os.Getenv("DB_PASSWORD")
	dbname string   = os.Getenv("DB_NAME")
	params string   = os.Getenv("DB_PARAMS")
)

func main() {
	Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Starting server") 

	flag.Parse()
	hub := newHub()
	go hub.run(Logger)
	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		notify(hub, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Listening on :8080")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func getOrganizationId(userId string) string {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s %s",
    	host, port, user, password, dbname, params)
	db, dberr := sql.Open("postgres", psqlInfo)
	if dberr != nil {
		log.Fatal("Failed to open a DB connection: ", dberr)
	}
	defer db.Close()

	var organizationId string
	sql := "SELECT organization_id from public.user where id = $1"
	sqlerr := db.QueryRow(sql, userId).Scan(&organizationId)
	if sqlerr != nil {
		log.Fatal("Failed to execute query: ", sqlerr)
	}

	return organizationId
}
