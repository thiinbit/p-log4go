# PLog4Go

> A simple, original log based, file rolling logger.

## Usage

Get
```shell script
go get github.com/thiinbit/plog4go
```

Use
```go
// import
import "github.com/thiinbit/plog4go"

// Get a logger, log to default.log. Default root path ./run-log
logger := plog4go.GetLogger(plog4go.DefaultLogFile)
// Get a logger, log to stat.log. File path: ./run-log/stat.log
statLogger := plog4go.GetLogger("stat.log")

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
```

File out path (log above)
```text
app_run_dir/run-log/default.log // DefaultLogFile
app_run_dir/run-log/stat.log    // stat.log
```

Output looks
```text
// log to app_run_dir/run-log/default.log
2020/01/30 11:38:33 Hello PLog4Go!
// log to app_run_dir/run-log/stat.log
2020/01/30 11:38:33 Hello PLog4Go!
2020/01/30 11:38:33 Hello PLog4Go!
```

## TODO
- Support config (root path, file roling rule, and more...).
- Support file rolling by time.
