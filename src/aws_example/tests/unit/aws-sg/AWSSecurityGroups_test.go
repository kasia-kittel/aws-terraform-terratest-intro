package test

import (
	"testing"
	helpers "aws_example/tests/aws_helpers"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAWSSecurityGroup(t *testing.T) {
	
	t.Parallel()

	expectedVpcCidr :=  "10.10.0.0/16"
	expectedPublicSubnetCidr :=  "10.10.1.0/24"
	expectedPrivateSubnetCidr :=  "10.10.2.0/24"
	region := "eu-west-2"

	terraformOptions := &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"region" : region,
			"vpc-cidr": expectedVpcCidr,
			"public-subnet-cidr": expectedPublicSubnetCidr,
			"private-subnet-cidr": expectedPrivateSubnetCidr,
		},
	}

	terraform.InitAndApply(t, terraformOptions)
	defer terraform.Destroy(t, terraformOptions)

	vpcId := terraform.Output(t, terraformOptions, "vpc-id")
	publicSshGroupId := terraform.Output(t, terraformOptions, "public-ssh-sg-id")
	privateSshGroupId := terraform.Output(t, terraformOptions, "private-ssh-sg-id")

	t.Run("test if public-ssh-security-group gives public access to ssh", func(t *testing.T){
		sg := helpers.GetSecurityGroupById(t, publicSshGroupId, region)

		sshIpPermission := helpers.IpPermission {
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
		sg := helpers.GetSecurityGroupById(t, privateSshGroupId, region)

		privateSshIpPermission := helpers.IpPermission {
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