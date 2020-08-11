package ec2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	credentialsRole        = "http://169.254.169.254/latest/meta-data/iam/security-credentials"
	maxRetries             = 4
	durationBetweenRetries = time.Second
)

type EC2Response struct {
	Code            string
	LastUpdated     string
	Type            string
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
		fmt.Fprintf(os.Stderr, "Attempt [%d/%d]: unable to get EC2 metadata response for '%s' from '%s': %v",
			i, maxRetries, respType, endpoint, err)
		time.Sleep(durationBetweenRetries)
	}

	return nil, err
}

func getEndpointResponse(client *http.Client, endpoint string, respType string) ([]byte, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to get response: %v", respType, err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: incorrect status code  %d", respType, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("EC2 instance metadata: unable to read response body: %v", err)
	}

	return body, nil
}

func getEC2Metadata(client *http.Client, endpoint string) (*EC2Response, error) {
	body, err := rawJsonData(client, endpoint, "EC2 metadata")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Received EC2 metadata: %s \n", string(body))

	var ec2Metadata EC2Response
	err = json.Unmarshal(body, &ec2Metadata)
	if err != nil {
		return nil, fmt.Errorf("EC2 metadata: unable to parse response body: %v", err)
	}

	return &ec2Metadata, nil
}

func getEC2InstanceRole(client *http.Client, roleuri string) (EC2RoleResponse string, ret error) {
	body, err := rawJsonData(client, roleuri, "IAM Role")
	if err != nil {
		return string(body), err
	}

	ec2Role := string(body)

	return ec2Role, nil
}

func ProcessRequest() {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	time.Sleep(5 * time.Second)

	role, rolerr := getEC2InstanceRole(client, credentialsRole)
	if rolerr != nil {
		fmt.Fprintf(os.Stderr, "Unable to get IAM role: %v", rolerr)
		os.Exit(1)
	}

	_, metaerr := getEC2Metadata(client, credentialsRole+"/"+role)
	if metaerr != nil {
		fmt.Fprintf(os.Stderr, "Unable to get IAM role credentials: %v", metaerr)
		os.Exit(1)
	}

}
