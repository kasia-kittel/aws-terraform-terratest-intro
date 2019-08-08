package test

import (
	"testing"
	"strconv"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Vpc is an Amazon Virtual Private Cloud.
type Vpc struct {
	Id				string
	Name    	string 
	CidrBlock string
	Subnets 	[]Subnet 
}

// Subnet is a subnet in an availability zone.
type Subnet struct {
	Id               		string 		
	AvailabilityZone 		string 	
	CidrBlock 					string 
	MapPublicIpOnLaunch bool 
}

// Singe Route
type Route struct {
	DestinationCidrBlock 	string
	GatewayId 						string
	Origin 								string
	State 								string
}

func GetVpcByID(t *testing.T, vpcId string, region string) *Vpc {
	var vpcIDFilterName = "vpc-id"

	ec2Client := terratest_aws.NewEc2Client(t, region)

	vpcIdFilter := ec2.Filter{
		Name: &vpcIDFilterName, 
		Values: []*string{&vpcId},
	}
	
	vpcs, vpcs_err := ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{Filters: []*ec2.Filter{&vpcIdFilter}})
	if vpcs_err != nil {
		t.Fatal(vpcs_err)
	}

	numVpcs := len(vpcs.Vpcs)
	if numVpcs != 1 {
		t.Fatalf("Expected to find at most one VPC in region %s but found %s", region, strconv.Itoa(numVpcs)) 
	}

	vpc := vpcs.Vpcs[0]

	snets, snets_err := ec2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{Filters: []*ec2.Filter{&vpcIdFilter}})
	if snets_err != nil {
		t.Fatal(snets_err)
	}

	snetslen := len(snets.Subnets)
	subnets := make([]Subnet, snetslen)

	if snetslen > 0 {
		for i, s := range snets.Subnets {
			subnets[i] = Subnet{
				Id: aws.StringValue(s.SubnetId), 
				AvailabilityZone: aws.StringValue(s.AvailabilityZone), 
				CidrBlock:  aws.StringValue(s.CidrBlock), 
				MapPublicIpOnLaunch: aws.BoolValue(s.MapPublicIpOnLaunch),
			}
		}	
	}

	return &Vpc{
		Id: aws.StringValue(vpc.VpcId), 
		Name: terratest_aws.FindVpcName(vpc), 
		CidrBlock: aws.StringValue(vpc.CidrBlock), 
		Subnets: subnets,
	}
}

func GetRoutesForSubnet(t *testing.T, subnetId string, region string) []Route {
	var subnetIdFilterName = "association.subnet-id"

	ec2Client := terratest_aws.NewEc2Client(t, region)

	subnetIdFilter := ec2.Filter{
		Name: &subnetIdFilterName, 
		Values: []*string{&subnetId},
	}

	rts, rts_err := ec2Client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{Filters: []*ec2.Filter{&subnetIdFilter}})
	if rts_err != nil {
		t.Fatal(rts_err)
	}

	rtslen := len(rts.RouteTables[0].Routes)
	routes := make([]Route, rtslen)

	for i, r := range rts.RouteTables[0].Routes {
		routes[i] = Route{
			DestinationCidrBlock: aws.StringValue(r.DestinationCidrBlock), 
			GatewayId: aws.StringValue(r.GatewayId), 
			Origin:  aws.StringValue(r.Origin), 
			State: aws.StringValue(r.State),
		}
	}
	return routes
}

func VerifyContainPublicRoute(routes []Route, igwId string) bool {
	
	var found bool = false

	for _, r := range routes {
		if (r.DestinationCidrBlock == "0.0.0.0/0" && r.GatewayId == igwId){
			found = true
		}
	}

	return found
}

func TestTerraformAWSNetwork(t *testing.T) {
		t.Parallel()

	expectedVpcCidr :=  "10.10.0.0/16"
	expectedVpcName := "main-vpc-test"
	expectedPublicSubnetCidr :=  "10.10.1.0/24"
	region := "eu-west-2"

	terraformOptions := &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"region" : region,
			"main-vpc-cidr": expectedVpcCidr,
			"main-vpc-name": expectedVpcName,
			"public-subnet-cidr": expectedPublicSubnetCidr,
		},
	}

  terraform.InitAndApply(t, terraformOptions)
	defer terraform.Destroy(t, terraformOptions)

	vpcId := terraform.Output(t, terraformOptions, "main-vpc-id")
	publicSubnetId := terraform.Output(t, terraformOptions, "public-subnet-id")
	defaultIgwId := terraform.Output(t, terraformOptions, "default-igw-id")
	

	vpc := GetVpcByID(t, vpcId, region)
	subnets := vpc.Subnets
	
	t.Run("test if VPC is created with correct CIDR", func(t *testing.T) {
		assert.Equal(t, expectedVpcCidr, vpc.CidrBlock)
	})

	t.Run("test if VPC has one, correct subnetwork", func(t *testing.T) {
		numSubnets := len(subnets)
		
		assert.Equal(t, 1, numSubnets)
		assert.Equal(t, subnets[0].Id, publicSubnetId)
		assert.Equal(t, subnets[0].CidrBlock, expectedPublicSubnetCidr)
	
	})

	t.Run("test if the subnetwork is public", func(t *testing.T){
		// A public subnet is a subnet that's associated with a 
		// route table that has a route to an Internet gateway.
		// source: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html

		routes := GetRoutesForSubnet(t, publicSubnetId, region)
		assert.True(t, VerifyContainPublicRoute(routes, defaultIgwId))
		assert.True(t, subnets[0].MapPublicIpOnLaunch)
	}) 
}