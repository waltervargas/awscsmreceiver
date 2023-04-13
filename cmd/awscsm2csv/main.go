// Package main provides a simple command-line application to listen for AWS Client-Side
// Monitoring (CSM) messages on a specified UDP address and writes the received messages
// as CSV lines to the standard output.
package main

import (
	"log"
	"os"

	"github.com/waltervargas/awscsmreceiver"
)

// main is the entry point of the application. It sets up a UDP server to listen for
// AWS CSM messages at the specified address ("127.0.0.1:31000") and uses the WriteCSV
// function from the awscsmreceiver package to write the received messages as CSV lines
// to the standard output.
//
// If the UDP server fails to start, the program will log the error message and exit.
func main() {
	// TODO: (walter) on next iteration, main will just log ot STDOUT and on
	// Ctrl-C will write the formated CSV.
	err := awscsmreceiver.ListenAndServe("127.0.0.1:31000", awscsmreceiver.WriteCSV(os.Stdout))
	if err != nil {
		log.Fatalf("unable to start UDP server: %s\n", err)
	}
}
