module github.com/yerdembek/AP2_assignment2/client-cli

go 1.22

require (
	// Исправленный путь к сгенерированному коду
	github.com/yerdembek/AP2_assignment2/generated v0.0.0
	google.golang.org/grpc v1.64.0
)

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

// Указываем Go искать сгенерированный код локально
replace github.com/yerdembek/AP2_assignment2/generated => ../generated
