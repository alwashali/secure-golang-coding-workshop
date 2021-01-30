package loger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logit logs important event to log file
func Logit(logContent string) {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	log := fmt.Sprintf("%s %s \n", timestamp, logContent)

	file.WriteString(log)

}
