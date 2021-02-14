module github.com/hrapovd1/final_project

go 1.15

replace github.com/hrapovd1/final_project/pkg/sysmon => ./pkg/sysmon

replace github.com/hrapovd1/final_project/pkg/smgrpc => ./pkg/smgrpc

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golangci/golangci-lint v1.36.0 // indirect
	github.com/hrapovd1/final_project/pkg/smgrpc v0.0.0-20210213222921-32734a35d015
	github.com/hrapovd1/final_project/pkg/sysmon v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.35.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
)
