package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Account struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"user_name"`
}

var account Account
var db *sql.DB
var err error

func init() {

	// TODO env vars
	dbHost := os.Getenv("DBHOST")                  // localhost
	dbPort, _ := strconv.Atoi(os.Getenv("DBPORT")) // 5432
	dbName := os.Getenv("DBNAME")                  // "go_db"
	dbUser := os.Getenv("DBUSER")                  // "postgresadmin"
	dbPass := os.Getenv("DBPASS")                  // "admin123"
	// === db connection===
	dbConnStr := fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%d sslmode=disable", dbName, dbUser, dbPass, dbHost, dbPort)
	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("DB connection used:", dbConnStr)
		log.Fatal("DB unreachable:", err)
	}
	// ˆˆˆdb connectionˆˆˆ
}

func main() {
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/accounts", createAccountHandler).Methods("POST")
	r.HandleFunc("/accounts", account.updateAccountHandler).Methods("PUT")
	r.HandleFunc("/accounts/{id}", getAccountHandler).Methods("GET")
	r.HandleFunc("/accounts/{id}", deleteAccountHandler).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// updateAccountHandler updates account entry. Defined as method to pass account.ID between handlers
func (account Account) updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Request received to update Account")
	json.NewDecoder(r.Body).Decode(&account) // parse json into 'account' struct

	// db query
	updStmt := `UPDATE account SET user_name = $1, first_name = $2, last_name = $3 WHERE id = $4`
	res, err := db.Exec(updStmt, account.UserName, account.FirstName, account.LastName, account.ID)
	if err != nil {
		log.Fatal("ERROR 1: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully updated account, affected count = %d\n", rowCnt)
	// response

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(account)
}

// createAccountHandler creates account entry
func createAccountHandler(w http.ResponseWriter, r *http.Request) {
	var count int // count of returned rows

	log.Print("Request received to create an Account")
	json.NewDecoder(r.Body).Decode(&account) // parse json into 'account' struct

	// check if account not present
	rows := db.QueryRow("select count(*) from account where id=$1", account.ID)
	err := rows.Scan(&count)
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// update if present
	if count != 0 {
		log.Print("INFO: account already present, updating...")
		account.updateAccountHandler(w, r)
		return
	}

	// db query
	insertStmt, err := db.Prepare("INSERT INTO account(id, user_name,first_name, last_name) VALUES($1, $2, $3, $4)")
	if err != nil {
		log.Fatal(err)
	}
	defer insertStmt.Close()

	res, err := insertStmt.Exec(account.ID, account.UserName, account.FirstName, account.LastName)
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully added account, affected count = %d\n", rowCnt)
	// response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// getAccountHandler retrieves account entry
func getAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account
	params := mux.Vars(r)
	id := params["id"]
	log.Print("Request received to get an account by account id: ", id)
	// db query
	rows, err := db.Query("select * from account where id=$1", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// parse result
	for rows.Next() {
		err := rows.Scan(&account.ID, &account.UserName, &account.FirstName, &account.LastName)
		if err != nil {
			log.Fatal(err)
		}
	}

	if (account == Account{}) {
		log.Print("Requested account not found for account id: ", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Print("Successfully retrieved account : ", account)
	// response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

// deleteAccountHandler deletes account entry
func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	log.Print("Request received to delete an account by account id: ", id)
	// db query
	delStmt, err := db.Prepare("delete from account where id=$1")
	if err != nil {
		log.Fatal(err)
	}
	defer delStmt.Close()
	// run it
	res, err := delStmt.Exec(id)
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// check result
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal("ERROR: %v ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rowCnt == 0 {
		log.Print("INFO: no rows affected")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Print("Successfully deleted account id: ", id)
	// OK response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
