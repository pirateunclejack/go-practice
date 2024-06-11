package helpers

import (
	"log"

	"github.com/joho/godotenv"
)

func InitHelper(){
    err := godotenv.Load()
    if err != nil {
        log.Fatal("failed to load .env file: ", err.Error())
        return
    }
}
