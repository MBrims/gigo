package main

import (
	"database/sql"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"

	// "path"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

var DatabaseInfo = struct {
	host     string
	port     int
	user     string
	password string
	dbname   string
	db       *sql.DB
	filePath string
}{
	host:     "localhost",
	port:     5432,
	user:     "postgres",
	password: "1234",
	dbname:   "gigo",
	filePath: "./stored_files/",
}

var validPath = regexp.MustCompile(`^\/file\?file=[.a-zA-Z0-9]*$`)
var validDownloadRequest = regexp.MustCompile(`^\/download\?file=[.a-zA-Z0-9]*`)

var tableInfo = struct {
	tableName string
	id        string
	fileName  string
	filePath  string
}{
	tableName: "stored_files",
	id:        "Id",
	fileName:  "file_name",
	filePath:  "file_path",
}

func PanicOnError(err error) {
	if err != nil {

		// Close the DB if we're quitting out
		if DatabaseInfo.db != nil {
			DatabaseInfo.db.Close()
		}

		panic(err)
	}
}

func handleFile(w http.ResponseWriter, r *http.Request) {
	// Check to see there's a file and a query.
	if validPath.MatchString(r.URL.RequestURI()) {
		// Do the magic in here I guess, for now we just want to serve the test files...
		query := r.URL.Query().Get("file")

		// this is adding a header for the suggestion of a filename.
		w.Header().Add("Content-Disposition", "attachment; filename="+query)
		http.ServeFile(w, r, "./testFiles/"+query)
		return
	}

	http.Error(w, "No file found.", http.StatusBadRequest)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	// Check to see there's a file and a query.
	if validDownloadRequest.MatchString(r.RequestURI) {
		// Do the magic in here I guess, for now we just want to serve the test files...
		query := r.URL.Query().Get("file")

		// this is adding a header for the suggestion of a filename.
		w.Header().Add("Content-Disposition", "attachment; filename="+query)

		getfile := filepath.Join(DatabaseInfo.filePath, query)
		//http.ServeFile(w, r, DatabaseInfo.filePath+"\\"+query)
		http.ServeFile(w, r, getfile)
		return
	}

	http.Error(w, "No file found.", http.StatusBadRequest)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// We're expecting 2 substrings
	split := strings.SplitAfterN(r.Header.Get("Content-Disposition"), "=", 2)
	filename := "Noname"
	if len(split) > 1 {
		filename = split[1]
	}

	//filename := r.Header.Get("Content-Disposition")

	// Grab the blob to write to a file.
	blob := make([]byte, r.ContentLength)

	len, err := io.ReadFull(r.Body, blob)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Read %d from %s of anticipated %d\n", len, filename, r.ContentLength)

	// Create the id, filepath and the actual file here.
	id := rand.Int64()
	s_id := fmt.Sprintf("%d", id)

	filePath := filepath.Join(DatabaseInfo.filePath, s_id)
	// filepath := fmt.Sprintf("%s\\%d", DatabaseInfo.filePath, id)

	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write to the file.
	len, err = file.Write(blob)
	file.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Wrote %d to %s\n", len, filePath)

	/*
			Id VARCHAR(255),
		        file_name VARCHAR(255) NOT NULL,
		        file_path VARCHAR(512) NOT NULL UNIQUE,
		        PRIMARY KEY (Id)
	*/

	// Finally add it the database as (in theory) everything is A-OK.
	const insert = `INSERT INTO stored_files (Id, file_name, file_path) VALUES ($1, $2, $3)`
	_, err = DatabaseInfo.db.Exec(insert, id, filename, filePath)

	if err != nil {

		// We aught to clean up after ourselves in case it fails.
		osErr := os.Remove(filePath)
		if osErr != nil {
			fmt.Println("Failed to remove file: ", osErr)
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(s_id))
}

// Panics on error
func InitDB() {
	// Connect to DB
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DatabaseInfo.host, DatabaseInfo.port, DatabaseInfo.user, DatabaseInfo.password, DatabaseInfo.dbname)

	db, dbErr := sql.Open("postgres", psqlconn)
	PanicOnError(dbErr)

	// Check the connection
	dbErr = db.Ping()
	PanicOnError(dbErr)

	// Checking + Creating table. This could be moved to outside setup, but it's pretty quick (I think), so eh.
	// Check to see if we've created a table.
	tableq, dbErr := db.Query("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = $1)", tableInfo.tableName)
	PanicOnError(dbErr)

	defer tableq.Close()
	table_alive := false
	tableq.Next()
	tableq.Scan(&table_alive)

	// If not, set it up.
	if !table_alive {
		tableCreateQ := fmt.Sprintf(
			"CREATE TABLE %[1]s ( %[2]s VARCHAR(255), %[3]s VARCHAR(255) NOT NULL, %[4]s VARCHAR(512) NOT NULL UNIQUE, PRIMARY KEY (%[2]s))",
			tableInfo.tableName,
			tableInfo.id,
			tableInfo.fileName,
			tableInfo.filePath)

		_, dbErr := db.Exec(tableCreateQ)

		PanicOnError(dbErr)
		fmt.Println("Created new table stored_files")
	}

	// Time to keep track of it.
	DatabaseInfo.db = db
}

func InitHandles() {
	http.HandleFunc("GET /file", handleFile)

	http.HandleFunc("GET /download", handleDownload)
	http.HandleFunc("POST /upload", handleUpload)

	// Probably want the serving folder to be more specific?
	fs := http.FileServer(http.Dir("dist"))
	http.Handle("GET /", fs)
}

func main() {
	const handle string = "localhost:8080"

	// A shitty work around to ensure the file path is correct.
	DatabaseInfo.filePath, _ = filepath.Abs(DatabaseInfo.filePath)

	// Initialise the DB
	InitDB()
	defer DatabaseInfo.db.Close()

	// Create our storage path if it doesn't exist.
	err := os.MkdirAll(DatabaseInfo.filePath, os.ModePerm)
	PanicOnError(err)

	// Initialise default HTTP handles.
	InitHandles()

	// Spin it up.
	fmt.Printf("Listening on http://%s\n", handle)
	PanicOnError(http.ListenAndServe(handle, nil))
}
