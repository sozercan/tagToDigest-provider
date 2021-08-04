// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package keychain creates credentials for authenticating to container image registries.
package keychain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const keyReference = "k8s://"

var createClientFn = createClient // override for testing

// Create a multi keychain based in input arguments
func Create(ctx context.Context, log logr.Logger, config *rest.Config, k8sRef string) (authn.Keychain, error) {
	if config == nil {
		log.Info("creating offline keychain")
		return authn.NewMultiKeychain(google.Keychain, authn.DefaultKeychain), nil
	}
	client, err := createClientFn(config)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes Clientset: %w", err)
	}
	kkc, err := createK8schain(ctx, log, client, k8sRef)
	if err != nil {
		return nil, fmt.Errorf("could not create k8schain: %w", err)
	}
	log.Info("creating k8s keychain")
	return authn.NewMultiKeychain(kkc, authn.DefaultKeychain), nil
}

func createClient(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

func createK8schain(ctx context.Context, log logr.Logger, client kubernetes.Interface, k8sRef string) (authn.Keychain, error) {
	namespace, secretName, err := parseRef(k8sRef)
	if err != nil {
		return nil, err
	}

	log.Info("creating k8schain", "namespace", namespace, "imagePullSecrets", secretName)
	return k8schain.New(ctx, client, k8schain.Options{
		Namespace:          namespace,
		ServiceAccountName: "provider-tagtodigest-sa",
		ImagePullSecrets:   []string{secretName},
	})
}

// the reference should be formatted as <namespace>/<secret name>
func parseRef(k8sRef string) (string, string, error) {
	s := strings.Split(strings.TrimPrefix(k8sRef, keyReference), "/")
	if len(s) != 2 {
		return "", "", errors.New("kubernetes specification should be in the format k8s://<namespace>/<secret>")
	}
	return s[0], s[1], nil
}
