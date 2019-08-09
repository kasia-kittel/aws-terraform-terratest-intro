package test

import (
    "testing"

    "github.com/gruntwork-io/terratest/modules/aws"
    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/stretchr/testify/assert"
)

func TestTerraformAWSEC2Generic(t *testing.T) {
    t.Parallel()

    expectedName :=  "aws-ec2-test"

    terraformOptions := &terraform.Options{
        // The path to where our Terraform code is located
        TerraformDir: ".",

        Vars: map[string]interface{}{
         "aws-ec2-name": expectedName,
        },
    }

    // This will run `terraform init` and `terraform apply` and fail the test if there are any errors
    terraform.InitAndApply(t, terraformOptions)

    // Cleanup at the end of the test
	defer terraform.Destroy(t, terraformOptions)
	

	// Run `terraform output` to get the value of an output variable
    instanceID := terraform.Output(t, terraformOptions, "aws-ec2-id")
    publicIp := terraform.Output(t, terraformOptions, "instance_public_ip")

	instanceTags := aws.GetTagsForEc2Instance(t, "eu-west-2", instanceID)
	nameTag := instanceTags["Name"]
    assert.Equal(t, expectedName, nameTag)
    
    assert.NotEmpty(t, publicIp)
}

