# Task

#### What
Task is a simple go library that does the following:
- Allows you to create a simple task and execute it

#### Use Cases
Some use cases:
- Execute build like command operations where you want to capture the output
- ...

#### Example
```go
package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/knspriggs/task"
)

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.JSONFormatter)
}

func main() {
	myTask := task.NewTask(
		[]string{"echo \"start complex\"", "sleep 2", "echo \"end of complex\""},
		"JobName",
		"Owner",
	)
	logChannel, errChannel, returnCodeChannel := myTask.Execute()
	for l := range logChannel { //LogMessage structs will be returned (each line of output from command)
		log.Info(l)
	}
	err := <-errChannel //any errors will be returned on this channel
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Return code:", <-returnCodeChannel) //The return code is also available
}
```
