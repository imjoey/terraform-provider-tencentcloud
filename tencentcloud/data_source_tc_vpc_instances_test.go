package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceTencentCloudVpcV3Instances_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccDataSourceTencentCloudVpcInstances,

				Check: resource.ComposeTestCheckFunc(

					//id filter
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_vpc_instances.id_instances"),
					resource.TestCheckResourceAttr("data.tencentcloud_vpc_instances.id_instances", "instance_list.#", "1"),
					resource.TestCheckResourceAttr("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.name", "guagua_vpc_instance_test"),
					resource.TestCheckResourceAttr("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.vpc_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.is_default"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.is_multicast"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.dns_servers.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.subnet_ids.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.id_instances", "instance_list.0.create_time"),

					//name filter ,Every VPC with a "guagua_vpc_instance_test" name will be found
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_vpc_instances.name_instances"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.name"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.cidr_block"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.vpc_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.is_default"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.is_multicast"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.dns_servers.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.subnet_ids.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_instances.name_instances", "instance_list.0.create_time"),
				),
			},
		},
	})
}

const TestAccDataSourceTencentCloudVpcInstances = `
resource "tencentcloud_vpc" "foo" {
    name="guagua_vpc_instance_test"
    cidr_block="10.0.0.0/16"
}

data "tencentcloud_vpc_instances" "id_instances" {
	vpc_id="${tencentcloud_vpc.foo.id}"
}

data "tencentcloud_vpc_instances" "name_instances" {
	name="${tencentcloud_vpc.foo.name}"
}
`
