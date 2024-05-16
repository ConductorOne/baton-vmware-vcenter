package connector

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const RoleMembershipEntitlement = "member"

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *govmomi.Client
}

func roleResource(role *types.AuthorizationRole) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_id":        role.RoleId,
		"role_name":      role.Name,
		"system_defined": role.System,
		"privileges":     strings.Join(role.Privilege, ","),
	}

	roleID := strconv.FormatInt(int64(role.RoleId), 10)
	res, err := rs.NewRoleResource(
		role.Name,
		roleResourceType,
		roleID,
		[]rs.RoleTraitOption{
			rs.WithRoleProfile(profile),
		},
		rs.WithDescription(role.Info.GetDescription().Summary),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	am := object.NewAuthorizationManager(r.client.Client)
	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to list roles: %w", err)
	}

	var rv []*v2.Resource
	for _, role := range roles {
		rr, err := roleResource(&role)
		if err != nil {
			return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to create role resource: %w", err)
		}

		rv = append(rv, rr)
	}

	return rv, "", nil, nil
}

func (r *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	// add assignment entitlement for representing a role assignment to a principal
	aOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType, groupResourceType),
		ent.WithDisplayName(fmt.Sprintf("Role %s %s", resource.DisplayName, RoleMembershipEntitlement)),
		ent.WithDescription(fmt.Sprintf("Role %s membership in vCenter Server", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, RoleMembershipEntitlement, aOptions...))

	// add permission entitlements for all privileges in the role
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to get role trait: %w", err)
	}

	// parse privileges from role trait
	privilegesPayload, ok := rs.GetProfileStringValue(roleTrait.Profile, "privileges")
	if !ok {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to get role trait privileges")
	}

	privileges := strings.Split(privilegesPayload, ",")
	for _, p := range privileges {
		pOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(userResourceType, groupResourceType),
			ent.WithDisplayName(fmt.Sprintf("Role %s %s", resource.DisplayName, p)),
			ent.WithDescription(fmt.Sprintf("Role %s privilege %s in vCenter Server", resource.DisplayName, p)),
		}

		rv = append(rv, ent.NewPermissionEntitlement(resource, p, pOptions...))
	}

	return rv, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	am := object.NewAuthorizationManager(r.client.Client)

	// parse role id to required int32
	roleID, err := strconv.ParseInt(resource.Id.Resource, 10, 32)
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to parse role id: %w", err)
	}

	// permissions represent how role is assigned to a principal
	permissions, err := am.RetrieveRolePermissions(ctx, int32(roleID))
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to list permissions for role %s: %w", resource.DisplayName, err)
	}

	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to get role trait: %w", err)
	}

	// parse privileges from role trait to make grant expandable
	privilegesPayload, ok := rs.GetProfileStringValue(roleTrait.Profile, "privileges")
	if !ok {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to get role trait privileges")
	}
	privileges := strings.Split(privilegesPayload, ",")

	// prepare format of entitlements to be expanded
	var entitlementIDs []string
	for _, priv := range privileges {
		entitlementIDs = append(entitlementIDs, fmt.Sprintf("role:%d:%s", roleID, priv))
	}

	var rv []*v2.Grant
	for _, perm := range permissions {
		var principalID *v2.ResourceId
		var err error
		if perm.Group {
			principalID, err = rs.NewResourceID(groupResourceType, perm.Principal)

			// add also format of group to expand this role to all members of group
			entitlementIDs = append(entitlementIDs, fmt.Sprintf("group:%s:%s", perm.Principal, GroupMembershipEntitlement))
		} else {
			principalID, err = rs.NewResourceID(userResourceType, perm.Principal)
		}

		if err != nil {
			return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to create principal resource id: %w", err)
		}

		// create grant for each permission
		rv = append(rv, grant.NewGrant(
			resource,
			RoleMembershipEntitlement,
			principalID,
			grant.WithAnnotation(&v2.GrantExpandable{EntitlementIds: entitlementIDs}),
		))
	}

	return rv, "", nil, nil
}

func newRoleBuilder(client *govmomi.Client) *roleBuilder {
	return &roleBuilder{
		client:       client,
		resourceType: roleResourceType,
	}
}
