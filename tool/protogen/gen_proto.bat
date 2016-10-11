@echo off
set PATH_AGENT="./export/agent/"
set PATH_GAME="./export/game/"
set PATH_CLIENT="./export/client/"

go get gopkg.in/urfave/cli.v2

::api.txt
go run ./api/api.go --pkgname agent --min 0 --max 1000 > %PATH_AGENT%/api.go; go fmt %PATH_AGENT%/api.go
go run ./api/api.go --pkgname game --min 1001 --max 32767 > %PATH_GAME%/api.go; go fmt %PATH_GAME%/api.go
go run ./api/api.go --template "templates/client/api.tmpl"  --min 0 --max 32767 > %PATH_CLIENT%/NetApi.cs

:: proto.txt
go run ./proto/proto.go --pkgname agent > %PATH_AGENT%/proto.go; go fmt %PATH_AGENT%/proto.go
go run ./proto/proto.go --pkgname game > %PATH_GAME%/proto.go; go fmt %PATH_GAME%/proto.go
go run ./proto/proto.go --template "templates/client/proto.tmpl" --binding "cs" > %PATH_CLIENT%/NetProto.cs

pause