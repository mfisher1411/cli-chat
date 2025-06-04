module github.com/mfisher1411/cli-chat/services/chat_server

replace github.com/mfisher1411/cli-chat/libraries/api => ../../libraries/api

go 1.24.2

require (
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/mfisher1411/cli-chat/libraries/api v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)
