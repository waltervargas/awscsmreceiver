package awscsmreceiver_test

import (
	"bytes"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/waltervargas/awscsmreceiver"
)

func TestParseCSMMessage(t *testing.T) {
	t.Parallel()
	input := `{
		"ClientId": "",
		"Api": "ListRoles",
		"Service": "IAM",
		"Timestamp": 1681236061717,
		"Type": "ApiCall",
		"AttemptCount": 1,
		"Latency": 817,
		"UserAgent": "APN/1.0 HashiCorp/1.0 Terraform/1.1.7 (+https://www.terraform.io) terraform-provider-aws/4.62.0 (+https://registry.terraform.io/providers/hashicorp/aws) aws-sdk-go/1.44.237 (go1.19.7; linux; arm64)",
		"Region": "eu-central-1",
		"XAmznRequestId": "c14c9ae3-ed1a-3382-75c1-765270f6922a",
		"FinalHttpStatusCode": 200,
		"MaxRetriesExceeded": 0
	}`

	want := awscsmreceiver.CSMMessage{
		Api:                 "ListRoles",
		Service:             "IAM",
		Type:                "ApiCall",
		Region:              "eu-central-1",
		Attempts:            1,
		Latency:             817,
		XAmznRequestId:      "c14c9ae3-ed1a-3382-75c1-765270f6922a",
		FinalHttpStatusCode: 200,
		Timestamp:           1681236061717,
		UserAgent:           "APN/1.0 HashiCorp/1.0 Terraform/1.1.7 (+https://www.terraform.io) terraform-provider-aws/4.62.0 (+https://registry.terraform.io/providers/hashicorp/aws) aws-sdk-go/1.44.237 (go1.19.7; linux; arm64)",
	}

	got, err := awscsmreceiver.ParseCSMMessage(input)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestListenAndServe(t *testing.T) {
	t.Parallel()
	addr := "127.0.0.1:31000"
	var wg sync.WaitGroup
	wg.Add(1)

	// Define the message handler function
	handler := func(got awscsmreceiver.CSMMessage) {
		defer wg.Done()
		want := awscsmreceiver.CSMMessage{
			Api:      "ListRoles",
			Service:  "IAM",
			Type:     "ApiCall",
			Region:   "eu-central-1",
			Attempts: 1,
			Latency:  817,
		}

		if !cmp.Equal(want, got) {
			t.Error(cmp.Diff(want, got))
		}
	}

	// Start the UDP server
	go func() {
		err := awscsmreceiver.ListenAndServe(addr, handler)
		if err != nil {
			t.Errorf("unable to start UDP server: %s", err)
		}
	}()

	// Wait a moment for the server to start listening
	time.Sleep(time.Millisecond * 100)

	// Send a test CSM message
	conn, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatalf("unable dial UDP server: %s", err)
	}
	defer conn.Close()

	testMsg := awscsmreceiver.CSMMessage{
		Api:      "ListRoles",
		Service:  "IAM",
		Type:     "ApiCall",
		Region:   "eu-central-1",
		Attempts: 1,
		Latency:  817,
	}

	payload, err := json.Marshal(testMsg)
	if err != nil {
		t.Fatalf("Error marshalling test message: %s", err)
	}

	_, err = conn.Write(payload)
	if err != nil {
		t.Fatalf("Error sending test message: %s", err)
	}

	wg.Wait()
}

func TestListenAndServeStop(t *testing.T) {
	t.Parallel()
	addr := "127.0.0.1:32000"

	// this channel enables to control the udp listener
	// when closed the udp listener is closed
	stopChan := make(chan struct{})

	// test have to wait until the handler function that is called inside a go
	// routine is able to process the message. Otherwise the call to t.Error is
	// a noop, which makes the ilussion that the test is passing.
	var wg sync.WaitGroup
	wg.Add(1)

	// an additional wait group is needed for the server, due the fact that the
	// server also runs in a go routine, the server go routine can receive the
	// message and call the handler but the handler may not be ready to process
	// the message, and then the wg.Done() call will never happens, making the
	// test to hang waiting for it.
	var serverReady sync.WaitGroup
	serverReady.Add(1)

	// Define the message handler function
	handler := func(got awscsmreceiver.CSMMessage) {
		defer wg.Done()
		want := awscsmreceiver.CSMMessage{
			Api:      "ListRoles",
			Service:  "IAM",
			Type:     "ApiCall",
			Region:   "eu-central-1",
			Attempts: 1,
			Latency:  817,
		}

		if !cmp.Equal(want, got) {
			t.Error(cmp.Diff(want, got))
		}

		close(stopChan)
	}

	// Start the UDP server
	go func() {
		serverReady.Done()
		err := awscsmreceiver.ListenAndServeStop(addr, handler, stopChan)
		if err != nil {
			t.Errorf("unable to start UDP server: %s", err)
		}
	}()

	serverReady.Wait()
	// Send a test CSM message
	conn, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatalf("unable dial UDP server: %s", err)
	}
	defer conn.Close()

	testMsg := awscsmreceiver.CSMMessage{
		Api:      "ListRoles",
		Service:  "IAM",
		Type:     "ApiCall",
		Region:   "eu-central-1",
		Attempts: 1,
		Latency:  817,
	}

	payload, err := json.Marshal(testMsg)
	if err != nil {
		t.Fatalf("Error marshalling test message: %s", err)
	}

	_, err = conn.Write(payload)
	if err != nil {
		t.Fatalf("Error sending test message: %s", err)
	}

	wg.Wait()
}

func TestWriteCSV(t *testing.T) {
	t.Parallel()
	testMsg := awscsmreceiver.CSMMessage{
		Type:                "TestType",
		Region:              "TestRegion",
		Service:             "TestService",
		Api:                 "TestApi",
		XAmznRequestId:      "TestRequestId",
		Attempts:            1,
		Latency:             100,
		Timestamp:           1234567890,
		Version:             2,
		HttpStatusCode:      200,
		FinalHttpStatusCode: 200,
		MaxRetriesExceeded:  0,
	}

	var got bytes.Buffer
	writerFunc := awscsmreceiver.WriteCSVHandler(&got)
	writerFunc(testMsg)

	want := bytes.NewBufferString("Type,Region,Service,Api,XAmznRequestId,Attempts,Latency,Timestamp,Version,HttpStatusCode,FinalHttpStatusCode,MaxRetriesExceeded\nTestType,TestRegion,TestService,TestApi,TestRequestId,1,100,1234567890,2,200,200,0\n")
	if !bytes.Equal(got.Bytes(), want.Bytes()) {
		t.Errorf("want:%v, got:%v", *want, got)
	}
}
