package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	VCenterServerURLField = field.StringField(
		"vcenter-server-url",
		field.WithDescription("The URL of the vCenter server."),
	)
	InsecureField = field.BoolField(
		"insecure",
		field.WithDescription("Whether to skip SSL verification."),
	)

	ConfigurationFields = []field.SchemaField{
		VCenterServerURLField,
		InsecureField,
	}

	Config = field.NewConfiguration(
		ConfigurationFields,
		field.WithConnectorDisplayName("VMware vCenter"),
	)
)
