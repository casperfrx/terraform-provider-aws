// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPC_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_ipam_pool_id"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_netmask_length"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPC_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc1, vpc2, vpc3 awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc2),
					testAccCheckVPCIDsEqual(&vpc2, &vpc1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc3),
					testAccCheckVPCIDsEqual(&vpc3, &vpc2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPC_tags_computed(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags_computed,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "tags.eip"),
				),
			},
		},
	})
}

func TestAccVPC_tags_null(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags_null,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTags_zeroValue(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, ""),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", ""),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, ""),
					testAccVPCConfig_tags1(acctest.CtKey2, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", ""),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, acctest.CtValue1),
					testAccVPCConfig_tags1(acctest.CtKey2, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", ""),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTags_providerOnlyTestAccVPC_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", acctest.CtProviderValue1),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("providerkey1", acctest.CtProviderValue1, "providerkey2", "providervalue2"),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", acctest.CtValue1),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTags_updateToProviderOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, acctest.CtValue1),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_DefaultTags_updateToResourceOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, acctest.CtValue1),
					testAccVPCConfig_basic,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_nonOverlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", acctest.CtProviderValue1),
					testAccVPCConfig_tags1("resourcekey1", acctest.CtResourceValue1),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", acctest.CtResourceValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", acctest.CtProviderValue1),
					testAccVPCConfig_tags2("resourcekey1", acctest.CtResourceValue1, "resourcekey2", acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", acctest.CtResourceValue2),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccVPCConfig_tags1("resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_overlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", acctest.CtProviderValue1),
					testAccVPCConfig_tags1("overlapkey1", acctest.CtResourceValue1),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("overlapkey1", acctest.CtProviderValue1, "overlapkey2", "providervalue2"),
					testAccVPCConfig_tags2("overlapkey1", acctest.CtResourceValue1, "overlapkey2", acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", acctest.CtResourceValue2),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", acctest.CtProviderValue1),
					testAccVPCConfig_tags1("overlapkey1", acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", acctest.CtResourceValue2),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_duplicateTag(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccVPCConfig_tags1("overlapkey", "overlapvalue"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_moveDuplicateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccVPCConfig_tags1("overlapkey", "overlapvalue"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccVPCConfig_basic,
					acctest.ConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccVPCConfig_tags1("overlapkey", "overlapvalue"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPC_Tags_EmptyTag_OnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_Tags_EmptyTag_OnUpdate_Add(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccVPCConfig_tags2(acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_Tags_EmptyTag_OnUpdate_Replace(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccVPC_DynamicResourceTagsMergedWithLocals_ignoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags (merged with local.tags) are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccVPC_DynamicResourceTagsMergedWithLocals_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccVPC_DynamicResourceTags_ignoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccVPC_DynamicResourceTags_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPC_defaultAndIgnoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					testAccCheckVPCUpdateTags(ctx, &vpc, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPC_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					testAccCheckVPCUpdateTags(ctx, &vpc, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:   acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey") + testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				PlanOnly: true,
			},
			{
				Config:   acctest.ConfigIgnoreTagsKeys("ignorekey1") + testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPC_tenancy(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcDedicated awstypes.Vpc
	var vpcDefault awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_dedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_default(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpcDefault),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					testAccCheckVPCIDsEqual(&vpcDedicated, &vpcDefault),
				),
			},
			{
				Config: testAccVPCConfig_dedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
					testAccCheckVPCIDsNotEqual(&vpcDedicated, &vpcDefault),
				),
			},
		},
	})
}

func TestAccVPC_updateDNSHostnames(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_default(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCConfig_enableDNSHostnames(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/1301
func TestAccVPC_bothDNSOptionsSet(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_bothDNSOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/10168
func TestAccVPC_disabledDNSSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_disabledDNSSupport(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_enableNetworkAddressUsageMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_enableNetworkAddressUsageMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
		},
	})
}

func TestAccVPC_assignGeneratedIPv6CIDRBlockWithNetworkBorderGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	azDataSourceName := "data.aws_availability_zone.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USWest2RegionID)
			// https://docs.aws.amazon.com/vpc/latest/userguide/Extend_VPCs.html#local-zone:
			// "You can request the IPv6 Amazon-provided IP addresses and associate them with the network border group
			//  for a new or existing VPCs only for us-west-2-lax-1a and use-west-2-lax-1b. All other Local Zones don't support IPv6."
			testAccPreCheckLocalZoneAvailable(ctx, t, "us-west-2-lax-1") //lintignore:AWSAT003
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_cidr_block_network_border_group", azDataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv4BasicNetmask(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv4(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					testAccCheckVPCCIDRPrefix(&vpc, "28"),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv4BasicExplicitCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	cidr := "172.2.0.32/28"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv4ExplicitCIDR(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", cidr),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	ipamPoolResourceName := "aws_vpc_ipam_pool.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv6(rName, 28),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExistsV2(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block_network_border_group"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_ipam_pool_id", ipamPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "56"),
				),
			},
		},
	})
}

func testAccCheckVPCDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc" {
				continue
			}

			_, err := tfec2.FindVPCByIDV2(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCUpdateTags(ctx context.Context, vpc *awstypes.Vpc, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		return tfec2.UpdateTagsV2(ctx, conn, aws.ToString(vpc.VpcId), oldTags, newTags)
	}
}

func testAccCheckVPCCIDRPrefix(vpc *awstypes.Vpc, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.ToString(vpc.CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: got %s, expected %s", aws.ToString(vpc.CidrBlock), expected)
		}

		return nil
	}
}

func testAccCheckVPCIDsEqual(vpc1, vpc2 *awstypes.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(vpc1.VpcId) != aws.ToString(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs are not equal")
		}

		return nil
	}
}

func testAccCheckVPCIDsNotEqual(vpc1, vpc2 *awstypes.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(vpc1.VpcId) == aws.ToString(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs are equal")
		}

		return nil
	}
}

const testAccVPCConfig_basic = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
`

func testAccVPCConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccVPCConfig_tags_computed = `
resource "aws_eip" "test" {
  domain = "vpc"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    eip = aws_eip.test.public_ip
  }
}
`

const testAccVPCConfig_tags_null = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = null
  }
}
`

func testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals(localTagKey1, localTagValue1 string) string {
	return fmt.Sprintf(`
locals {
  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = merge(local.tags, {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  })

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`, localTagKey1, localTagValue1)
}

const testAccVPCConfig_ignoreChangesDynamicTags = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  }

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`

func testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName string, assignGeneratedIpv6CidrBlock bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = %[2]t
  cidr_block                       = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, assignGeneratedIpv6CidrBlock)
}

func testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName string, localZoneNetworkBorderGroup bool) string { // lintignore:AWSAT003
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "zone-type"
    values = ["local-zone"]
  }

  filter {
    name   = "opt-in-status"
    values = ["opted-in"]
  }

  filter {
    name   = "group-name"
    values = ["us-west-2-lax-1"]
  }
}

data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}

resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block     = true
  cidr_block                           = "10.1.0.0/16"
  ipv6_cidr_block_network_border_group = %[2]t ? data.aws_availability_zone.test.network_border_group : data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}
`, rName, localZoneNetworkBorderGroup)
}

func testAccVPCConfig_default(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_enableDNSHostnames(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_dedicatedTenancy(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  instance_tenancy = "dedicated"
  cidr_block       = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_bothDNSOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.2.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_disabledDNSSupport(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block         = "10.2.0.0/16"
  enable_dns_support = false

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_enableNetworkAddressUsageMetrics(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                           = "10.2.0.0/16"
  enable_network_address_usage_metrics = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_baseIPAMIPv4(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}
`, rName)
}

func testAccVPCConfig_ipamIPv4(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = %[2]d

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}

func testAccVPCConfig_ipamIPv4ExplicitCIDR(rName, cidr string) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = %[2]q

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, cidr))
}

func testAccVPCConfig_baseIPAMIPv6(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv6"
  ipam_scope_id  = aws_vpc_ipam.test.public_default_scope_id
  locale         = data.aws_region.current.name
  aws_service    = "ec2"

  public_ip_source              = "amazon"
  allocation_max_netmask_length = 128

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = 52
}
`, rName)
}

func testAccVPCConfig_ipamIPv6(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv6(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block          = "10.1.0.0/16"
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}
