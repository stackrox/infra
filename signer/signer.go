// Package signer facilitates the generation of signed GCS URLS.
package signer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	cfg    jwt.Config
	client *storage.Client
}

// NewFromEnv constructs a Signer from the GOOGLE_APPLICATION_CREDENTIALS set
// in the working environment.
func NewFromEnv() (*Signer, error) {
	credentialsFilename, found := os.LookupEnv(googleCredentialsEnvVar)
	if !found {
		return nil, fmt.Errorf("environment variable %q was not set", googleCredentialsEnvVar)
	}

	data, err := os.ReadFile(credentialsFilename)
	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	jwtCfg, err := google.JWTConfigFromJSON(data)
	if err != nil {
		return nil, err
	}

	return &Signer{
		cfg:    *jwtCfg,
		client: client,
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

// Contents returns the raw contents of the named GCS object. It is expected
// that these are argo workflow artifacts either single files tar gzip'd or
// plain files.
func (s Signer) Contents(gcsBucketName, gcsBucketKey string) ([]byte, error) {
	br, err := s.client.Bucket(gcsBucketName).Object(gcsBucketKey).NewReader(context.Background())
	if err != nil {
		return nil, err
	}

	attrs, err := s.client.Bucket(gcsBucketName).Object(gcsBucketKey).Attrs(context.Background())
	if err != nil {
		return nil, err
	}

	if !strings.Contains(attrs.ContentType, "gzip") {
		return io.ReadAll(br)
	}

	gr, err := gzip.NewReader(br)
	if err != nil {
		return nil, err
	}
	defer gr.Close() // nolint:errcheck

	// Archive is a normal tar archive.
	tr := tar.NewReader(gr)

	// We're expecting 1 and only 1 file in the archive, so read just the
	// first entry.
	if _, err := tr.Next(); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("unexpected EOF reading artifact")
		}
		return nil, err
	}

	return io.ReadAll(tr)
}
