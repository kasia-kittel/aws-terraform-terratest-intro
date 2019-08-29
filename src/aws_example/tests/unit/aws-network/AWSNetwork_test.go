package test

import (
	"testing"
	helpers "aws_example/tests/aws_helpers"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAWSNetwork(t *testing.T) {
	t.Parallel()

	expectedVpcCidr :=  "10.10.0.0/16"
	expectedPublicSubnetCidr :=  "10.10.1.0/24"
	expectedPrivateSubnetCidr :=  "10.10.2.0/24"
	region := "eu-west-2"

	terraformOptions := &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"region" : region,
			"main-vpc-cidr": expectedVpcCidr,
			"public-subnet-cidr": expectedPublicSubnetCidr,
			"private-subnet-cidr": expectedPrivateSubnetCidr,
		},
	}

  terraform.InitAndApply(t, terraformOptions)
	defer terraform.Destroy(t, terraformOptions)

	vpcId := terraform.Output(t, terraformOptions, "main-vpc-id")
	publicSubnetId := terraform.Output(t, terraformOptions, "public-subnet-id")
	privateSubnetId := terraform.Output(t, terraformOptions, "private-subnet-id") 
	defaultIgwId := terraform.Output(t, terraformOptions, "default-igw-id")

	vpc := helpers.GetVpcByID(t, vpcId, region)
	subnets := vpc.Subnets
	
	t.Run("test if VPC is created with correct CIDR", func(t *testing.T) {
		assert.Equal(t, expectedVpcCidr, vpc.CidrBlock)
	})

	t.Run("test if VPC has two, correct subnetworks", func(t *testing.T) {
		numSubnets := len(subnets)
		
		assert.Equal(t, 2, numSubnets)
		assert.Equal(t, vpc.GetSubnetById(t, publicSubnetId).CidrBlock, expectedPublicSubnetCidr)
		assert.Equal(t, vpc.GetSubnetById(t, privateSubnetId).CidrBlock, expectedPrivateSubnetCidr)
	})

	t.Run("test if the public subnetwork is correctly setup", func(t *testing.T){
		routes := helpers.GetRoutesForSubnet(t, publicSubnetId, region)
		assert.True(t, helpers.VerifyContainPublicRoute(routes, defaultIgwId))
		assert.True(t, vpc.GetSubnetById(t, publicSubnetId).MapPublicIpOnLaunch)
	}) 

	t.Run("test if the private subnetwork is correctly setup", func(t *testing.T){
		routes := helpers.GetRoutesForSubnet(t, privateSubnetId, region)
		assert.False(t, helpers.VerifyContainPublicRoute(routes, defaultIgwId))
	})
}