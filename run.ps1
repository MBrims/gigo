"Compiling Client"
cd ./client
npm run build
cd ../

"Compiling Server"
go build -v -o server.exe

# This section is just 'cause I keep forgetting to start the service...
"Checking PostGres"
if ( ( Get-Service -Name postgresql-x64-17 ).Status -eq "Stopped" ) { 
    try
    {
    "Starting up PostGres service..."
     Start-Service -Name postgresql-x64-17 -ErrorAction Stop
    }
    catch 
    {
        "Failed to start... Try running as Admin."
        exit
    }

}

"Complete; Starting server..."
./server.exe
# "C:\Program Files\LibreWolf\librewolf.exe" -P "Debug" -start-debugger-server 6000 