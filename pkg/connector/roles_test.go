package connector

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

func TestRoleEntitlementsReadPrivilegesFromProfile(t *testing.T) {
	role := &types.AuthorizationRole{
		RoleId:    -1,
		Name:      "Admin",
		Privilege: []string{"System.View", "System.Read"},
		Info:      &types.Description{Label: "Admin", Summary: "Full access"},
	}

	res, err := roleResource(role)
	if err != nil {
		t.Fatalf("roleResource: %v", err)
	}

	r := newRoleBuilder(nil)
	ents, _, _, err := r.Entitlements(context.Background(), res, nil)
	if err != nil {
		t.Fatalf("Entitlements: %v", err)
	}

	want := []string{
		"role:-1:member",
		"role:-1:System.View",
		"role:-1:System.Read",
	}
	if len(ents) != len(want) {
		t.Fatalf("got %d entitlements, want %d", len(ents), len(want))
	}
	for i, e := range ents {
		if e.GetId() != want[i] {
			t.Errorf("entitlement %d: got %q, want %q", i, e.GetId(), want[i])
		}
	}
}
