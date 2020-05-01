package plog4go

import (

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUsage(t *testing.T) {
	// Get a logger, log to default.log. Default root path ./run-log
	logger := GetLogger(DefaultLogFile)
	// Get a logger, log to stat.log. File path: ./run-log/stat.log
	statLogger := GetLogger("stat.log")

	// api same like go log!
	logger.Println("Hello PLog4Go!")
	statLogger.Printf("Hello %s!", "PLog4Go")
	statLogger.Print("Hello", " PLog4Go!")

	//logger.Panicln("don't panic!")
	//logger.Panicf("don't panic!")
	//logger.Panic("don't panic!")
	//
	//logger.Fatalln("fatal!")
	//logger.Fatalf("fatal!")
	//logger.Fatal("fatal!")
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger(DefaultLogFile)
	logger.Println("logger get!")

	assert.True(t, logger != nil)
}
