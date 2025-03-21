package list

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyFlavorListResponse struct {
	v1.FlavorListResponse
}

func (p prettyFlavorListResponse) PrettyPrint(cmd *cobra.Command) {
	for _, flavor := range p.GetFlavors() {
		cmd.Printf("%s ", flavor.GetID())
		if flavor.GetID() == p.GetDefault() {
			cmd.Printf("(default)")
		}
		cmd.Println()
		cmd.Printf("  Name:         %s\n", flavor.GetName())
		cmd.Printf("  Description:  %s\n", flavor.GetDescription())
		cmd.Printf("  Availability: %s\n", flavor.GetAvailability())
	}
}

func (p prettyFlavorListResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	b := new(bytes.Buffer)
	m := jsonpb.Marshaler{EnumsAsInts: false, EmitDefaults: true, Indent: "  "}
	if err := m.Marshal(b, &p.FlavorListResponse); err != nil {
		return err
	}
	data := b.Bytes()

	cmd.Printf("%s\n", string(data))
	return nil
}
