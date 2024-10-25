package instrumentation

import "time"

type Instrumentation interface {
	Debug(msg string, args map[string]interface{})
	Info(msg string, args map[string]interface{})
	Warn(msg string, args map[string]interface{})
	Error(msg string, args map[string]interface{})
	TimeTheFunction(start time.Time, functionName string)
}
