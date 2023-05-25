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
	errorsv1 "k8s.io/apimachinery/pkg/api/errors"
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
	infraStatus := v1.InfraStatus{
		Maintainer: configMap.Data["maintainer"],
	}

	maintenanceActiveValue, ok := configMap.Data["maintenanceActive"]
	if !ok || maintenanceActiveValue == "" {
		infraStatus.MaintenanceActive = false
	} else {
		maintenanceActive, err := strconv.ParseBool(maintenanceActiveValue)
		if err != nil {
			return nil, err
		}
		infraStatus.MaintenanceActive = maintenanceActive
	}

	return &infraStatus, nil
}

func (s *statusImpl) convertInfraStatusToConfigMap(infraStatus *v1.InfraStatus) *applyConfigurationv1.ConfigMapApplyConfiguration {
	configMap := applyConfigurationv1.ConfigMap(s.infraStatusName, s.infraNamespace)
	return configMap.WithData(map[string]string{
		"maintainer":        infraStatus.GetMaintainer(),
		"maintenanceActive": strconv.FormatBool(infraStatus.GetMaintenanceActive()),
	})
}

func (s *statusImpl) createEmptyInfraStatus(ctx context.Context) (*v1.InfraStatus, error) {
	emptyInfraStatus := &v1.InfraStatus{}
	configMap := s.convertInfraStatusToConfigMap(emptyInfraStatus)
	_, err := s.k8sConfigMapClient.Apply(ctx, configMap, metav1.ApplyOptions{FieldManager: "infra"})
	if err != nil {
		return nil, err
	}
	return emptyInfraStatus, nil
}

// GetStatus shows infra maintenance status.
func (s *statusImpl) GetStatus(ctx context.Context, _ *empty.Empty) (*v1.InfraStatus, error) {
	configMap, err := s.k8sConfigMapClient.Get(ctx, s.infraStatusName, metav1.GetOptions{})
	if err != nil {
		if errorsv1.IsNotFound(err) {
			infraStatus, err := s.createEmptyInfraStatus(ctx)
			if err != nil {
				return nil, err
			}
			log.Infow("initialized infra status lazily")
			return infraStatus, nil
		}
		return nil, err
	}
	infraStatus, err := s.convertConfigMapToInfraStatus(configMap)
	if err != nil {
		return nil, err
	}
	return infraStatus, nil
}

// SetStatus activates maintenance and sets the maintainer to the user from the context
func (s *statusImpl) SetStatus(ctx context.Context, infraStatus *v1.InfraStatus) (*v1.InfraStatus, error) {
	configMap := s.convertInfraStatusToConfigMap(infraStatus)

	_, err := s.k8sConfigMapClient.Apply(ctx, configMap, metav1.ApplyOptions{FieldManager: "infra"})
	if err != nil {
		return nil, err
	}
	log.Infow("new status set",
		"maintainer", infraStatus.GetMaintainer(),
		"maintenance-active", infraStatus.GetMaintenanceActive(),
	)
	return infraStatus, nil
}

// ResetStatus sets the maintenance active to false and clears the maintainer.
func (s *statusImpl) ResetStatus(ctx context.Context, _ *empty.Empty) (*v1.InfraStatus, error) {
	infraStatus, err := s.createEmptyInfraStatus(ctx)
	if err != nil {
		return nil, err
	}
	log.Infow("status was reset")
	return infraStatus, nil
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
