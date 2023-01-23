module github.com/onlineconf/onlineconf/admin/go

go 1.13

require (
	github.com/bmatcuk/doublestar/v4 v4.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/rs/zerolog v1.28.0
	github.com/ugorji/go/codec v1.1.7
	github.com/ugorji/go/codec/codecgen v1.1.7
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/bmatcuk/doublestar/v4 v4.4.0 => github.com/AndrewDeryabin/doublestar/v4 v4.0.0-20230123130924-38953b2ce9a0
