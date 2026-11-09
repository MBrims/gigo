"Compiling Client"
cd ./client
npm run build
cd ../

"Compiling Server"
go build -v -x -o server.exe

"Complete; Starting server..."
./server.exe
# "C:\Program Files\LibreWolf\librewolf.exe" -P "Debug" -start-debugger-server 6000 