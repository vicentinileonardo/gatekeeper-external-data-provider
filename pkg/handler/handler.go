package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	//"strings"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/open-policy-agent/gatekeeper-external-data-provider/pkg/utils"
	"k8s.io/klog/v2"
)

const (
	AI_INFERENCE_SERVER_BASE_URL   = "http://ai-inference-server"
	AI_INFERENCE_SERVER_K8S_SUFFIX = ".default.svc.cluster.local"
	AI_INFERENCE_SERVER_PORT       = "8080"
	AI_INFERENCE_SERVER_ENDPOINT   = "/scheduling"
)

type SchedulingResponse struct {
	SchedulingTime     string `json:"schedulingTime"`
	SchedulingProvider string `json:"schedulingProvider"`
	SchedulingRegion   string `json:"schedulingRegion"`
}

func Handler(w http.ResponseWriter, req *http.Request) {

	// TESTING, log the environment variables
	baseURL, k8sSuffix, port, endpoint := getConfig()
	klog.InfoS("environment variables", "baseURL", baseURL, "k8sSuffix", k8sSuffix, "port", port, "endpoint", endpoint)

	// only accept POST requests
	if req.Method != http.MethodPost {
		utils.SendResponse(nil, "only POST is allowed", w)
		return
	}

	// read request body
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		utils.SendResponse(nil, fmt.Sprintf("unable to read request body: %v", err), w)
		return
	}

	klog.InfoS("received request", "body", requestBody)

	// parse request body
	var providerRequest externaldata.ProviderRequest
	err = json.Unmarshal(requestBody, &providerRequest)
	if err != nil {
		utils.SendResponse(nil, fmt.Sprintf("unable to unmarshal request body: %v", err), w)
		return
	}

	results := make([]externaldata.Item, 0)
	// iterate over all keys
	for _, key := range providerRequest.Request.Keys {
		// Providers should add a caching mechanism to avoid extra calls to external data sources.

		// add checks to validate the key,
		//TODO: change to deal with multiple keys
		if key != "not_scheduled" {
			utils.SendResponse(nil, fmt.Sprintf("invalid key: %s", key), w)
			return
		} else if key == "not_scheduled" {

			schedulingRegion, err := getSchedulingRegion()
			if err != nil {
				utils.SendResponse(nil, fmt.Sprintf("error getting scheduling region: %v", err), w)
				return
			}

			klog.InfoS("scheduling region", "region", schedulingRegion)

			results = append(results, externaldata.Item{
				Key:   key,
				Value: schedulingRegion,
			})
		}
	}
	utils.SendResponse(&results, "", w)
}

func getSchedulingRegion() (string, error) {

	// Construct the URL
	url := fmt.Sprintf("%s%s:%s%s", AI_INFERENCE_SERVER_BASE_URL, AI_INFERENCE_SERVER_K8S_SUFFIX, AI_INFERENCE_SERVER_PORT, AI_INFERENCE_SERVER_ENDPOINT)
	resp, err := http.Get(url)

	if err != nil {
		return "", fmt.Errorf("error making request to ai-inference-server: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Unmarshal the response into a SchedulingResponse struct
	var schedulingResponse SchedulingResponse
	err = json.Unmarshal(body, &schedulingResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	// check if the response is empty
	if schedulingResponse.SchedulingRegion == "" {
		return "", fmt.Errorf("scheduling region is empty")
	}

	// Return the schedulingRegion
	return schedulingResponse.SchedulingRegion, nil
}

func getConfig() (string, string, string, string) {
	baseURL := os.Getenv("AI_INFERENCE_SERVER_BASE_URL")
	k8sSuffix := os.Getenv("AI_INFERENCE_SERVER_K8S_SUFFIX")
	port := os.Getenv("AI_INFERENCE_SERVER_PORT")
	endpoint := os.Getenv("AI_INFERENCE_SERVER_ENDPOINT")
	return baseURL, k8sSuffix, port, endpoint
}
