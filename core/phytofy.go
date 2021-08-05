// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This is the main code of the application
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Registers a handler for terminating signals
func signalHandling() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		// Closes the CLI if any terminating signal occurs
		<-signals
		fmt.Printf("\nExiting\n")
		os.Exit(0)
	}()
}

type cliFunction func(string, string, *log.Logger) (string, error)

type cliCommand struct {
	commandName         string
	argumentName        string
	argumentDescription string
	commandHandler      cliFunction
}

// Shows application usage
func showUsage(commands []cliCommand) {
	fmt.Printf("Usage: phytofy COMMAND ARGUMENTS\n")
	commandNames := make([]string, 0)
	for _, entry := range commands {
		commandNames = append(commandNames, entry.commandName)
	}
	fmt.Printf("Commands: %s\n", strings.Join(commandNames, ", "))
}

// Main dispatch of the application
func execute(logger *log.Logger) {
	commands := make([]cliCommand, 0)
	commands = append(commands, cli0Commands()...)
	commands = append(commands, cli1Commands()...)
	if len(os.Args) < 2 {
		showUsage(commands)
		os.Exit(1)
	}
	for _, entry := range commands {
		if os.Args[1] == entry.commandName {
			if len(os.Args) < 3 {
				fmt.Printf("This command requires an argument: %s (%s)\n", entry.argumentName, entry.argumentDescription)
				os.Exit(1)
			} else {
				if result, fail := entry.commandHandler(os.Args[1], os.Args[2], logger); fail != nil {
					fmt.Printf("Error: %s\n", fail)
				} else {
					fmt.Printf("%s\n", result)
				}
			}
			return
		}
	}
	showUsage(commands)
}

// Returns the base directory for logs
func logBase() string {
	return path.Join(path.Dir(os.Args[0]), "logs")
}

// Initializes logging for the application
func logInit() *log.Logger {
	var output io.Writer
	if toConsole, fail := strconv.ParseBool(os.Getenv("PHYTOFY_CONSOLE_LOGGING")); fail == nil && toConsole {
		output = os.Stdout
	} else {
		output = &lumberjack.Logger{
			Filename: path.Join(logBase(), "phytofy.log"),
			MaxSize:  100,
			MaxAge:   64,
		}
	}
	logger := log.New(output, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	return logger
}

// Collects log files in a ZIP
func logCollect(logger *log.Logger) ([]byte, error) {
	buffer := bytes.Buffer{}
	zipped := zip.NewWriter(&buffer)
	base := logBase()
	files, fail := ioutil.ReadDir(base)
	if fail != nil {
		return nil, fail
	}
	for _, file := range files {
		if !file.IsDir() {
			data, fail := ioutil.ReadFile(path.Join(base, file.Name()))
			if fail != nil {
				return nil, fail
			}
			entry, fail := zipped.Create(file.Name())
			if fail != nil {
				return nil, fail
			}
			_, fail = entry.Write(data)
			if fail != nil {
				return nil, fail
			}
		}
	}
	if fail := zipped.Close(); fail != nil {
		return nil, fail
	}
	return buffer.Bytes(), nil
}

// This if the main function for this application
func main() {
	signalHandling()
	execute(logInit())
}
