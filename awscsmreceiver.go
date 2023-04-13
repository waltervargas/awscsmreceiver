// Package awscsmreceiver provides functionality for parsing and processing
// AWS Client Side Monitoring (CSM) messages received via UDP.
package awscsmreceiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// CSMMessage represents a single AWS Client Side Monitoring (CSM) message.
type CSMMessage struct {
	Api                 string `json:"Api"`
	Type                string `json:"Type"`
	Region              string `json:"Region"`
	Service             string `json:"Service"`
	AccessKey           string `json:"AccessKey"`
	UserAgent           string `json:"UserAgent"`
	XAmznRequestId      string `json:"XAmznRequestId"`
	Timestamp           int    `json:"Timestamp"`
	Attempts            int    `json:"AttemptCount"`
	Latency             int    `json:"Latency"`
	Version             int    `json:"Version"`
	HttpStatusCode      int    `json:"HttpStatusCode"`
	FinalHttpStatusCode int    `json:"FinalHttpStatusCode"`
	MaxRetriesExceeded  int    `json:"MaxRetriesExceeded"`
}

// ParseCSMMessage parses a JSON payload string into a CSMMessage struct.
// It returns an error if the payload is empty or unmarshaling fails.
func ParseCSMMessage(payload string) (CSMMessage, error) {
	if payload == "" {
		return CSMMessage{}, fmt.Errorf("unable to parse an empty payload")
	}

	var msg CSMMessage
	err := json.Unmarshal([]byte(payload), &msg)
	if err != nil {
		return CSMMessage{}, err
	}

	return msg, nil
}

// MessageHandler is a function type that takes a CSMMessage as input.
type MessageHandler func(CSMMessage)

// ListenAndServe starts a UDP server that listens on the given address
// and processes incoming CSM messages using the provided MessageHandler.
// It returns an error if the address cannot be resolved or the connection
// cannot be established.
func ListenAndServe(addr string, handler MessageHandler) error {
	// Resolve the address and create a UDP connection
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create a buffer to receive the data
	buf := make([]byte, 1024)

	for {
		// Read the incoming data into the buffer
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		// Parse the CSM message
		msg, err := ParseCSMMessage(string(buf[:n]))
		if err != nil {
			continue
		}

		// Process the message using the provided handler
		handler(msg)
	}
}

// WriteCSVHandler returns a function that takes a CSMMessage as input and writes it as a CSV line
// to the provided io.Writer. The returned function can be used as a MessageHandler for the
// ListenAndServe function. The CSV header is written to the io.Writer before returning the function.
//
// Usage:
//
//	buf := new(bytes.Buffer)
//	csvWriter := WriteCSVHandler(buf)
//	msg := CSMMessage{...}
//	csvWriter(msg)
//
// Parameters:
//   - buf: An io.Writer to which the CSV header and CSMMessage data will be written.
//
// Returns:
//   - A function that takes a CSMMessage as input and writes its data as a CSV line to the provided io.Writer.
func WriteCSVHandler(buf io.Writer) func(CSMMessage) {
	csvHead := "Type,Region,Service,Api,XAmznRequestId,Attempts,Latency,Timestamp,Version,HttpStatusCode,FinalHttpStatusCode,MaxRetriesExceeded"
	fmt.Fprintln(buf, csvHead)
	return func(msg CSMMessage) {
		fmt.Fprintf(buf, "%s,%s,%s,%s,%s,%d,%d,%d,%d,%d,%d,%d\n",
			msg.Type,
			msg.Region,
			msg.Service,
			msg.Api,
			msg.XAmznRequestId,
			msg.Attempts,
			msg.Latency,
			msg.Timestamp,
			msg.Version,
			msg.HttpStatusCode,
			msg.FinalHttpStatusCode,
			msg.MaxRetriesExceeded,
		)
	}
}
