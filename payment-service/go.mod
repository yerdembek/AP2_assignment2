module github.com/yerdembek/AP2_assignment2/payment-service

go 1.22

require (
	// Используем твой путь к сгенерированному коду
	github.com/yerdembek/AP2_assignment2/generated v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
)

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

// Указываем Go искать этот модуль в соседней папке generated
replace github.com/yerdembek/AP2_assignment2/generated => ../generated