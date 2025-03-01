package ssmcontacts_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testContactChannelDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactChannelResourceName := "aws_ssmcontacts_contact_channel.test"
	dataSourceName := "data.aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "activation_status", contactChannelResourceName, "activation_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delivery_address.#", contactChannelResourceName, "delivery_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delivery_address.0.simple_address", contactChannelResourceName, "delivery_address.0.simple_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", contactChannelResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", contactChannelResourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "contact_id", contactChannelResourceName, "contact_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", contactChannelResourceName, "arn"),
				),
			},
		},
	})
}

func testAccContactChannelDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}

resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-for-%[2]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = %[3]q
  }

  name = %[2]q
  type = "EMAIL"
}

data "aws_ssmcontacts_contact_channel" "test" {
  arn = aws_ssmcontacts_contact_channel.test.arn
}
`, acctest.Region(), rName, acctest.DefaultEmailAddress)
}
