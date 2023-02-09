package main

import (
	"log"
	"os"
	"time"
)

const dirName = "logs"

func LogInit() error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fileLogName := time.Now().Format("2006-01-02") + ".txt"
	f, err := os.OpenFile(dirName+"/"+fileLogName, os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}
