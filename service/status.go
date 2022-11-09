package service

import (
	"context"
	"log"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/kube"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyConfigurationv1 "k8s.io/client-go/applyconfigurations/core/v1"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type statusImpl struct {
	k8sConfigMapClient k8sv1.ConfigMapInterface
	infraNamespace     string
	infraStatusName    string
}

var (
	_ middleware.APIService       = (*statusImpl)(nil)
	_ v1.InfraStatusServiceServer = (*statusImpl)(nil)
)

const (
	infraNamespace  = "infra"
	infraStatusName = "status"
)

// NewStatusService creates a new InfraStatusService.
func NewStatusService() (middleware.APIService, error) {
	// TODO: can we remove hardcoding?
	k8sConfigMapClient, err := kube.GetK8sConfigMapClient(infraNamespace)
	if err != nil {
		return nil, err
	}
	return &statusImpl{
		k8sConfigMapClient: k8sConfigMapClient,
		infraNamespace:     infraNamespace,
		infraStatusName:    infraStatusName,
	}, nil
}

func (s *statusImpl) convertConfigMapToInfraStatus(configMap *corev1.ConfigMap) (*v1.InfraStatus, error) {
	// TODO: this is a bit clumsy. Catching the case that maintenanceActive is undefined
	maintainer := configMap.Data["maintainer"]
	maintenanceActiveValue := configMap.Data["maintenanceActive"]
	if maintenanceActiveValue == "" {
		maintenanceActiveValue = "false"
	}
	maintenanceActive, err := strconv.ParseBool(maintenanceActiveValue)
	if err != nil {
		return nil, err
	}

	return &v1.InfraStatus{
		Maintainer:        maintainer,
		MaintenanceActive: maintenanceActive,
	}, nil
}

func (s *statusImpl) convertInfraStatusToConfigMap(infraStatus *v1.InfraStatus) *applyConfigurationv1.ConfigMapApplyConfiguration {
	configMap := applyConfigurationv1.ConfigMap(s.infraStatusName, s.infraNamespace)
	return configMap.WithData(map[string]string{
		"maintainer":        infraStatus.GetMaintainer(),
		"maintenanceActive": strconv.FormatBool(infraStatus.GetMaintenanceActive()),
	})
}

// GetStatus shows infra maintenance status.
func (s *statusImpl) GetStatus(ctx context.Context, _ *empty.Empty) (*v1.InfraStatus, error) {

	configMap, err := s.k8sConfigMapClient.Get(ctx, s.infraStatusName, metav1.GetOptions{})
	if err != nil {
		// if err = doesn't exist, create empty, like in ResetStatus
		return nil, err
	}
	infraStatus, err := s.convertConfigMapToInfraStatus(configMap)
	if err != nil {
		return nil, err
	}
	return infraStatus, nil
}

func (s *statusImpl) SetStatus(ctx context.Context, infraStatus *v1.InfraStatus) (*v1.InfraStatus, error) {
	configMap := s.convertInfraStatusToConfigMap(infraStatus)

	_, err := s.k8sConfigMapClient.Apply(ctx, configMap, metav1.ApplyOptions{})
	if err != nil {
		return nil, err
	}
	log.Printf("New Status was set by maintainer %s\n", infraStatus.Maintainer)
	return infraStatus, nil
}

func (s *statusImpl) ResetStatus(ctx context.Context, _ *empty.Empty) (*v1.InfraStatus, error) {
	emptyInfraStatus := &v1.InfraStatus{}
	configMap := s.convertInfraStatusToConfigMap(emptyInfraStatus)
	_, err := s.k8sConfigMapClient.Apply(ctx, configMap, metav1.ApplyOptions{})
	if err != nil {
		return nil, err
	}
	log.Println("Status was reset")
	return emptyInfraStatus, nil
}

// Access configures access for this service.
func (s *statusImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.InfraStatusService/GetStatus": middleware.Anonymous,
		// TODO: change both modifying commands to middleware.Admin
		"/v1.InfraStatusService/ResetStatus": middleware.Authenticated,
		"/v1.InfraStatusService/SetStatus":   middleware.Authenticated,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *statusImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterInfraStatusServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *statusImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterInfraStatusServiceHandler(ctx, mux, conn)
}
