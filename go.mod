module github.com/hrapovd1/final_project

go 1.15

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/hrapovd1/final_project/pkg/smgrpc v0.0.0-20210131153301-b555f914dff4
	github.com/hrapovd1/final_project/pkg/sysmon v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20210126160654-44e461bb6506 // indirect
	google.golang.org/grpc v1.35.0
)

replace github.com/hrapovd1/final_project/pkg/sysmon => ./pkg/sysmon
