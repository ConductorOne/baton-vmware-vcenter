package connector

import (
	"context"
	"errors"
	"fmt"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/ssoadmin/types"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *govmomi.Client
}

func userResource(user *types.AdminPersonUser) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"email":     user.Details.EmailAddress,
		"firstName": user.Details.FirstName,
		"lastName":  user.Details.LastName,
		"user_id":   user.Id.Name,
	}

	var status v2.UserTrait_Status_Status
	if user.Disabled || user.Locked {
		status = v2.UserTrait_Status_STATUS_DISABLED
	} else {
		status = v2.UserTrait_Status_STATUS_ENABLED
	}

	res, err := rs.NewUserResource(
		user.Id.Name,
		userResourceType,
		user.Id.Name,
		[]rs.UserTraitOption{
			rs.WithEmail(user.Details.EmailAddress, true),
			rs.WithUserProfile(profile),
			rs.WithStatus(status),
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	// adding a 5 second timeout to not block the sync too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sso, err := ssoadmin.NewClient(ctx, u.client.Client)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, "", nil, nil
		}

		return nil, "", nil, err
	}

	users, err := sso.FindPersonUsers(ctx, "*")
	if err != nil {
		return nil, "", nil, fmt.Errorf("vmware-vcenter-connector: failed to list users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		ur, err := userResource(&user)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *govmomi.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
