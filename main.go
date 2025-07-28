package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	sakura "github.com/sacloud/terraform-provider-sakuracloud/internal/provider"
	ver "github.com/sacloud/terraform-provider-sakuracloud/version"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/sacloud/sakuracloud",
		Debug:   debug,
	}
	err := providerserver.Serve(context.Background(), sakura.New(ver.Version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
