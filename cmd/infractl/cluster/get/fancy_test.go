package get_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	durationpb "github.com/golang/protobuf/ptypes/duration"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/get"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FakeClusterServiceClient struct {
	infoFn      func(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error)
	listFn      func(ctx context.Context, req *v1.ClusterListRequest) (*v1.ClusterListResponse, error)
	lifespanFn  func(ctx context.Context, req *v1.LifespanRequest) (*durationpb.Duration, error)
	createFn    func(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error)
	artifactsFn func(ctx context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error)
	deleteFn    func(ctx context.Context, clusterID *v1.ResourceByID) (*emptypb.Empty, error)
	logsFn      func(ctx context.Context, clusterID *v1.ResourceByID) (*v1.LogsResponse, error)
}

var _ v1.ClusterServiceServer = (*FakeClusterServiceClient)(nil)

func (csc *FakeClusterServiceClient) Info(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	if csc.infoFn != nil {
		return csc.infoFn(ctx, clusterID)
	}

	return nil, errors.New("this method was not set up with a response - must set infoFn")
}

func (csc *FakeClusterServiceClient) List(ctx context.Context, req *v1.ClusterListRequest) (*v1.ClusterListResponse, error) {
	if csc.listFn != nil {
		return csc.listFn(ctx, req)
	}

	return nil, errors.New("this method was not set up with a response - must set listFn")
}

func (csc *FakeClusterServiceClient) Lifespan(ctx context.Context, req *v1.LifespanRequest) (*durationpb.Duration, error) {
	if csc.lifespanFn != nil {
		return csc.lifespanFn(ctx, req)
	}

	return nil, errors.New("this method was not set up with a response - must set lifespanFn")
}

func (csc *FakeClusterServiceClient) Create(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error) {
	if csc.createFn != nil {
		return csc.createFn(ctx, req)
	}
	return nil, errors.New("this method was not set up with a response - must set createFn")
}

func (csc *FakeClusterServiceClient) Artifacts(ctx context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error) {
	if csc.artifactsFn != nil {
		return csc.artifactsFn(ctx, clusterID)
	}
	return nil, errors.New("this method was not set up with a response - must set artifactsFn")
}

func (csc *FakeClusterServiceClient) Delete(ctx context.Context, clusterID *v1.ResourceByID) (*emptypb.Empty, error) {
	if csc.deleteFn != nil {
		return csc.deleteFn(ctx, clusterID)
	}
	return nil, errors.New("this method was not set up with a response - must set deleteFn")
}

func (csc *FakeClusterServiceClient) Logs(ctx context.Context, clusterID *v1.ResourceByID) (*v1.LogsResponse, error) {
	if csc.logsFn != nil {
		return csc.logsFn(ctx, clusterID)
	}
	return nil, errors.New("this method was not set up with a response - must set logsFn")
}

// generateTLSCertification generates a fake TLS certificate for testing purposes
func generateTLSCertification() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Test Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Store cert PEM block
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return tls.Certificate{}, err
	}

	// Get and store key PEM block. Note this is an ECDSA cert
	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}); err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
}

func TestGetClusterJSONOutput(t *testing.T) {
	testTime, err := ptypes.TimestampProto(time.Date(2022, time.April, 1, 1, 0, 0, 0, time.UTC))
	assert.NoError(t, err)
	csc := &FakeClusterServiceClient{
		infoFn: func(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
			return &v1.Cluster{
				ID:          "test-123",
				Status:      v1.Status_FAILED,
				Flavor:      v1.Flavor_stable.String(),
				Owner:       "me@redhat.com",
				CreatedOn:   testTime,
				DestroyedOn: nil,
				Lifespan:    ptypes.DurationProto(10800 * time.Second),
				Description: "My test cluster",
			}, nil
		},
	}

	// Note that port 0 is a dynamic port - not actually port 0
	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)

	cert, err := generateTLSCertification()
	// This failure is most likely a problem with the testing code, not the code we're trying to test
	assert.NoError(t, err)

	gsrv := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	v1.RegisterClusterServiceServer(gsrv, csc)
	go func() {
		err := gsrv.Serve(l)
		assert.NoError(t, err)
	}()

	cmd := &cobra.Command{
		SilenceUsage: true,
		Use:          os.Args[0],
		Version:      buildinfo.Version(),
	}

	common.AddCommonFlags(cmd)
	cmd.AddCommand(get.Command())
	cmd.SetArgs([]string{
		"get",
		"test-123",
		fmt.Sprintf("--endpoint=%s", l.Addr()),
		"--insecure",
		"--json",
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	assert.NoError(t, err)

	expected := `{
  "ID": "test-123",
  "Status": "FAILED",
  "Flavor": "stable",
  "Owner": "me@redhat.com",
  "CreatedOn": {
    "seconds": "1648774800"
  },
  "Lifespan": {
    "seconds": "10800"
  },
  "Description": "My test cluster"
}
`
	assert.Equal(t, expected, buf.String())
}
