package test

import (
		"bufio"
    "crypto/x509"
    "encoding/pem"
    "os"
		"testing"
		"strings"
		"fmt"
		"time"

    "github.com/gruntwork-io/terratest/modules/terraform"
		"github.com/gruntwork-io/terratest/modules/retry"
		"github.com/gruntwork-io/terratest/modules/ssh"
)

func CleanUp(t *testing.T, terraformOptions *terraform.Options, 
	currentWorkspace string, testWorkspace string){
	
	terraform.Destroy(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, currentWorkspace)
	// also removes state file stored in remote backend
	terraform.RunTerraformCommand(t, terraformOptions, "workspace", "delete", testWorkspace)
}

// loads the private key into ssh.KeyPair struct that can be used with the
// Terratest ssh module
// to generate compatible key pair use: `ssh-keygen -t rsa -b 4096 -m PEM`
func LoadSshKeyFromFile(t *testing.T, pathToPrivateKey string) *ssh.KeyPair {
	// this is not safe!
	privateKeyFile, file_err := os.Open(pathToPrivateKey)
	
	if file_err != nil {
    t.Fatalf("Not possible to load the key file: %v", file_err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyFile)
	_, _ = buffer.Read(pembytes)
	data, _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()

	privateKeyImported, pk_err := x509.ParsePKCS1PrivateKey(data.Bytes) //*rsa.PrivateKey
	if pk_err != nil {
			t.Fatalf("Key parsing not possible: %v.", pk_err)
	}

	keyPemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKeyImported),
	}

	keyPem := string(pem.EncodeToMemory(keyPemBlock))
	return &ssh.KeyPair{PublicKey: "", PrivateKey: keyPem}

}

func TestTerraformSshConnection(t *testing.T,) {

	workspaceNameForTest := "test-workspace"
	region := "eu-west-3"

	terraformOptions := &terraform.Options{
		TerraformDir: "../../staging",
		
		Vars: map[string]interface{}{
			"region" : region,
		},

	}

	// Safe current workspace
	currentWorkspace := terraform.RunTerraformCommand(t, terraformOptions, "workspace", "show")

	// Change to the test workspace
	terraform.WorkspaceSelectOrNew(t, terraformOptions, workspaceNameForTest)
	
	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Cleanup at the end of the test
	// TODO defer not called when InitAndApply failed on missing AMI. Why?
	defer CleanUp(t, terraformOptions, currentWorkspace, workspaceNameForTest)
	
	publicInstanceIP := terraform.Output(t, terraformOptions, "bastion_ip")
	privateInstanceIP := terraform.Output(t, terraformOptions, "host_private_ip")


	keyPair := LoadSshKeyFromFile(t, "/Users/kasia/.ssh/terratest-examples") //TODO test with ~

	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshKeyPair:  keyPair,
		SshUserName: "ubuntu",
	}

	privateHost := ssh.Host{
		Hostname:    privateInstanceIP,
		SshKeyPair:  keyPair,
		SshUserName: "ubuntu",
	}

	t.Run("test if ssh connection via bastion to the private host works", func(t *testing.T){
		expectedText := "Hello, World"
		command := fmt.Sprintf("echo -n '%s'", expectedText)

		maxRetries := 30
		timeBetweenRetries := 5 * time.Second
		description := fmt.Sprintf("SSH to private host %s via public host %s", privateInstanceIP, publicInstanceIP)

		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			actualText, ssh_err := ssh.CheckPrivateSshConnectionE(t, publicHost, privateHost, command)

			if ssh_err != nil {
				t.Fatalf("Problem connecting via ssh to the host %v", ssh_err)
			}
		
			if strings.TrimSpace(actualText) != expectedText {
				t.Fail()
			}
			return "", nil
		})

	})
	
}

