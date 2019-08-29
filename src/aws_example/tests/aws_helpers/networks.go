package aws_helpers

import (
	"testing"
	"strconv"

	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
		
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

// TODO add more info from https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#Route
// Single Route
type Route struct {
	DestinationCidrBlock 	string
	InstanceId 						string
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
			InstanceId: aws.StringValue(r.InstanceId), 
			GatewayId: aws.StringValue(r.GatewayId), 
			Origin:  aws.StringValue(r.Origin), 
			State: aws.StringValue(r.State),
		}
	}

	return routes
}

// A public subnet is a subnet that's associated with a 
// route table that has a route to an Internet gateway.
// source: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html
func VerifyContainPublicRoute(routes []Route, igwId string) bool {
	var found bool = false

	for _, r := range routes {
		if (r.DestinationCidrBlock == "0.0.0.0/0" && r.GatewayId == igwId){
			found = true
		}
	}

	return found
}

func VerifyContainNatRoute(routes []Route, natId string) bool {
	var found bool = false

	for _, r := range routes {
		if (r.DestinationCidrBlock == "0.0.0.0/0" && r.InstanceId == natId && r.State == "active"){
			found = true
		}
	}

	return found
}