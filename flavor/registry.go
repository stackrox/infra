// Package flavor enables the configuration and registration of flavors and
// their associated Argo workflows.
package flavor

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// pair represents a tuple of an Argo workflow and a flavor.
type pair struct {
	workflow v1alpha1.Workflow
	flavor   v1.Flavor
}

// Registry represents the set of all configured flavors.
type Registry struct {
	flavors       map[string]pair
	defaultFlavor string
}

// Flavors returns a sorted list of all registered flavors.
func (r *Registry) Flavors() []v1.Flavor {
	results := make([]v1.Flavor, 0, len(r.flavors))
	for _, pair := range r.flavors {
		results = append(results, pair.flavor)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Availability != results[j].Availability {
			return results[i].Availability > results[j].Availability
		}
		return results[i].ID < results[j].ID
	})

	return results
}

// add registers the given flavor and workflow.
func (r *Registry) add(flavor v1.Flavor, workflow v1alpha1.Workflow) error {
	// Validate that another flavor with the same ID was not already added.
	if _, found := r.flavors[flavor.ID]; found {
		return fmt.Errorf("duplicate flavor id %q", flavor.ID)
	}

	// Validate that the flavor parameters and workflow parameters are
	// perfectly equivalent.
	if err := CheckWorkflowEquivalence(flavor, workflow); err != nil {
		return err
	}

	// Register this flavor.
	r.flavors[flavor.ID] = pair{
		workflow: workflow,
		flavor:   flavor,
	}
	log.Printf("registered flavor %q (%s)\n", flavor.ID, flavor.Name)

	// Register a default flavor if one has not already been registered.
	if flavor.Availability == v1.Flavor_default {
		// There is more than 1 default flavor!
		if r.defaultFlavor != "" {
			return fmt.Errorf("both %q and %q configured as default flavors", r.defaultFlavor, flavor.ID)
		}
		r.defaultFlavor = flavor.ID
		log.Printf("registered default flavor %q (%s)\n", flavor.ID, flavor.Name)
	}

	return nil
}

// Default returns the default flavor
func (r *Registry) Default() string {
	return r.defaultFlavor
}

// Get returns the named flavor, and if it exists.
func (r *Registry) Get(id string) (v1.Flavor, v1alpha1.Workflow, bool) {
	if pair, found := r.flavors[id]; found {
		return pair.flavor, pair.workflow, true
	}

	return v1.Flavor{}, v1alpha1.Workflow{}, false
}

// check validates that a default flavor was added.
func (r *Registry) check() (*Registry, error) {
	if r.defaultFlavor == "" {
		return nil, errors.New("no default flavor configured")
	}
	return r, nil
}

// NewFromConfig parses the given flavor config file, along with all referenced
// Argo workflows, and returns a registry containing all flavors.
func NewFromConfig(filename string) (*Registry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read flavorCfg config file %q", filename)
	}

	var flavorsCfg []config.FlavorConfig
	if err := yaml.Unmarshal(data, &flavorsCfg); err != nil {
		return nil, err
	}

	registry := &Registry{
		flavors: make(map[string]pair),
	}

	for _, flavorCfg := range flavorsCfg {
		// Sanity check and convert the configured availability.
		availability, found := v1.FlavorAvailability_value[flavorCfg.Availability]
		if !found {
			return nil, fmt.Errorf("unknown availability %q", flavorCfg.Availability)
		}

		parameters := make(map[string]*v1.Parameter, len(flavorCfg.Parameters))
		for order, parameter := range flavorCfg.Parameters {
			param := &v1.Parameter{
				Name:        parameter.Name,
				Description: parameter.Description,
				Value:       parameter.Value,
				Help:        parameter.Help,
				FromFile:    parameter.FromFile,
				Order:       int32(order) + 1,
			}

			switch parameter.Kind {
			case config.ParameterHardcoded:
				param.Internal = true
				param.Optional = true
			case config.ParameterRequired:
				param.Internal = false
				param.Optional = false
			case config.ParameterOptional:
				param.Internal = false
				param.Optional = true
			}

			parameters[parameter.Name] = param
		}

		artifacts := make(map[string]*v1.FlavorArtifact, len(flavorCfg.Artifacts))
		for _, artifact := range flavorCfg.Artifacts {
			// Pack the list of tags into a set of tags.
			tags := make(map[string]*empty.Empty, len(artifact.Tags))
			for _, tag := range artifact.Tags {
				tags[tag] = &empty.Empty{}
			}

			artifacts[artifact.Name] = &v1.FlavorArtifact{
				Name:        artifact.Name,
				Description: artifact.Description,
				Tags:        tags,
			}
		}

		flavor := v1.Flavor{
			ID:           flavorCfg.ID,
			Name:         flavorCfg.Name,
			Description:  flavorCfg.Description,
			Availability: v1.FlavorAvailability(availability),
			Parameters:   parameters,
			Artifacts:    artifacts,
		}

		// Parse the references Argo workflow file.
		data, err := os.ReadFile(flavorCfg.WorkflowFile)
		if err != nil {
			return nil, err
		}

		var workflow v1alpha1.Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return nil, err
		}

		// Register the flavor and workflow pair.
		if err := registry.add(flavor, workflow); err != nil {
			return nil, err
		}
	}

	return registry.check()
}

// CheckWorkflowEquivalence verifies that the given flavor parameters and
// workflow parameters are equivalent sets.
//
// - All parameter names must be unique.
//
// - All parameters from one set must be in the other.
func CheckWorkflowEquivalence(flavor v1.Flavor, workflow v1alpha1.Workflow) error {
	// Workflow have a list of parameters, so convert to a set.
	workflowParamSet := make(map[string]struct{})
	for _, param := range workflow.Spec.Arguments.Parameters {
		if _, found := workflowParamSet[param.Name]; found {
			return fmt.Errorf("flavor %q workflow had duplicate parameter %q", flavor.ID, param.Name)
		}
		workflowParamSet[param.Name] = struct{}{}
	}

	// Verify that every workflow parameter has a matching flavor parameter.
	for workflowParamName := range workflowParamSet {
		if _, found := flavor.Parameters[workflowParamName]; !found {
			return fmt.Errorf("flavor %q workflow had parameter %q but manifest did not", flavor.ID, workflowParamName)
		}
	}

	// Verify that every flavor parameter has a matching workflow parameter.
	for flavorParamName, param := range flavor.Parameters {
		// This parameter has a hardcoded value, and therefore doesn't need to
		// be specified by the user.
		if param.Value != "" {
			continue
		}
		if _, found := workflowParamSet[flavorParamName]; !found {
			return fmt.Errorf("flavor %q manifest had parameter %q but workflow did not", flavor.ID, flavorParamName)
		}
	}

	// Sets are equivalent!
	return nil
}
