package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/go-containerregistry/pkg/crane"
	"go.uber.org/zap"
)

var log logr.Logger

type ProviderCacheKey struct {
	ProviderName string `json:"providerName,omitempty"`
	OutboundData string `json:"outboundData,omitempty"`
}

func (k ProviderCacheKey) MarshalText() ([]byte, error) {
	type p ProviderCacheKey
	return json.Marshal(p(k))
}

func (k *ProviderCacheKey) UnmarshalText(text []byte) error {
	type x ProviderCacheKey
	return json.Unmarshal(text, (*x)(k))
}

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
	w.Header().Set("Content-Type", "application/json")

	var input map[ProviderCacheKey]string

	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error(err, "unable to read request body")
		return
	}

	for i := range input {
		if !strings.Contains(i.OutboundData, "sha256") {
			digest, err := crane.Digest(i.OutboundData)
			if err != nil {
				log.Error(err, "unable to get digest")
				return
			}
			input[i] = i.OutboundData + "@" + digest
		} else {
			input[i] = i.OutboundData
		}
	}

	out, err := json.Marshal(input)
	if err != nil {
		log.Error(err, "unable to marshal to output")
		return
	}

	log.Info("mutate", "response", out)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(out))
}
