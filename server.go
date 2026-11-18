package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"encoding/json/v2"

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
var validDownloadRequest = regexp.MustCompile(`^\/request\?file=[.a-zA-Z0-9]*$`)
var validStatRequest = regexp.MustCompile(`^\/statistics$`)

// Info about the sql table.
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

// Info about the server; this should be initialised once and then not touched.
// I understand that this isn't a good way of handling things, but hey, it's a small server.
var serverInfo = struct {
	ip         string        // The Ip to use.
	port       int           // The port to use.
	limitSpace bool          // If true, the server will return an error if a file causes totalAlloc to go over totalSpace.
	totalSpace uint          // Should this be something else?
	totalAlloc atomic.Uint64 // Total allocated.
}{
	ip:         "localhost",
	port:       8080,
	limitSpace: true,
	totalSpace: 10_485_760, // ~10mb
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

type StatStore struct {
	Limited  bool
	Total    uint
	Allotted uint
}

func getStats() StatStore {
	return StatStore{
		Limited:  serverInfo.limitSpace,
		Total:    serverInfo.totalSpace,
		Allotted: uint(serverInfo.totalAlloc.Load()), // This should be changed as necessary.
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if !validStatRequest.MatchString(r.URL.RequestURI()) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get any server statistics, put it into a struct, marshal it and write it.
	stats := getStats()

	bytes, err := json.Marshal(stats)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
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
		http.ServeFile(w, r, getfile) // This handles ranging already, yippee!
		return
	}

	http.Error(w, "No file found.", http.StatusBadRequest)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// We're expecting 2 substrings
	split := strings.SplitAfterN(r.Header.Get("Content-Disposition"), "=", 2)
	filename := ""
	if len(split) > 1 {
		filename = split[1]
	} else {
		http.Error(w, "No filename given.", http.StatusBadRequest)
		return
	}

	// Sanity check; might want to do more here later.
	if r.ContentLength <= 0 {
		http.Error(w, "File has no size.", http.StatusBadRequest)
		return
	}

	// TODO: Fix this mess.
	if uint(uint64(r.ContentLength)+serverInfo.totalAlloc.Load()) > serverInfo.totalSpace {
		http.Error(w, "Not enough space on server.", http.StatusRequestEntityTooLarge)
		return
	}

	initFilePath := func() (uint64, string, string) {
		__id := rand.Uint64()
		__s_id := fmt.Sprintf("%d", __id)
		__fp := filepath.Join(DatabaseInfo.filePath, __s_id)
		return __id, __s_id, __fp
	}

	var id uint64
	var s_id string
	var filePath string

	var file *os.File
	var err error

	// Ensure we don't have a clash with a pre-existing file.
	for {
		id, s_id, filePath = initFilePath()

		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0222)
		if errors.Is(err, fs.ErrExist) {
			continue
		}

		break
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If we have a file to write to, store it in the DB.
	const insert = `INSERT INTO stored_files (Id, file_name, file_path) VALUES ($1, $2, $3)`
	_, err = DatabaseInfo.db.Exec(insert, id, filename, filePath)

	if err != nil {
		file.Close()

		// We aught to clean up after ourselves in case it fails.
		osErr := os.Remove(filePath)
		if osErr != nil {
			fmt.Println("Failed to remove file: ", osErr)
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We defer down here as we previously want to manually control the close for (possible) deletion.
	defer file.Close()

	// Actually write to the file.
	len, err := file.ReadFrom(r.Body)
	fmt.Printf("Read %d from %s of anticipated %d\n", len, filename, r.ContentLength)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Wrote %d to %s\n", len, filePath)
	serverInfo.totalAlloc.Add(uint64(len))

	// Finally just return the resulting id back to the client.
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

func handleServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dist/index.html")
}

func InitHandles() {
	http.HandleFunc("GET /file", handleFile)

	http.HandleFunc("GET /request", handleDownload)
	http.HandleFunc("GET /statistics", handleStats)
	http.HandleFunc("POST /upload", handleUpload)

	// Probably want the serving folder to be more specific?

	fs := http.FileServer(http.Dir("dist"))
	// http.HandleFunc("GET /static/", GetFallbackFileServer(fs))

	http.Handle("GET /static/", fs)
	// http.Handle("GET /", fs)

	http.HandleFunc("GET /", handleServeIndex)
}

func main() {
	handle := fmt.Sprintf("%s:%d", serverInfo.ip, serverInfo.port)

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGINT) // We'll close on any of these.

	// Spin it up.
	server := &http.Server{
		Addr: handle,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Println("Shutting down complete.")
			} else {
				log.Fatalf("HTTP server error: %v", err)
			}
		}
	}()

	log.Printf("Listening on http://%s\n", handle)

	// Setup console scanning.
	stdScanner := bufio.NewScanner(os.Stdin)
	var command string

	// We'll loop until either aborted or exit is called.
	for {
		select {
		case <-sigChan:
			goto shutdown
		default:
			stdScanner.Scan()
			command = strings.TrimSpace(stdScanner.Text())
			switch command {
			case "exit":
				goto shutdown
			case "stats":
				printStats()
			default:
				if len(command) > 0 { // Sanity check for instances of: Random new lines, ^C, etc.
					fmt.Printf("Unrecognised command: %v", command)
				}
			}
		}
	}

	// Mostly just to handle the shutdown nicely.
shutdown:
	log.Println("Shutdown started.")

	timeout, timeoutRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutRelease()

	server.Shutdown(timeout)
}

// Print stats to default log.
func printStats() {
	stats := getStats()
	formatStr := "Stats\n\tUsing limited space: %v\n"
	values := []any{stats.Limited}

	if stats.Limited {
		formatStr += "\tTotal space allowed: %v\n"
		values = append(values, stats.Total)
	}

	formatStr += "\tTotal allotted: %v\n"
	values = append(values, stats.Allotted)

	log.Printf(formatStr, values...)
}
