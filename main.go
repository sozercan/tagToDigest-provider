package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/goccy/go-json"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/sozercan/tagToDigest-provider/pkg/keychain"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

var log logr.Logger

const (
	timeout      = 3 * time.Second
	providerName = "tagtodigest-provider"
)

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

	for i := range input {
		if i.ProviderName == providerName && !strings.Contains(i.OutboundData, "sha256") {
			digest, err := crane.Digest(i.OutboundData, crane.WithAuthFromKeychain(kc))
			if err != nil {
				log.Error(err, "unable to get digest")
				input[i] = i.OutboundData
			} else {
				input[i] = i.OutboundData + "@" + digest
			}
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
