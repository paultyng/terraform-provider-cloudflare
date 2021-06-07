package main

import (
	"github.com/cloudflare/terraform-provider-cloudflare/cloudflare"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudflare.Provider})
}
