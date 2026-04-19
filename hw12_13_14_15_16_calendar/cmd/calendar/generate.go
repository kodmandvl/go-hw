package main

// Директива go:generate вызывается из каталога cmd/calendar (так работает go generate).
// Относительные пути ведут к корню модуля.
//
//go:generate sh -c "mkdir -p ../../internal/pb && protoc -I ../../api/proto --go_out=../../internal/pb --go_opt=paths=source_relative --go-grpc_out=../../internal/pb --go-grpc_opt=paths=source_relative calendar/v1/calendar.proto"
