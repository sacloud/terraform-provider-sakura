// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
