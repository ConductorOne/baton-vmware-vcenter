package main

import (
	cfg "github.com/conductorone/baton-vmware-vcenter/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("vmware-vcenter", cfg.Config)
}
