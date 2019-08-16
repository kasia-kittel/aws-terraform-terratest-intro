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

// Single Route
type Route struct {
	DestinationCidrBlock 	string
	GatewayId 						string
	Origin 								string
	State 								string
}

type IpPermission struct {
	FromPort			int64
	ToPort				int64
	IpProtocol		string
	Ips 					[]string
}

type SecurityGroup struct {
	VpcId					string
	IpPermissions []IpPermission
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

func (vpc Vpc) GetSubnetById(t *testing.T, subnetId string) *Subnet {

	for _, s := range vpc.Subnets {
		if(s.Id == subnetId){
			return &s
		}
	}

	//TODO fatal or error, what is better?
	t.Error("Subnet doesn't exists in given vpc")
	return nil
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

	if len(rts.RouteTables) == 0 {
		return make([]Route, 0)
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

func GetSecurityGroupById(t *testing.T, groupId string, region string) *SecurityGroup {
	var groupIDFilterName = "group-id"

	ec2Client := terratest_aws.NewEc2Client(t, region)

	groupIdFilter := ec2.Filter{
		Name: &groupIDFilterName, 
		Values: []*string{&groupId},
	}

	sg, sg_err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{Filters: []*ec2.Filter{&groupIdFilter}})
	
	if sg_err != nil {
		t.Fatal(sg_err)
	}

	sglen := len(sg.SecurityGroups)

	// TODO is it safe to assume there will be at most one group?
	if sglen >1 {
		t.Fatal("Too many security groups. This should never happen")
	}

	if sglen == 0 {
		t.Fatal("No security group found!")
	}

	iplen := len(sg.SecurityGroups[0].IpPermissions)
	ips := make([]IpPermission, iplen)

	if iplen > 0 {
		for i, ip := range sg.SecurityGroups[0].IpPermissions {
			
			iprs := make([]string, len(ip.IpRanges))
			for ii, ipr := range ip.IpRanges {
				iprs[ii] = *ipr.CidrIp
			}

			ips[i] = IpPermission {
				FromPort:	aws.Int64Value(ip.FromPort),
				ToPort: aws.Int64Value(ip.ToPort),
				IpProtocol: aws.StringValue(ip.IpProtocol),
				Ips: iprs,
			}
		}	
	}

	return &SecurityGroup{
		VpcId:  aws.StringValue(sg.SecurityGroups[0].VpcId),
		IpPermissions: ips,
	}
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
		//t.Parallel()

	expectedVpcCidr :=  "10.10.0.0/16"
	expectedVpcName := "main-vpc-test"
	expectedPublicSubnetCidr :=  "10.10.1.0/24"
	expectedPrivateSubnetCidr :=  "10.10.2.0/24"
	region := "eu-west-2"

	terraformOptions := &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"region" : region,
			"main-vpc-cidr": expectedVpcCidr,
			"main-vpc-name": expectedVpcName,
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
	publicSshGroupId := terraform.Output(t, terraformOptions, "public-ssh-sg-id")
	privateSshGroupId := terraform.Output(t, terraformOptions, "private-ssh-sg-id")

	vpc := GetVpcByID(t, vpcId, region)
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
		// A public subnet is a subnet that's associated with a 
		// route table that has a route to an Internet gateway.
		// source: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html

		routes := GetRoutesForSubnet(t, publicSubnetId, region)
		assert.True(t, VerifyContainPublicRoute(routes, defaultIgwId))
		assert.True(t, vpc.GetSubnetById(t, publicSubnetId).MapPublicIpOnLaunch)
	}) 

	t.Run("test if the private subnetwork is correctly setup", func(t *testing.T){
		routes := GetRoutesForSubnet(t, privateSubnetId, region)
		assert.False(t, VerifyContainPublicRoute(routes, defaultIgwId))
	})

	t.Run("test if public-ssh-security-group gives public access to ssh", func(t *testing.T){
		sg := GetSecurityGroupById(t, publicSshGroupId, region)

		sshIpPermission := IpPermission {
			FromPort:	22,
			ToPort: 22,
			IpProtocol: "tcp",
			Ips: []string{"0.0.0.0/0"},
		}
		assert.Equal(t, sg.VpcId, vpcId)
		assert.Equal(t, len(sg.IpPermissions), 1)
		assert.Equal(t, sg.IpPermissions[0], sshIpPermission)
	})

	t.Run("test if private-ssh-security-group gives access to ssh only from public subnet", func(t *testing.T){
		sg := GetSecurityGroupById(t, privateSshGroupId, region)

		privateSshIpPermission := IpPermission {
			FromPort:	22,
			ToPort: 22,
			IpProtocol: "tcp",
			Ips: []string{expectedPublicSubnetCidr},
		}
		assert.Equal(t, sg.VpcId, vpcId)
		assert.Equal(t, len(sg.IpPermissions), 1)
		assert.Equal(t, sg.IpPermissions[0], privateSshIpPermission)
	})
}