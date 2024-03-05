package api

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func envACCOUNTSID() string {
	println(godotenv.Unmarshal(".env"))
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("failed to load .env file to get ACCOUNTSID: %v", err)
	}
	return os.Getenv("TWILIO_ACCOUNT_SID")
}

func envAUTHTOKEN() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("faild to load .env file to get AUTHTOKEN: %v", err)
	}
	return os.Getenv("TWILIO_AUTHTOKEN")
}

func envSERVICE_ID() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("faild to load .env file to get SERVICE_ID: %v", err)
	}
	return os.Getenv("TWILIO_SERVICE_ID")
}
