package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	vcenterconfig "github.com/conductorone/baton-vmware-vcenter/pkg/config"
	"github.com/conductorone/baton-vmware-vcenter/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-vmware-vcenter",
		getConnector,
		vcenterconfig.ConfigurationSchema,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.VMwareVCenter{}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	vcenterServerURL := v.GetString(vcenterconfig.VCenterServerURLField.FieldName)
	insecure := v.GetBool(vcenterconfig.InsecureField.FieldName)

	u, err := url.Parse(vcenterServerURL)
	if err != nil {
		l.Error("error parsing vCenter server URL", zap.Error(err))
		return nil, err
	}

	cb, err := connector.New(ctx, u, insecure)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return c, nil
}
