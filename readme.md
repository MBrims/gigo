# --- WIP --- #

Currently this is just a way to get some experience with React and Golang.


## Building & Running
Presumably both Node and Go(lang) are installed. Call **npm install** in *./client* and **go install** in root. Running **run.ps1** will build the server and start it up. It expects the **Postgres** service to be running already.

serverOptions.json specifies several options that can be modified. It's read in once at the very beginning of the servers lifetime.