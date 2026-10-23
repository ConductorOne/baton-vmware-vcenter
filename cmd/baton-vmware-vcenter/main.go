package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cfg "github.com/conductorone/baton-vmware-vcenter/pkg/config"
	"github.com/conductorone/baton-vmware-vcenter/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-vmware-vcenter",
		getConnector,
		cfg.Config,
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

// validateConfig returns an error if the loaded configuration isn't valid.
func validateConfig(v *viper.Viper) error {
	if v.GetString(cfg.VCenterServerURLField.FieldName) == "" {
		return status.Errorf(codes.InvalidArgument, "vCenter server URL must be provided, use --help for more information")
	}

	return nil
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	if err := validateConfig(v); err != nil {
		return nil, err
	}

	u, err := url.Parse(v.GetString(cfg.VCenterServerURLField.FieldName))
	if err != nil {
		l.Error("error parsing vCenter server URL", zap.Error(err))
		return nil, err
	}

	cb, err := connector.New(ctx, u, v.GetBool(cfg.InsecureField.FieldName))
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
