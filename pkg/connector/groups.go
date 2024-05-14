package connector

import (
	"context"
	"errors"
	"fmt"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/ssoadmin/types"
)

const GroupMembershipEntitlement = "member"

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *govmomi.Client
}

func groupResource(group *types.AdminGroup) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"group_id":    group.Id.Name,
		"group_alias": group.Alias.Name,
	}

	res, err := rs.NewGroupResource(
		group.Id.Name,
		groupResourceType,
		group.Id.Name,
		[]rs.GroupTraitOption{
			rs.WithGroupProfile(profile),
		},
		rs.WithDescription(group.Details.Description),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (g *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	// adding a 5 second timeout to not block the sync too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sso, err := ssoadmin.NewClient(ctx, g.client.Client)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, "", nil, nil
		}

		return nil, "", nil, err
	}

	groups, err := sso.FindGroups(ctx, "*")
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to list groups: %w", err)
	}

	var rv []*v2.Resource
	for _, group := range groups {
		rr, err := groupResource(&group)
		if err != nil {
			return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to create group resource: %w", err)
		}

		rv = append(rv, rr)
	}

	return rv, "", nil, nil
}

func (g *groupBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	// add assignment entitlement for representing a group assignment to a principal
	aOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("Group %s %s", resource.DisplayName, GroupMembershipEntitlement)),
		ent.WithDescription(fmt.Sprintf("Group %s membership in vCenter Server", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, GroupMembershipEntitlement, aOptions...))

	return rv, "", nil, nil
}

func (g *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sso, err := ssoadmin.NewClient(ctx, g.client.Client)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, "", nil, nil
		}

		return nil, "", nil, err
	}

	members, err := sso.FindUsersInGroup(ctx, resource.Id.Resource, "*")
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to list group members: %w", err)
	}

	var rv []*v2.Grant
	for _, m := range members {
		userID, err := rs.NewResourceID(userResourceType, m.Id.Name)
		if err != nil {
			return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to create user resource id: %w", err)
		}

		rv = append(rv, grant.NewGrant(
			resource,
			GroupMembershipEntitlement,
			userID,
		))
	}

	return rv, "", nil, nil
}

func newGroupBuilder(client *govmomi.Client) *groupBuilder {
	return &groupBuilder{
		client:       client,
		resourceType: groupResourceType,
	}
}
