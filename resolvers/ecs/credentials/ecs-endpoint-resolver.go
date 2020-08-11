package credentials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	rootEndpoint           = "http://169.254.170.2"
	maxRetries             = 4
	durationBetweenRetries = time.Second
)

type ECSResponse struct {
	RoleArn         string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

func rawJsonData(client *http.Client, endpoint string, respType string) ([]byte, error) {
	var resp []byte
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, err = getEndpointResponse(client, endpoint, respType)
		if err == nil {
			return resp, nil
		}
		fmt.Fprintf(os.Stderr, "Attempt [%d/%d]: unable to get ECS metadata response for '%s' from '%s': %v",
			i, maxRetries, respType, endpoint, err)
		time.Sleep(durationBetweenRetries)
	}

	return nil, err
}

func getEndpointResponse(client *http.Client, endpoint string, respType string) ([]byte, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to get response from ECS metadata service: %v", respType, err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: incorrect status code  %d", respType, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Task metadata: unable to read response body: %v", err)
	}

	return body, nil
}

func getECSMetadata(client *http.Client, credentialsuri string) (*ECSResponse, error) {
	body, err := rawJsonData(client, rootEndpoint+"/"+credentialsuri, "ECS metadata")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Received ECS metadata: %s \n", string(body))

	var ecsMetadata ECSResponse
	err = json.Unmarshal(body, &ecsMetadata)
	if err != nil {
		return nil, fmt.Errorf("ECS metadata: unable to parse response body: %v", err)
	}

	return &ecsMetadata, nil
}

func ProcessRequest() (fargate bool) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	time.Sleep(5 * time.Second)

	credentialsEndpoint := ""

	for _, pair := range os.Environ() {
		result := strings.Split(pair, "=")
		if result[1] == "AWS_ECS_FARGATE" {
			fargate = true
		}
		if result[0] == "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI" {
			credentialsEndpoint = result[1]
		}
	}

	_, err := getECSMetadata(client, credentialsEndpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get ECS metadata: %v", err)
		os.Exit(1)
	}
	return fargate
}
