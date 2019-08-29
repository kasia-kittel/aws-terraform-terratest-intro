package test

import (
	"testing"

	helpers "aws_example/tests/aws_helpers"

  "github.com/gruntwork-io/terratest/modules/terraform"
  "github.com/stretchr/testify/assert"
)

func TestTerraformAWSNat(t *testing.T) {
	t.Parallel()

	region := "eu-west-2"
	privateSubnetCidr := "10.20.2.0/24"
	publicSubnetCidr := "10.20.1.0/24"

	terraformOptions := &terraform.Options{
        TerraformDir: ".",

        Vars: map[string]interface{}{
					"region" : region,
					"private-subnet-cidr": privateSubnetCidr,
					"public-subnet-cidr": publicSubnetCidr,
				},
    }

  // Cleanup at the end of the test
	defer terraform.Destroy(t, terraformOptions)

  // This will run `terraform init` and `terraform apply` and fail the test if there are any errors
  terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of an output variable
  natID := terraform.Output(t, terraformOptions, "nat-instance-id")
	vpcID := terraform.Output(t, terraformOptions, "test-vpc-id")
	privateSubnetID := terraform.Output(t, terraformOptions, "private-subnet-id")

	// TODO test specific nat setup of the ec2 instance
	t.Run("test if nat instance is created", func(t *testing.T) {
		assert.NotNil(t, natID)
	})

	t.Run("test if nat security group exists", func(t *testing.T) {
		sg := helpers.GetSecurityGroupByVpcIdAndName(t, vpcID, "NATSG", region)
		internetCidr := "0.0.0.0/0"

		httpIngressIpPermission := helpers.IpPermission {
			FromPort:	80,
			ToPort: 80,
			IpProtocol: "tcp",
			Ips: []string{privateSubnetCidr},
		}

		httpsIngressIpPermission := helpers.IpPermission {
			FromPort:	443,
			ToPort: 443,
			IpProtocol: "tcp",
			Ips: []string{privateSubnetCidr},
		}

		httpEgressIpPermission := helpers.IpPermissionEgress {
			FromPort:	80,
			ToPort: 80,
			IpProtocol: "tcp",
			Ips: []string{internetCidr},
		}

		httpsEgressIpPermission := helpers.IpPermissionEgress {
			FromPort:	443,
			ToPort: 443,
			IpProtocol: "tcp",
			Ips: []string{internetCidr},
		}

		assert.Equal(t, len(sg), 1)
		assert.Equal(t, len(sg[0].IpPermissions), 2)
		assert.Equal(t, len(sg[0].IpPermissionsEgress), 2)
		assert.True(t, sg[0].ContainsIpPermissions(t, httpIngressIpPermission))
		assert.True(t, sg[0].ContainsIpPermissions(t, httpsIngressIpPermission))
		assert.True(t, sg[0].ContainsIpPermissionsEgress(t, httpEgressIpPermission))
		assert.True(t, sg[0].ContainsIpPermissionsEgress(t, httpsEgressIpPermission))
	})

	t.Run("test if private subnet has outbound route traversing the nat", func(t *testing.T) {
		routes := helpers.GetRoutesForSubnet(t, privateSubnetID, region)		
		assert.True(t, helpers.VerifyContainNatRoute(routes, natID))
	})

	
}


