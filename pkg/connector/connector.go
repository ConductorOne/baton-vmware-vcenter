package connector

import (
	"context"
	"io"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/vmware/govmomi"
)

type VMwareVCenter struct {
	client *govmomi.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (vc *VMwareVCenter) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(vc.client),
		newRoleBuilder(vc.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (vc *VMwareVCenter) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (vc *VMwareVCenter) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "VMwareVCenter",
		Description: "Connector syncing VMware vCenter Server users, groups and roles to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (vc *VMwareVCenter) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, vCenterURL *url.URL, insecure bool) (*VMwareVCenter, error) {
	govmiClient, err := govmomi.NewClient(ctx, vCenterURL, insecure)
	if err != nil {
		return nil, err
	}

	return &VMwareVCenter{
		client: govmiClient,
	}, nil
}
