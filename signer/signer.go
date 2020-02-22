// Package signer facilitates the generation of signed GCS URLS.
package signer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

const (
	// googleCredentialsEnvVar is the name of the environment variable, that
	// itself contains the name of a GCP IAM credentials JSON file.
	googleCredentialsEnvVar = "GOOGLE_APPLICATION_CREDENTIALS"

	// gcsSignedURLLifespan is the length of time a generated GCS signed URL
	// will be valid for.
	gcsSignedURLLifespan = 10 * time.Minute
)

// Signer facilitates the generation of signed GCS URLS.
type Signer struct {
	cfg jwt.Config
}

// NewFromEnv constructs a Signer from the GOOGLE_APPLICATION_CREDENTIALS set
// in the working environment.
func NewFromEnv() (*Signer, error) {
	credentialsFilename, found := os.LookupEnv(googleCredentialsEnvVar)
	if !found {
		return nil, fmt.Errorf("environment variable %q was not set", googleCredentialsEnvVar)
	}

	data, err := ioutil.ReadFile(credentialsFilename)
	if err != nil {
		return nil, err
	}

	jwtCfg, err := google.JWTConfigFromJSON(data)
	if err != nil {
		return nil, err
	}

	return &Signer{
		cfg: *jwtCfg,
	}, nil
}

// Generate generates a url that can be used to download the given GCS object
// for some amount of time.
//
// This is accomplished by creating a GCS signed URL. For more information see:
// https://cloud.google.com/storage/docs/access-control/signed-urls
func (s Signer) Generate(gcsBucketName, gcsBucketKey string) (string, error) {
	return storage.SignedURL(gcsBucketName, gcsBucketKey, &storage.SignedURLOptions{
		GoogleAccessID: s.cfg.Email,
		PrivateKey:     s.cfg.PrivateKey,
		Method:         http.MethodGet,
		Expires:        time.Now().Add(gcsSignedURLLifespan),
	})
}
