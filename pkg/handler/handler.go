package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/open-policy-agent/gatekeeper-external-data-provider/pkg/utils"
	"k8s.io/klog/v2"
)

const (
	RegionNotScheduled = "region_not_scheduled"
	TimeNotScheduled   = "time_not_scheduled"
)

type SchedulingResponse struct {
	SchedulingTime     string `json:"schedulingTime"`
	SchedulingProvider string `json:"schedulingProvider"`
	SchedulingRegion   string `json:"schedulingRegion"`
}

type serverConfig struct {
	baseURL   string
	k8sSuffix string
	port      string
	endpoint  string
}

func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		utils.SendResponse(nil, "only POST is allowed", w)
		return
	}

	providerRequest, err := parseRequest(req)
	if err != nil {
		utils.SendResponse(nil, err.Error(), w)
		return
	}

	results, err := processKeys(providerRequest.Request.Keys)
	if err != nil {
		utils.SendResponse(nil, err.Error(), w)
		return
	}

	utils.SendResponse(&results, "", w)
}

func parseRequest(req *http.Request) (*externaldata.ProviderRequest, error) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %v", err)
	}

	klog.InfoS("received request", "body", requestBody)

	var providerRequest externaldata.ProviderRequest
	if err := json.Unmarshal(requestBody, &providerRequest); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %v", err)
	}

	klog.InfoS("keys", "keys", providerRequest.Request.Keys)
	return &providerRequest, nil
}

func processKeys(keys []string) ([]externaldata.Item, error) {
	schedulingResponse, err := getSchedulingResponse()
	if err != nil {
		return nil, fmt.Errorf("error getting scheduling response: %v", err)
	}

	results := make([]externaldata.Item, 0, len(keys))
	for _, key := range keys {
		item, err := processKey(key, schedulingResponse)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

func processKey(key string, resp SchedulingResponse) (externaldata.Item, error) {
	switch key {
	case RegionNotScheduled:
		klog.InfoS("assigned region", "region", resp.SchedulingRegion)
		return externaldata.Item{
			Key:   key,
			Value: resp.SchedulingRegion,
		}, nil

	case TimeNotScheduled:
		klog.InfoS("assigned time", "time", resp.SchedulingTime)
		return externaldata.Item{
			Key:   key,
			Value: resp.SchedulingTime,
		}, nil

	default:
		return externaldata.Item{}, fmt.Errorf("invalid key: %s", key)
	}
}

func getSchedulingResponse() (SchedulingResponse, error) {
	config := loadConfig()
	url := fmt.Sprintf("%s%s:%s%s", config.baseURL, config.k8sSuffix, config.port, config.endpoint)

	klog.InfoS("environment variables",
		"baseURL", config.baseURL,
		"k8sSuffix", config.k8sSuffix,
		"port", config.port,
		"endpoint", config.endpoint)

	resp, err := http.Get(url)
	if err != nil {
		return SchedulingResponse{}, fmt.Errorf("error making request to ai-inference-server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SchedulingResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SchedulingResponse{}, fmt.Errorf("error reading response body: %v", err)
	}

	var schedulingResponse SchedulingResponse
	if err := json.Unmarshal(body, &schedulingResponse); err != nil {
		return SchedulingResponse{}, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return schedulingResponse, nil
}

func loadConfig() serverConfig {
	return serverConfig{
		baseURL:   os.Getenv("AI_INFERENCE_SERVER_BASE_URL"),
		k8sSuffix: os.Getenv("AI_INFERENCE_SERVER_K8S_SUFFIX"),
		port:      os.Getenv("AI_INFERENCE_SERVER_PORT"),
		endpoint:  os.Getenv("AI_INFERENCE_SERVER_ENDPOINT"),
	}
}
