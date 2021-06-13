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
2021/06/01 11:48:26.709755 plog4go_test.go:20: [WARN] WARN. Should see this.
2021/06/01 11:48:26.709813 plog4go_test.go:21: [ERROR] ERROR. Should see this.
```

#### Example 2. Start trace log.
Code
```go
// Use start trace to open trace log.
testLogger.StartTrace()
testLogger.Trace("Opened trace. Should see this.")
testLogger.StopTrace()
testLogger.Trace("Opened trace. Shouldn't see this.")
```

Output looks
```text
// ./logs/test.log
2021/06/01 11:48:26.709844 plog4go_test.go:24: [TRACE] Opened trace. Should see this.
```

#### Example 3. Redirect stdOut/stdErr to file.
Code
```go
    // Use start trace to open trace log.
	testLogger.StartTrace()
	testLogger.Trace("Opened trace. Should see this.")
	testLogger.StopTrace()
	testLogger.Trace("Opened trace. Shouldn't see this.")
```

Output looks
```text
// ./logs/stdout.log
To console log. should see in file.
```

## TODO
- more conf
