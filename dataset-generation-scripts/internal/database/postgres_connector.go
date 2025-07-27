package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func GetPostgresDBConnection() *bun.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pgUsername := os.Getenv("POSTGRES_DB_USERNAME")
	pgPswd := os.Getenv("POSTGRES_DB_PSWD")
	pgHost := os.Getenv("POSTGRES_DB_HOST")
	pgPort := os.Getenv("POSTGRES_DB_PORT")
	pgDbName := os.Getenv("POSTGRES_TMP_DB_NAME")

	dsn :=  fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUsername, pgPswd, pgHost, pgPort, pgDbName)
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqlDB, pgdialect.New())
	return db
}
