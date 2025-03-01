package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMRoleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfig_basic(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "assume_role_policy", testAccRoleDataSourceConfig_AssumeRolePolicy_ExpectedJSON),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_date", resourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_session_duration", resourceName, "max_session_duration"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMRoleDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfig_tags(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "assume_role_policy", testAccRoleDataSourceConfig_AssumeRolePolicy_ExpectedJSON),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_date", resourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_session_duration", resourceName, "max_session_duration"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag2", "test-value2"),
				),
			},
		},
	})
}

const testAccRoleDataSourceConfig_AssumeRolePolicy_ExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      }
    }
  ]
}`

func testAccRoleDataSourceConfigBase() string {
	return `
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}
`
}

func testAccRoleDataSourceConfig_basic(roleName string) string {
	return acctest.ConfigCompose(
		testAccRoleDataSourceConfigBase(),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/testpath/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_role" "test" {
  name = aws_iam_role.test.name
}
`, roleName))
}

func testAccRoleDataSourceConfig_tags(roleName string) string {
	return acctest.ConfigCompose(
		testAccRoleDataSourceConfigBase(),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

data "aws_iam_role" "test" {
  name = aws_iam_role.test.name
}
`, roleName))
}
