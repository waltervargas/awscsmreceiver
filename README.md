[![Go Reference](https://pkg.go.dev/badge/github.com/waltervargas/awscsmreceiver.svg)](https://pkg.go.dev/github.com/waltervargas/awscsmreceiver)[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)[![Go Report Card](https://goreportcard.com/badge/github.com/waltervargas/awscsmreceiver)](https://goreportcard.com/report/github.com/waltervargas/awscsmreceiver)

# AWS Client-Side Monitoring (CSM) Receiver Package

`awscsmreceiver` is a Go package that provides functionality for parsing and
processing AWS Client Side Monitoring (CSM) messages received via UDP. 

## Features

- Parse JSON payload string into a CSMMessage struct.
- Listen and serve on a given address, processing incoming CSM messages using a provided MessageHandler.
- Write parsed CSM messages as CSV lines to an io.Writer.

# Installation

```sh
go get github.com/waltervargas/awscsmreceiver@v0.1.1
```

# Usage

First, import the awscsmreceiver package in your Go application:

```go
import "github.com/waltervargas/awscsmreceiver"
```

## Example: Listening for CSM Messages and Writing as CSV

```go
package main

import (
	"log"
	"os"

	"github.com/waltervargas/awscsmreceiver"
)

func main() {
	err := awscsmreceiver.ListenAndServe("127.0.0.1:31000", awscsmreceiver.WriteCSV(os.Stdout))
	if err != nil {
		log.Fatalf("unable to start UDP server: %s\n", err)
	}
}
```

# `awscsm2csv`

This package also provides a simple command-line application (`awscsm2csv`) to
listen for AWS Client-Side Monitoring (CSM) messages on a localhost UDP and
writes the received messages as CSV lines to the standard output.

## Install

```sh
go install github.com/waltervargas/awscsmreceiver/cmd/awscsm2csv@latest
```

## Usage

```sh
awscsm2csv | tee awscsm.csv
```

### Example usage with Terraform

In one terminal, run the program `awscsm2csv` and pipe the output to a file (`awscsm.csv`).
```sh
go run cmd/awscsm2csv/main.go | tee awscsm.csv
```

In another terminal run `terraform` with the environment variable `AWS_CSM_ENABLED` set to `true`: 
```sh
AWS_CSM_ENABLED=true terraform apply -auto-approve && AWS_CSM_ENABLED=true terraform destroy -auto-approve
```

```hcl
...
aws_sns_topic.example: Destruction complete after 1s

Destroy complete! Resources: 1 destroyed.
```

When terraform is done applying and destroying press <kbd>Ctrl</kbd> +
<kbd>C</kbd> on the terminal where `awscsm2csv` is running and look at the file `awscsm.csv`.
```sh
ApiCall,eu-west-1,SNS,DeleteTopic,72d0679e-c765-5a86-a8c5-3b2f313c1afc,1,272,1681388569581,0,0,200,0
^Csignal: interrupt
```
```sh
xsv table awscsm.csv | head -n 5 
Type            Region     Service  Api                  XAmznRequestId                        Attempts  Latency  Timestamp      Version  HttpStatusCode  FinalHttpStatusCode  MaxRetriesExceeded
ApiCallAttempt  eu-west-1  SNS      CreateTopic          7c30e94d-286e-5065-b5f8-b5fcde3bc555  0         0        1681388566036  1        200             0                    0
ApiCall         eu-west-1  SNS      CreateTopic          7c30e94d-286e-5065-b5f8-b5fcde3bc555  1         288      1681388566036  0        0               200                  0
ApiCallAttempt  eu-west-1  SNS      SetTopicAttributes   f1811115-4cf2-5bc8-9c0c-4da4baa41ef8  0         0        1681388566091  1        200             0                    0
ApiCall         eu-west-1  SNS      SetTopicAttributes   f1811115-4cf2-5bc8-9c0c-4da4baa41ef8  1         54       1681388566091  0        0               200                  0
```

# Use Cases

The `awscsmreceiver` package offers a flexible solution for developers and
operations teams looking to gain deeper insights into their AWS services through
Client Side Monitoring (CSM) messages. By utilizing this package, users can
build custom monitoring and alerting systems tailored to their specific
requirements. Some potential use cases include:

- Analyzing AWS SDK usage patterns in your applications to identify performance
  bottlenecks, such as high latency or excessive retries, which can then be
  addressed to improve application efficiency and user experience.

- Implementing real-time monitoring of API calls to detect anomalies or errors,
  allowing for proactive incident response and reduced downtime.

- Listing all AWS API calls made by a given terraform module enables you to for
  example write down a set of IAM policies required to deploy the module.
    
- Combining CSM data with other telemetry sources to create comprehensive
  monitoring dashboards, enabling better visibility into your AWS infrastructure
  and facilitating data-driven decision making.

- Generating detailed reports on AWS service usage and performance metrics to
  identify opportunities for cost optimization and resource allocation
  improvements.

- Enriching log data with CSM information to enhance troubleshooting
  capabilities and streamline root cause analysis.

# License

This package is licensed under the MIT License - see the LICENSE file for details.

# References

- https://docs.aws.amazon.com/sdk-for-go/api/aws/csm/
