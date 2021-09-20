package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type gcs struct {
	bucket string
	prefix string
}

type live struct {
	liveDir string
}

func (live *live) load(path string) ([]byte, error) {
	path = filepath.Join(live.liveDir, path)

	return os.ReadFile(path)
}

func (gcs *gcs) test(ctx context.Context) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bkt := client.Bucket(gcs.bucket)
	it := bkt.Objects(ctx, &storage.Query{Prefix: ""})
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (gcs *gcs) load(ctx context.Context, path string) ([]byte, error) {
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

	defer reader.Close() // nolint:errcheck

	return io.ReadAll(reader)
}

func (gcs *gcs) save(ctx context.Context, path string, data []byte) error {
	path = filepath.Join(gcs.prefix, path)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil
	}

	bh := client.Bucket(gcs.bucket)

	obj := bh.Object(path)
	w := obj.NewWriter(ctx)
	defer w.Close() // nolint:errcheck
	r := bytes.NewBuffer(data)

	_, err = io.Copy(w, r)
	return err
}
