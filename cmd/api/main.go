package main

import (
	"fmt"
	"os"

	"github.com/faqihyugos/coffee-pos/config"
	"github.com/faqihyugos/coffee-pos/internal/handler"
	"github.com/faqihyugos/coffee-pos/pkg/database"
	pkgredis "github.com/faqihyugos/coffee-pos/pkg/redis"
	pkgvalidator "github.com/faqihyugos/coffee-pos/pkg/validator"
)

func main() {
	fmt.Println("Coffee Shop POS starting...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	db, err := database.NewMySQL(cfg.MysqlDSN())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("MySQL connected.")

	rdb, err := pkgredis.NewRedis(cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer rdb.Close()
	fmt.Println("Redis connected.")

	v := pkgvalidator.New()
	r := handler.NewRouter(db, cfg, v)

	fmt.Println("Starting server on port :" + cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
