package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	sakura "github.com/sacloud/terraform-provider-sakura/internal/provider"
	ver "github.com/sacloud/terraform-provider-sakura/version"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/sacloud/sakura",
		Debug:   debug,
	}
	err := providerserver.Serve(context.Background(), sakura.New(ver.Version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
