package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/goccy/go-json"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/sozercan/tagToDigest-provider/pkg/keychain"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

var log logr.Logger

const (
	timeout    = 1 * time.Second
	apiVersion = "externaldata.gatekeeper.sh/v1alpha1"
	kind       = "ProviderResponse"
)

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("unable to initialize logger: %v", err))
	}
	log = zapr.NewLogger(zapLog)
	log.WithName("tagToDigest-provider")

	log.Info("starting server...")
	http.HandleFunc("/mutate", mutate)

	if err = http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}

func mutate(w http.ResponseWriter, req *http.Request) {
	// only accept POST requests
	if req.Method != http.MethodPost {
		sendResponse(nil, "only POST is allowed", w)
		return
	}

	// read request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sendResponse(nil, fmt.Sprintf("unable to read request body: %v", err), w)
		return
	}

	// parse request body
	var providerRequest externaldata.ProviderRequest
	err = json.Unmarshal(requestBody, &providerRequest)
	if err != nil {
		sendResponse(nil, fmt.Sprintf("unable to unmarshal request body: %v", err), w)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "unable to get in-cluster config")
		return
	}

	secretKeyRef := os.Getenv("SECRET_NAME")
	var kc authn.Keychain
	if secretKeyRef != "" {
		kc, err = keychain.Create(ctx, log, config, secretKeyRef)
		if err != nil {
			log.Error(err, "unable to create keychain")
			return
		}
	}

	results := make([]externaldata.Item, 0)
	for _, image := range providerRequest.Request.Keys {
		item := externaldata.Item{
			Key:   image,
			Value: image,
		}
		if !strings.Contains(image, "@sha256") {
			digest, err := crane.Digest(image, crane.WithAuthFromKeychain(kc))
			if err != nil {
				log.Error(err, "unable to get digest")
				item.Error = err.Error()
			}
			item.Value = fmt.Sprintf("%s@%s", image, digest)
		}
		results = append(results, item)
	}

	sendResponse(&results, "", w)
}

// sendResponse sends back the response to Gatekeeper.
func sendResponse(results *[]externaldata.Item, systemErr string, w http.ResponseWriter) {
	response := externaldata.ProviderResponse{
		APIVersion: apiVersion,
		Kind:       kind,
		Response: externaldata.Response{
			Idempotent: true,
		},
	}

	if results != nil {
		response.Response.Items = *results
	} else {
		response.Response.SystemError = systemErr
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}

	log.Info("mutate", "response", response)
}
