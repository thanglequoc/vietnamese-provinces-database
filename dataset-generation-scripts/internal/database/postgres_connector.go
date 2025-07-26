package common

import (
	"context"
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

// Useful method to execute SQL script located in this project
func ExecuteSQLScript(pathToSQL string) error {
	bytesVal, err := os.ReadFile(pathToSQL)
	if err != nil {
		panic(err)
	}
	query := string(bytesVal)
	db := GetPostgresDBConnection()
	ctx := context.Background()
	_, err = db.ExecContext(ctx, query)
	ctx.Done()
	if err != nil {
		return err
	}
	return nil
}
