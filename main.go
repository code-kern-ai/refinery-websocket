package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"database/sql"

	_ "github.com/lib/pq"
)

var addr = flag.String("addr", ":8080", "http service address")

var (
	Logger *log.Logger
)

func getDbConnString() string {
	dsn := os.Getenv("DB_DSN")
	dsnAtSplit := strings.Split(dsn, "@")
	dsnUserPassSplit := strings.Split(strings.Replace(dsnAtSplit[0], "postgresql://", "", 1), ":")
	dsnHostPortSplit := strings.Split(dsnAtSplit[1], "/")
	dsnHostSplit := strings.Split(dsnHostPortSplit[0], ":")
	dsnDbParamsSplit := strings.Split(dsnHostPortSplit[1], "?")
	dsnParamsSplit := strings.Split(dsnDbParamsSplit[1], "&")
	
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s %s",
    	dsnHostSplit[0], 5432, dsnUserPassSplit[0], dsnUserPassSplit[1], dsnDbParamsSplit[0], strings.Join(dsnParamsSplit, " "))
}

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
	psqlInfo := getDbConnString()
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
