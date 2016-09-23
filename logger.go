package main

import (
	"fmt"
	"os"
)

type Logger struct {
	Verbose bool
}

var log = Logger{
	Verbose: false,
}

func (l *Logger) Debug(v ...interface{}) {
	if l.Verbose {
		fmt.Print("DEBUG: ")
		fmt.Println(v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	fmt.Print("ERROR: ")
	fmt.Println(v...)
	os.Exit(1)
}
