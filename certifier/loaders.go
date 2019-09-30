package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path/filepath"

	"cloud.google.com/go/storage"
)

type loader interface {
	Load(context.Context, string) ([]byte, error)
}

type saver interface {
	Save(context.Context, string, []byte) error
}

type GCS struct {
	bucket string
	prefix string
}

type Live struct {
	liveDir string
}

func (live *Live) Load(ctx context.Context, path string) ([]byte, error) {
	path = filepath.Join(live.liveDir, path)

	return ioutil.ReadFile(path)
}

func (gcs *GCS) Load(ctx context.Context, path string) ([]byte, error) {
	path = filepath.Join(gcs.prefix, path)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bh := client.Bucket(gcs.bucket)
	obj := bh.Object(path)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	return ioutil.ReadAll(reader)
}

func (gcs *GCS) Save(ctx context.Context, path string, data []byte) error {
	path = filepath.Join(gcs.prefix, path)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil
	}

	bh := client.Bucket(gcs.bucket)

	obj := bh.Object(path)
	w := obj.NewWriter(ctx)
	defer w.Close()
	r := bytes.NewBuffer(data)

	_, err = io.Copy(w, r)
	return err
}
