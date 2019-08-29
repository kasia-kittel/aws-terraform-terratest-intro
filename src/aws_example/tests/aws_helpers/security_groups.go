package aws_helpers

import (
	"testing"
	"reflect"

	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
		
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// TODO improve the type to cover all possible cases
// how to make optional members?
// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#IpPermission
type IpPermission struct {
	FromPort					int64
	ToPort						int64
	IpProtocol				string
	Ips 							[]string //as https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#IpRange
}

type IpPermissionEgress struct {
	FromPort					int64
	ToPort						int64
	IpProtocol				string
	Ips 							[]string
}

type SecurityGroup struct {
	VpcId					string
	GroupId							string
	IpPermissions 			[]IpPermission
	IpPermissionsEgress []IpPermissionEgress
}

// TODO how to combine these two following function in one?
func (sg SecurityGroup) ContainsIpPermissions(t *testing.T, permission IpPermission) bool {
	for _, p:= range sg.IpPermissions {
		if reflect.DeepEqual(p, permission){
			return true
		}
	}
	return false
}

func (sg SecurityGroup) ContainsIpPermissionsEgress(t *testing.T, permission IpPermissionEgress) bool {
	for _, p:= range sg.IpPermissionsEgress {
		if reflect.DeepEqual(p, permission){
			return true
		}
	}
	return false
}


func GetSecurityGroup(t *testing.T, filters []*ec2.Filter, region string) []SecurityGroup {
	ec2Client := terratest_aws.NewEc2Client(t, region)

	sgResponse, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{Filters: filters})

	if err != nil {
		t.Fatal(err)
	}

	sgLen := len(sgResponse.SecurityGroups)

	// TODO error or return emtpy array?
	if sgLen == 0 {
		t.Fatal("No security group found!")
	}

	securityGroups := make([]SecurityGroup, sgLen)

	for j, s := range sgResponse.SecurityGroups {

		//ingress permissions
		permissionsLen := len(s.IpPermissions)
		permissions := make([]IpPermission, permissionsLen)
		
		if permissionsLen > 0 {
			for i, p := range s.IpPermissions {
					
				permissions[i] = IpPermission {
					FromPort:	aws.Int64Value(p.FromPort),
					ToPort: aws.Int64Value(p.ToPort),
					IpProtocol: aws.StringValue(p.IpProtocol),
				}
				
				if (p.IpRanges!=nil) {
					ips := make([]string, len(p.IpRanges))
					for n, ip := range p.IpRanges {
						ips[n] = *ip.CidrIp
					}
					permissions[i].Ips = ips
				}
			}
		}

		// egress permissions
		permissionsEgressLen := len(s.IpPermissionsEgress)
		permissionsEgress := make([]IpPermissionEgress, permissionsEgressLen)

		if permissionsEgressLen > 0 {
			for i, p := range s.IpPermissionsEgress {
				
				permissionsEgress[i] = IpPermissionEgress {
					FromPort:	aws.Int64Value(p.FromPort),
					ToPort: aws.Int64Value(p.ToPort),
					IpProtocol: aws.StringValue(p.IpProtocol),
				}

				if (p.IpRanges!=nil) {
					ips := make([]string, len(p.IpRanges))
					for n, ip := range p.IpRanges {
						ips[n] = *ip.CidrIp
					}
					permissionsEgress[i].Ips = ips
				}
			}
		}

		securityGroups[j] = SecurityGroup{
			VpcId:  aws.StringValue(s.VpcId),
			GroupId: aws.StringValue(s.GroupId), 
			IpPermissions: permissions,
			IpPermissionsEgress: permissionsEgress,
		}
	}

	return securityGroups
}

func GetSecurityGroupByVpcIdAndName(t *testing.T, vpcId string, tagNameValue string, region string) []SecurityGroup {
	var vpcIDFilterName = "vpc-id"
	vpcIdFilter := ec2.Filter{
		Name: &vpcIDFilterName, 
		Values: []*string{&vpcId},
	}
	
	var tagNameFilterName = "tag:Name"
	tagNameFilter := ec2.Filter{
		Name: &tagNameFilterName, 
		Values: []*string{&tagNameValue},
	}

	filters := []*ec2.Filter{&vpcIdFilter, &tagNameFilter}

	return GetSecurityGroup(t, filters, region)
}

func GetSecurityGroupById(t *testing.T, groupId string, region string) *SecurityGroup {
	var groupIDFilterName = "group-id"
	
	groupIdFilter := ec2.Filter{
		Name: &groupIDFilterName, 
		Values: []*string{&groupId},
	}

	filters := []*ec2.Filter{&groupIdFilter}

	// TODO improve it
	retval := GetSecurityGroup(t, filters, region)[0]
	return &retval
}

