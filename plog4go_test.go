package p_log4go

import (
	"fmt"
	"log"
	"testing"
)

func TestUsage(t *testing.T) {

	// Test Log Level
	testLogger, err := GetLogger("./logs/test.log", WARN, Hourly, 3)
	if err != nil {
		fmt.Printf("%v", err)
	}

	testLogger.Trace("TRACE. Shouldn't see this.")
	testLogger.Debug("DEBUG. Shouldn't see this.")
	testLogger.Info("INFO. Shouldn't see this.")
	testLogger.Warn("WARN. Should see this.")
	testLogger.Error("ERROR. Should see this.")

	// Test trace flag
	testLogger.StartTrace()
	testLogger.Trace("Trace on. Should see this.")
	testLogger.StopTrace()
	testLogger.Trace("Trace off. Shouldn't see this.")

	// Test DEBUG level
	testLogger2, err := GetLogger("./logs/test2.log", DEBUG, Hourly, 3)
	if err != nil {
		fmt.Printf("%v", err)
	}

	testLogger2.Debug("DEBUG. Should see this.")
	testLogger2.Info("INFO. Should see this.")

	// Test std in/out
	log.Print("To console")
	log.Print("Init to file, console log will forward to file.")

	InitStd(StdOutToConf{To: ToFile, ToDir: "./logs"})
	fmt.Printf("To console log. Should see in file.")

	// Test hourly rotate
	testToFileLogger, err := GetLogger("./logs/test3.log", INFO, Hourly, 3)
	if err != nil {
		fmt.Printf("%v", err)
	}
	testToFileLogger.Trace("TRACE. Shouldn't see this.")
	testToFileLogger.Debug("DEBUG. Shouldn't see this.")
	testToFileLogger.Info("INFO. Should see this.")
	testToFileLogger.Warn("WARN. Should see this.")
	testToFileLogger.Error("ERROR. Should see this.")

	// Test Default Logger, v
	Trace("TRACE. Should see this in %s", "./logs/app.log.")
	Debug("DEBUG. Should see this in %s", "./logs/app.log.")
	Info("INFO. Should see this in %s", "./logs/app.log.")
	Warn("WARN. Should see this in %s", "./logs/app.log.")
	Error("ERROR. Should see this in %s", "./logs/app.log.")

	// Test Panic And Fatal
	//Panic("PANIC...%s. Shouldn see this in %s", "P1", "./logs/app.log.")
	Fatal("Fatal...%s. Should see this in %s", "F1", "./logs/app.log.")
	Info("AfterFatal. Shouldn't see this")
}
