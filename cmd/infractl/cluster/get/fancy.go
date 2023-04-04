package get

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/spf13/cobra"

	"github.com/buger/jsonparser"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
)

type prettyCluster struct {
	v1.Cluster
}

func (p prettyCluster) PrettyPrint(cmd *cobra.Command) {
	var (
		createdOn, _   = ptypes.Timestamp(p.CreatedOn)
		lifespan, _    = ptypes.Duration(p.Lifespan)
		destroyedOn, _ = ptypes.Timestamp(p.DestroyedOn)
		remaining      = time.Until(createdOn.Add(lifespan))
	)

	cmd.Printf("ID:          %s\n", p.ID)
	cmd.Printf("Flavor:      %s\n", p.Flavor)
	cmd.Printf("Owner:       %s\n", p.Owner)
	if p.Description != "" {
		cmd.Printf("Description: %s\n", p.Description)
	}
	cmd.Printf("Status:      %s\n", p.Status)
	cmd.Printf("Created:     %v\n", common.FormatTime(createdOn))
	if p.URL != "" {
		cmd.Printf("URL:         %s\n", p.URL)
	}
	if destroyedOn.Unix() != 0 {
		cmd.Printf("Destroyed:   %v\n", common.FormatTime(destroyedOn))
	}
	cmd.Printf("Lifespan:    %s\n", common.FormatExpiration(remaining))
}

func (p prettyCluster) PrettyJSONPrint(cmd *cobra.Command) error {
	b := new(bytes.Buffer)
	m := jsonpb.Marshaler{EnumsAsInts: false, EmitDefaults: true, Indent: "  "}

	if err := m.Marshal(b, &p.Cluster); err != nil {
		return err
	}

	data := b.Bytes()
	// Because we're omitted defaults with the jsonpb Marshaler, and because we actually use a default (i.e., 0) enum
	// value for Status, if there is no Status in the marshaled json object, we need to add that default state. First,
	// check if there is a Status set, and if none is set, we should be able to safely assume that it should be set to
	// the default state
	// Another thing to note is that deleting keys is better than adding them (with this implementation, at least)
	// because adding will mess with the pretty formatting of the output
	checkDelete := []string{
		"Description", "Connect", "URL", "DestroyedOn.seconds", "DestroyedOn.nanos",
		"DestroyedOn", "CreatedOn.nanos", "Lifespan.nanos",
	}
	var toDelete []string
	for _, cd := range checkDelete {
		val, dataType, _, err := jsonparser.Get(data, strings.Split(cd, ".")...)
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				continue
			}
			return err
		}

		switch dataType {
		case jsonparser.String:
			sval := string(val)
			if sval == "" {
				toDelete = append(toDelete, cd)
			}
		case jsonparser.Number:
			fval, err := strconv.ParseFloat(string(val), 64)
			if err != nil {
				return err
			}
			if fval == 0 {
				toDelete = append(toDelete, cd)
			}
		case jsonparser.Null:
			toDelete = append(toDelete, cd)
		}
	}

	for _, e := range toDelete {
		data = jsonparser.Delete(data, strings.Split(e, ".")...)
	}

	cmd.Printf("%s\n", data)
	return nil
}
