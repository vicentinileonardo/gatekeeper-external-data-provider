package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	//"strings"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/open-policy-agent/gatekeeper-external-data-provider/pkg/utils"
	"k8s.io/klog/v2"
)

func Handler(w http.ResponseWriter, req *http.Request) {
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
		//key must be schedulingRegion
		if key != "not_scheduled" {
			utils.SendResponse(nil, fmt.Sprintf("invalid key: %s", key), w)
			return
		} else if key == "not_scheduled" {
			results = append(results, externaldata.Item{
				Key:   key,
				Value: "us-central1",
			})
		}

		//TODO: make request to external data source

	}
	utils.SendResponse(&results, "", w)
}
