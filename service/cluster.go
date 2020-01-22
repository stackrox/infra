package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stackrox/infra/config"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type clusterImpl struct {
	clusterFlavors       map[string]*v1.Flavor
	defaultClusterFlavor *v1.Flavor
}

var (
	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService(clustersCfg []config.FlavorConfig) (middleware.APIService, error) {
	impl := clusterImpl{
		clusterFlavors:       make(map[string]*v1.Flavor, len(clustersCfg)),
		defaultClusterFlavor: nil,
	}

	for _, clusterCfg := range clustersCfg {
		// Sanity check and convert the configured availability.
		availability, found := v1.FlavorAvailability_value[clusterCfg.Availability]
		if !found {
			return nil, fmt.Errorf("unknown availability %q", clusterCfg.Availability)
		}

		cluster := &v1.Flavor{
			ID:           clusterCfg.ID,
			Name:         clusterCfg.Name,
			Description:  clusterCfg.Description,
			Availability: v1.FlavorAvailability(availability),
		}

		impl.clusterFlavors[cluster.ID] = cluster

		// Save off the default cluster separately.
		if cluster.Availability == v1.Flavor_default {
			// Ensure that more than one default flavor was not configured.
			if impl.defaultClusterFlavor != nil {
				return nil, fmt.Errorf("both %q and %q configured as default flavors", impl.defaultClusterFlavor.ID, cluster.ID)
			}
			impl.defaultClusterFlavor = cluster
		}
	}

	// Ensure a default flavor was configured.
	if impl.defaultClusterFlavor == nil {
		return nil, errors.New("no default flavor configured")
	}

	return &impl, nil
}

// List implements ClusterService.List.
func (s *clusterImpl) Flavors(context.Context, *empty.Empty) (*v1.FlavorsResponse, error) {
	resp := v1.FlavorsResponse{
		Flavors: make([]*v1.Flavor, 0, len(s.clusterFlavors)),
	}

	for _, cluster := range s.clusterFlavors {
		resp.Flavors = append(resp.Flavors, cluster)
	}

	if s.defaultClusterFlavor != nil {
		resp.Default = s.defaultClusterFlavor.ID
	}

	sort.Slice(resp.Flavors, func(i, j int) bool {
		if resp.Flavors[i].Availability != resp.Flavors[j].Availability {
			return resp.Flavors[i].Availability > resp.Flavors[j].Availability
		}
		return resp.Flavors[i].ID < resp.Flavors[j].ID
	})

	return &resp, nil
}

// AllowAnonymous declares that this service can be called anonymously.
func (s *clusterImpl) AllowAnonymous() bool {
	return false
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *clusterImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterClusterServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *clusterImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterClusterServiceHandler(ctx, mux, conn)
}
