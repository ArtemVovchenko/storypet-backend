module github.com/ArtemVovchenko/storypet-backend

// +heroku goVersion go1.16
// +heroku install ./...
go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible // indirect
	github.com/go-redis/redis/v7 v7.4.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/lib/pq v1.2.0
	github.com/myesui/uuid v1.0.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/twinj/uuid v1.0.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc
	golang.org/x/net v0.0.0-20210423184538-5f58ad60dda6
	gopkg.in/stretchr/testify.v1 v1.2.2 // indirect
)
