package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"telegram-pusher/database"
	"telegram-pusher/pusher"
	"telegram-pusher/serv"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("START")

	log.Println("runtime.GOMAXPROCS:", runtime.GOMAXPROCS(0))

	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Cant load .env: ", err)
	}

	// telegramReportGroupID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_GROUP_ID"), 10, 64)

	mysqlURL := os.Getenv("MYSQL_URL")

	if os.Getenv("MYSQL_URL_FILE") != "" {
		mysqlURL_, err := os.ReadFile(os.Getenv("MYSQL_URL_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		mysqlURL = strings.TrimSpace(string(mysqlURL_))
	}

	dbService, err := database.NewService(mysqlURL)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("dbService OK")
	}

	_, err = pusher.NewService(dbService)
	if err != nil {
		log.Fatal(err)
	}

	// telegramService.Send(telegramReportGroupID, fmt.Sprintf("dmca started"))

	httpService, err := serv.NewService(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	httpService.Run()
}
