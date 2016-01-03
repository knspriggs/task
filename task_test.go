package task

import (
	"github.com/Sirupsen/logrus"
	a "github.com/stretchr/testify/assert"

	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	commands = []string{"date"}
	job      = "TaskTest"
	owner    = "Tester"
)

func TestNewTask(t *testing.T) {
	task := NewTask(commands, job, owner)
	a.NotNil(t, task)
}

func TestSimpleCommand(t *testing.T) {
	task := NewTask(commands, job, owner)
	logChannel, errChannel, _ := task.Execute()
	date := <-logChannel
	a.NotNil(t, date)
	<-errChannel
}

func TestShellScriptCommand(t *testing.T) {
	pwd, _ := os.Getwd()
	scriptPath := filepath.Join(pwd, "./test/example.sh")
	task := NewTask(
		[]string{scriptPath},
		job,
		owner,
	)
	logChannel, errChannel, _ := task.Execute()
	for l := range logChannel {
		log.WithFields(logrus.Fields{
			"application":   "conductor_TestShellScriptCommand",
			"job":           job,
			"owner":         owner,
			"log_timestamp": l.Timestamp,
		}).Info(l.Message)
		a.NotNil(t, l)
	}
	err := <-errChannel
	a.Nil(t, err)
}

func TestComplexCommand(t *testing.T) {
	task := NewTask(
		[]string{"echo \"start complex\"", "sleep 2", "echo \"end of complex\""},
		job,
		owner,
	)
	logChannel, errChannel, _ := task.Execute()
	for l := range logChannel {
		log.WithFields(logrus.Fields{
			"application":   "conductor_TestComplexCommand",
			"job":           job,
			"owner":         owner,
			"log_timestamp": l.Timestamp,
		}).Info(l.Message)
		a.NotNil(t, l)
	}
	err := <-errChannel
	a.Nil(t, err)
}

func TestComplexCommandWithExternalCalls(t *testing.T) {
	task := NewTask(
		[]string{"ping -c 5 google.com"},
		job,
		owner,
	)
	logChannel, errChannel, _ := task.Execute()
	for l := range logChannel {
		log.WithFields(logrus.Fields{
			"application":   "conductor_TestComplexCommand",
			"job":           job,
			"owner":         owner,
			"log_timestamp": l.Timestamp,
		}).Info(l.Message)
		a.NotNil(t, l)
	}
	err := <-errChannel
	a.Nil(t, err)
}

func TestCommandReturnCodeSuccess(t *testing.T) {
	task := NewTask(
		[]string{"date"},
		job,
		owner,
	)
	logChannel, errChannel, returnCodeChannel := task.Execute()
	for l := range logChannel {
		log.WithFields(logrus.Fields{
			"application":   "conductor_TestCommandReturnCodeSuccess",
			"job":           job,
			"owner":         owner,
			"log_timestamp": l.Timestamp,
		}).Info(l.Message)
		a.NotNil(t, l)
	}
	err := <-errChannel
	a.Nil(t, err)
	a.Equal(t, 0, <-returnCodeChannel)
}

func TestInvalidCommand(t *testing.T) {
	task := NewTask(
		[]string{"tlop"},
		job,
		owner,
	)
	logChannel, errChannel, returnCodeChannel := task.Execute()
	_, ok := <-logChannel
	a.Equal(t, false, ok)
	err := <-errChannel
	a.NotNil(t, err)
	a.Equal(t, 1, <-returnCodeChannel)
}

func TestGenerateRandomString(t *testing.T) {
	arr := []string{}
	for i := 0; i < 10000; i++ {
		str := generateRandomString(12)
		if contains(arr, str) {
			a.Fail(t, "Generated two identical strings")
		}
		arr = append(arr, str)
	}
}

func TestCreateCommandFile(t *testing.T) {
	task := &Task{}
	commands := []string{"this", "is", "a", "test"}
	fp := "test/cfile"
	createCommandFile(task, fp, commands)
	if _, err := os.Stat(fmt.Sprintf("%s.sh", fp)); os.IsNotExist(err) {
		a.Fail(t, "Command file not created")
	} else {
		log.Printf("%s.sh exists", fp)
	}
	os.Remove(fmt.Sprintf("%s.sh", fp))
}

func TestCleanupCommandFile(t *testing.T) {
	task := &Task{}
	commands := []string{"this", "is", "a", "test"}
	fp := "test/cfile"
	createCommandFile(task, fp, commands)
	if _, err := os.Stat(fmt.Sprintf("%s.sh", fp)); os.IsExist(err) {
		cleanupCommandFile(fp)
		if _, err := os.Stat(fmt.Sprintf("%s.sh", fp)); os.IsExist(err) {
			a.Fail(t, "Command file not deleted")
		}
	}
}

func TestCleanupCommandFileNotExist(t *testing.T) {
	fp := "test/cfile"
	if _, err := os.Stat(fmt.Sprintf("%s.sh", fp)); os.IsExist(err) {
		cleanupCommandFile(fp)
		if _, err := os.Stat(fmt.Sprintf("%s.sh", fp)); os.IsExist(err) {
			a.Fail(t, "Command file not deleted")
		}
	}
}

//helpers
func contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}
