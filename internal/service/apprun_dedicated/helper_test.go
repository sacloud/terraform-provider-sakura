// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	apprun_dedicated "github.com/sacloud/apprun-dedicated-api-go"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

// `acctest` has `RandTLSCert()` but its return value lacks SANs.
// We need to roll our own version
//
//nolint:nakedret
func OreSign(domain string) (certPEM, keyPEM []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}

	key, err := x509.MarshalECPrivateKey(priv)

	if err != nil {
		return
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	if err != nil {
		return
	}

	tpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: domain},
		DNSNames:              []string{domain, `*.` + domain},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &priv.PublicKey, priv)

	if err != nil {
		return
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key})
	return
}

var (
	globalClient      saclient.Client
	globalClusterID   string // filled below
	globalClusterName string = "tfacc-" + acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
)

func TestMain(m *testing.M) {
	var created *v1.CreatedCluster
	var err error
	ctx := context.Background()
	client, err := apprun_dedicated.NewClient(&globalClient)

	if err != nil {
		fmt.Printf("test setup failed: %q", err.Error())
		os.Exit(1)
	}

	api := cluster.NewClusterOp(client)
	spid, ok := os.LookupEnv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")

	if ok {
		created, err = api.Create(ctx, cluster.CreateParams{
			Name:               globalClusterName,
			ServicePrincipalID: spid,
			Ports:              make([]v1.CreateLoadBalancerPort, 0),
		})

		if err != nil {
			fmt.Printf("test setup failed: %q", err.Error())
			os.Exit(1)
		}

		globalClusterID = uuid.UUID(created.ClusterID).String()
	}

	ret := m.Run()

	if created != nil {
		_ = api.Delete(ctx, created.ClusterID)
	}

	os.Exit(ret)
}

func AccPreCheck(t *testing.T) func() {
	return func() {
		test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_APPRUN_DEDICATED_TEST", "SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
		test.SkipIfFakeModeEnabled(t)

		spid := os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
		if spid == "" {
			t.Fatalf("need valid SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID environment variable")
		}
	}
}
