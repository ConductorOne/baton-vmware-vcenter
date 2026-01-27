package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	VCenterServerURLField = field.StringField(
		"vcenter-server-url",
		field.WithDescription("The URL of the vCenter server"),
		field.WithRequired(true),
	)

	InsecureField = field.BoolField(
		"insecure",
		field.WithDescription("Whether to skip SSL verification"),
	)

	ConfigurationFields = []field.SchemaField{
		VCenterServerURLField,
		InsecureField,
	}

	ConfigurationSchema = field.Configuration{
		Fields: ConfigurationFields,
	}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(ConfigurationFields)
