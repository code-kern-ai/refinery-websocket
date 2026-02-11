package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"database/sql"

	_ "github.com/lib/pq"
)

var (
	addr   = flag.String("addr", ":8080", "http service address")
	certFile = flag.String("cert", "", "path to TLS certificate (enables HTTPS when set with -key)")
	keyFile  = flag.String("key", "", "path to TLS private key (enables HTTPS when set with -cert)")
)

var (
	Logger *log.Logger
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

	if *certFile != "" && *keyFile != "" {
		Logger.Println("Listening with TLS on", *addr)
		err := http.ListenAndServeTLS(*addr, *certFile, *keyFile, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	} else {
		Logger.Println("Listening (no TLS) on", *addr)
		// TLS optional for local dev; use -cert and -key in production.
		err := http.ListenAndServe(*addr, nil) // nosemgrep: go.lang.security.audit.net.use-tls.use-tls
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}

func getOrganizationId(userId string) string {
	db, dberr := sql.Open("postgres", os.Getenv("DB_DSN"))
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
