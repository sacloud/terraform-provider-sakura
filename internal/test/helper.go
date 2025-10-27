// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func SkipIfEnvIsNotSet(t *testing.T, key ...string) {
	for _, k := range key {
		if os.Getenv(k) == "" {
			t.Skipf("Environment valiable %q is not set", k)
		}
	}
}

func BuildConfigWithArgs(config string, args ...string) string {
	data := make(map[string]string)
	for i, v := range args {
		key := fmt.Sprintf("arg%d", i)
		data[key] = v
	}

	buf := bytes.NewBufferString("")
	err := template.Must(template.New("tmpl").Parse(config)).Execute(buf, data)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func RandomName() string {
	rand := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	return fmt.Sprintf("terraform-acctest-%s", rand)
}

func RandomPassword() string {
	return acctest.RandString(20)
}

func RandStringFromCharSet(length int, charSet string) string {
	if len(charSet) == 0 {
		charSet = acctest.CharSetAlpha
	}
	return acctest.RandStringFromCharSet(length, charSet)
}
