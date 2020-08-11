package main

import (
	"fmt"
	ec2cred "meta-role-checker/resolvers/ec2/credentials"
	ecscred "meta-role-checker/resolvers/ecs/credentials"
	"net/http"
	"os"
	"time"
)

const (
	ecsMetadataEndpoint = "http://169.254.170.2"
	ec2MetadataEndpoint = "http://169.254.169.254"
)

func getEndpointAvailability(client *http.Client, endpoint string) error {
	_, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("%s: Unable to establish a connection: %v", endpoint, err)
	}
	return nil
}

func main() {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	time.Sleep(5 * time.Second)

	errecs := getEndpointAvailability(client, ecsMetadataEndpoint)
	if errecs != nil {
		errec2 := getEndpointAvailability(client, ec2MetadataEndpoint)
		if errec2 != nil {
			fmt.Printf("Error ECS: %v\n", errecs)
			fmt.Printf("Error EC2: %v\n", errec2)
			os.Exit(1)
		} else {
			ec2cred.ProcessRequest() //EKS only
		}
	} else {
		//task.ProcessRequest()
		if fargate := ecscred.ProcessRequest(); !fargate {
			ec2cred.ProcessRequest() //All ECS EC2 task types, no Fargate
		}
	}
}
