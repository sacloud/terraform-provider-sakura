package sakura

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func buildConfigWithArgs(config string, args ...string) string {
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

func randomName() string {
	rand := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	return fmt.Sprintf("terraform-acctest-%s", rand)
}

func randomPassword() string {
	return acctest.RandString(20)
}

func testCheckSakuraDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource is not exists: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("id is not set: %s", n)
		}
		return nil
	}
}
