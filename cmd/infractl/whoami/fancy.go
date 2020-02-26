package whoami

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type whoamiResp v1.WhoamiResponse

func (r whoamiResp) PrettyPrint() {
	switch p := r.Principal.(type) {
	case *v1.WhoamiResponse_User:
		panic("authenticating as a user is not possible in this context")
	case *v1.WhoamiResponse_ServiceAccount:
		fmt.Println("Service Account")
		fmt.Printf("  Name:        %s\n", p.ServiceAccount.GetName())
		fmt.Printf("  Description: %s\n", p.ServiceAccount.GetDescription())
		fmt.Printf("  Email:       %s\n", p.ServiceAccount.GetEmail())
	case nil:
		fmt.Println("Anonymous")
	}
}
