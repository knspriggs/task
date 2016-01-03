package task

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/juju/deputy"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	log           = logrus.New()
	tempDirectory = os.Getenv("TEMP_DIR")
)

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	if tempDirectory == "" {
		tempDirectory = "/tmp"
	}
}

// Task struct for work details performed by workers.
type Task struct {
	ID          string   `gorethink:"id,omitempty",json:"id,omitempty"`
	Commands    []string `gorethink:"commands",json:"commands"`
	Job         string   `gorethink:"job",json:"job"`
	Owner       string   `gorethink:"owner",json:"owner"`
	CommandFile string   `gorethink:"command_file",json:"command_file"`
}

// LogMessage struct for log messages being stored in database.
type LogMessage struct {
	ID        string    `gorethink:"id,omitempty",json:"id,omitempty"`
	Message   string    `gorethink:"message",json:"message"`
	Timestamp time.Time `gorethink:"timestamp",json:"timestamp"`
	Job       string    `gorethink:"job",json:"job"`
	Owner     string    `gorethink:"owner",json:"owner"`
}

// NewTask creates a new task with commands, job name, and owner name.
func NewTask(c []string, j string, o string) *Task {
	task := Task{
		Commands: c,
		Job:      j,
		Owner:    o,
	}
	return &task
}

// Execute runs the commands listed in the task's commands slice.
func (t *Task) Execute() (chan LogMessage, chan error, chan int) {
	cf := filepath.Join(tempDirectory, generateRandomString(12))
	createCommandFile(t, cf, t.Commands)
	t.CommandFile = fmt.Sprintf("%s.sh", cf)
	logChannel := make(chan LogMessage, 100)
	errChannel := make(chan error, 100)
	returnCodeChannel := make(chan int, 1)
	go func(logChannel chan LogMessage, errChannel chan error) {
		d := deputy.Deputy{
			Errors: deputy.FromStderr,
			StdoutLog: func(b []byte) {
				logChannel <- LogMessage{
					Message:   string(b),
					Timestamp: time.Now(),
					Job:       t.Job,
					Owner:     t.Owner,
				}
			},
		}
		if err := d.Run(exec.Command(t.CommandFile)); err != nil {
			returnCodeChannel <- 1
			log.WithFields(logrus.Fields{
				"application": "conductor",
				"component":   "task",
				"err":         err.Error(),
				"job":         t.Job,
				"owner":       t.Owner,
			}).Info("Command failed")
			close(logChannel)
			errChannel <- err
			close(errChannel)
		} else {
			returnCodeChannel <- 0
			errChannel <- nil
			close(logChannel)
			close(errChannel)
			cleanupCommandFile(cf)
		}
	}(logChannel, errChannel)
	return logChannel, errChannel, returnCodeChannel
}

func generateRandomString(size int) string {
	rb := make([]byte, size)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)
	return rs
}

func createCommandFile(t *Task, fp string, commands []string) {
	if commands[0] != "#!/bin/bash" {
		commands = append([]string{"#!/bin/bash"}, commands...)
	}
	command := strings.Join(commands, "\n")
	commandFile := fmt.Sprintf("%s.sh", fp)
	log.WithFields(logrus.Fields{
		"application": "conductor",
		"component":   "task",
		"job":         t.Job,
		"owner":       t.Owner,
	}).Info("Created command file:", commandFile)
	ioutil.WriteFile(commandFile, []byte(command), 0777)
}

func cleanupCommandFile(fp string) {
	err := os.Remove(fmt.Sprintf("%s.sh", fp))
	if err != nil {
		log.Error("Unable to remove command file")
	}
}
