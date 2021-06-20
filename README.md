# PLog4Go

> A simple, original log based, file rolling logger. It's looks like time based file rotate in log4j.

Rolling log as: 
- Daily(One file per day) -> app.log.2021-06-13   
- Hourly(One file per hour) -> app.log.2021-06-13_23   
- Weekly(One file per 7 day) - > app.log.2021-06-13   


## Usage

### Import
```hell script
...
import (
  "github.com/thiinbit/p-log4go" // imports as package "PLog4Go"
)
...
```

### Code Examples
#### Example 1. Log level.
Code
```go
// GetLogger (${FileFullPath}, ${LogLevel}, ${RotateInterval}, ${RetainLogFileCount})
testLogger, err := GetLogger("./logs/test.log", WARN, Hourly, 3)
if err != nil {
    fmt.Printf("%v", err)
}

testLogger.Debug("DEBUG. shouldn't see this.")
testLogger.Info("INFO. Shouldn't see this.")
testLogger.Warn("WARN. Should see this.")
testLogger.Error("ERROR. Should see this.")
```

Output looks
```text
// ./logs/test.log
[WARN] 2021/06/20 17:20:34.787450 plog4go_test.go:20: WARN. Should see this.
[ERROR] 2021/06/20 17:20:34.787792 plog4go_test.go:21: ERROR. Should see this.
```

#### Example 2. Start trace log.
Code
```go
// Use start trace to open trace log.
testLogger.StartTrace()
testLogger.Trace("Trace on. Should see this.")
testLogger.StopTrace()
testLogger.Trace("Trace off. Shouldn't see this.")
```

Output looks
```text
// ./logs/test.log
[TRACE] 2021/06/20 17:20:34.787828 plog4go_test.go:25: Trace on. Should see this.
```

#### Example 3. Redirect stdOut/stdErr to file.
Code
```go
    // Use std to file
	InitStd(StdOutToConf{To: ToFile, ToDir: "./logs"})
	fmt.Printf("To console log. Should see in file.")
```

Output looks
```text
// ./logs/stdout.log // Auto rotate file when app restart. Filename like stdout.log.20210620.172034
To console log. should see in file.
```

## TODO
- more conf
- and more
