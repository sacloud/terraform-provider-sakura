// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ssh_key_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSSHKey_basic(t *testing.T) {
	resourceName := "sakura_ssh_key.foobar"
	rand := test.RandomName()

	var sshKey iaas.SSHKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSSHKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSSHKeyExists(resourceName, &sshKey),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "public_key", testAccPublicKey),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", testAccFingerprint),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSSHKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSSHKeyExists(resourceName, &sshKey),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "public_key", testAccPublicKeyUpd),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", testAccFingerprintUpd),
				),
			},
		},
	})
}

func testCheckSakuraSSHKeyExists(n string, ssh_key *iaas.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SSHKey ID is set")
		}

		client := test.AccClientGetter()
		keyOp := iaas.NewSSHKeyOp(client)

		foundSSHKey, err := keyOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundSSHKey.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found SSHKey: %s", rs.Primary.ID)
		}

		*ssh_key = *foundSSHKey
		return nil
	}
}

func testCheckSakuraSSHKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	keyOp := iaas.NewSSHKeyOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_ssh_key" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := keyOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists SSHKey: %s", rs.Primary.ID)
		}
	}

	return nil
}

var testAccSakuraSSHKey_basic = fmt.Sprintf(`
resource "sakura_ssh_key" "foobar" {
  name        = "{{ .arg0 }}"
  public_key  = "%s"
  description = "description"
}`, testAccPublicKey)

var testAccSakuraSSHKey_update = fmt.Sprintf(`
resource "sakura_ssh_key" "foobar" {
  name        = "{{ .arg0 }}-upd"
  public_key  = "%s"
  description = "description-upd"
}`, testAccPublicKeyUpd)

const testAccPublicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDq94EJW1+KAQLHNLC1KKdJq2aTIg/FSYeuKBiA7HWsCeG384uPo9afBS/+flXZfYzLlphQuS3HNC94CqlpNny3h7UdeUXcM0NOlhUBEuY5asVi60LnTAFCemlySXl0lQNKN/ly6oTVVe5auOFKl+wmRzJWETM71wg6908+n4M8BLzJcxoHWJ6m4KLXAS7WMbzsB+KyDQ/vp84hsvfhdgUj5NLt/WrVtdSY7CguNkV/P/ws7Fhi86qxu2V34e9/blZYTNqISTkwRriYYT0aCBB2vaN56pDcVzt+Wz41dXKymyheuTMPRUljFUfjIzgH5/vWSHpUEWDKTOwfjsCD6rv1`
const testAccFingerprint = `45:95:56:9c:ef:e3:0f:63:66:21:b4:2c:b9:53:00:00`

const testAccPublicKeyUpd = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDx8YEPX97c6vTm1q8s+bDZgalEPJdfYo73pgqLPCnfpqqmPmQzt4WPn713/dEV0erZWe796L8d36ub4w2E1Coqdn3UHal+h4peWyPYnSh1iBATDzYQwiJJ0yjAxGu2XR4IKfRBBISE2rw07GI7akUwCDqohE96vptqflH3zHwjJYp6tzai8h+Z/b2D5+F060jHVqNtkUWyoCmcrWsW53gr+o4NE1sBWJc9RF/TOmNg+2GnysCx9oPh0AssNXNCBYMtq2yH3yK6kCUXPCnNphL7LWc5/SUtZ6P4R1qeLubPmrM4rfn+H3oDfRjsCPVJ0+oNuTQBchN3BEqPAemeKthB`
const testAccFingerprintUpd = `61:08:83:1d:17:ee:26:c6:bb:fa:44:27:78:cb:cc:c8`
