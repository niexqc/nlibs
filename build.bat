go build -o build/cpumen.exe  cpumem/main.go 


set GOARCH=amd64
set GOOS=linux
go build -o build/cpumen.so cpumem/main.go 
