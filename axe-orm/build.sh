goimports -w .
gofmt -w .
go build -o c.out main.go
./c.out
