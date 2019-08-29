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
		"net"

    "github.com/gruntwork-io/terratest/modules/terraform"
		"github.com/gruntwork-io/terratest/modules/retry"
		"github.com/gruntwork-io/terratest/modules/ssh"
)

// TODO this is a bit fragile to problems with workspaces. Hot to improve it?
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

func TestTerraformSshHTTPConnection(t *testing.T,) {

	workspaceNameForTest := "test-workspace"
	region := "eu-west-3"

	terraformOptions := &terraform.Options{
		TerraformDir: "../../staging",
		
		Vars: map[string]interface{}{
			"region" : region,
		},

	}

	// Save current workspace
	currentWorkspace := terraform.RunTerraformCommand(t, terraformOptions, "workspace", "show")

	// Cleanup at the end of the test
	// TODO defer not called when InitAndApply failed on missing AMI. Why?
	defer CleanUp(t, terraformOptions, currentWorkspace, workspaceNameForTest)
	
	// Change to the test workspace
	terraform.WorkspaceSelectOrNew(t, terraformOptions, workspaceNameForTest)
	
	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	bastionIP := terraform.Output(t, terraformOptions, "bastion-ip")
	backendIP := terraform.Output(t, terraformOptions, "backend-private-ip")
	frontendPrivateIP := terraform.Output(t, terraformOptions, "frontend-private-ip")
	frontendPublicIP := terraform.Output(t, terraformOptions, "frontend-public-ip")
	natIP := terraform.Output(t, terraformOptions, "nat-ip")

	keyPair := LoadSshKeyFromFile(t, "/Users/kasia/.ssh/terratest-examples") //TODO test with ~

	bastion := ssh.Host{
		Hostname:    bastionIP,
		SshKeyPair:  keyPair,
		SshUserName: "ubuntu",
	}

	backend := ssh.Host{
		Hostname:    backendIP,
		SshKeyPair:  keyPair,
		SshUserName: "ubuntu",
	}

	frontendPrivate := ssh.Host{
		Hostname:    frontendPrivateIP,
		SshKeyPair:  keyPair,
		SshUserName: "ubuntu",
	}

	maxRetries := 3
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH and TCP tests.")

	// the ssh tests are redundand - left as examples
	t.Run("test if ssh connection via bastion to the backend host works", func(t *testing.T){
		expectedText := "Hello, World"
		command := fmt.Sprintf("echo -n '%s'", expectedText)

		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			actualText, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, backend, command)

			if ssh_err != nil {
				t.Fatalf("Problem connecting ssh to the backend: %s via bastion: %s: %v",backendIP, bastionIP, ssh_err)
			}

			if strings.TrimSpace(actualText) != expectedText {
				t.Errorf("Executing commands via ssh on the backend: %s not possible", backendIP)
			}
		
			return "", nil
		})
	})

	t.Run("test if ssh connection via bastion to the frontend host works", func(t *testing.T){
		expectedText := "Hello, World"
		command := fmt.Sprintf("echo -n '%s'", expectedText)

		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			actualText, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, frontendPrivate, command)

			if ssh_err != nil {
				t.Fatalf("Problem connecting ssh to the frontend: %s via bastion: %s: %v", frontendPrivateIP, bastionIP, ssh_err)
			}

			if strings.TrimSpace(actualText) != expectedText {
				t.Errorf("Executing commands via ssh on the frontend: %s not possible", frontendPrivateIP)
			}

			return "", nil
		})
	})

	// TODO how to prove that connecting directly to ssh on public IP of the frontend is blocked?

	// this approach is good when we test infra only, without services running on deployed instances
	t.Run("test HTTP connection to the frontend", func(t *testing.T){
		
		// remotely setup nc listener on port 80
		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			_, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, frontendPrivate, "sudo nohup nc -l 80 </dev/null >/dev/null 2>/dev/null &")
		
			if ssh_err != nil {
				t.Fatalf("Problem connecting ssh to the fontend %s: %v", frontendPrivateIP, ssh_err)
			}

			return "", nil
		})

		// check if is it possible to connect via tcp on port 80
		addr := fmt.Sprintf("%s:80", frontendPublicIP)
		_, err := net.Dial("tcp", addr)

		if err != nil {
			t.Errorf("TCP connection to %s not possible.", addr)
		}
	})

	t.Run("test HTTP connection between frontend and backend", func(t *testing.T){
		
		// remotely setup nc listener on port 80 on the backend host
		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			_, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, backend,  "sudo nohup nc -l 80 </dev/null >/dev/null 2>/dev/null &")
		
			if ssh_err != nil {
				t.Fatalf("Problem setting up nc listener in %s: %v", backendIP, ssh_err)
			}

			return "", nil
		})

		// connect to the listener via nc
		ncCommand := fmt.Sprintf("nc -zv %s 80", backendIP)
		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			ncResponse, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, frontendPrivate, ncCommand)
		
			if !strings.Contains(ncResponse, "succeeded") {
				t.Fatalf("TCP connection to %s from %s not possible: %s", backendIP, frontendPrivateIP, ncResponse)
			}

			// this will fail always when nc fails
			if ssh_err != nil {
				t.Fatalf("Problem connecting via ssh to the host %v", ssh_err)
			}

			return "", nil
		})
	})

	t.Run("test NAT backend", func(t *testing.T){

		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			ipfyOut, ssh_err := ssh.CheckPrivateSshConnectionE(t, bastion, backend,  "curl -s https://api.ipify.org?format=text")
		
			if strings.TrimSpace(ipfyOut) != natIP {
				t.Errorf("Connection not made via NAT: %s", ipfyOut)
			}

			if ssh_err != nil {
				t.Fatalf("Problem ssh %s: %v", ipfyOut, ssh_err)
			}

			return "", nil
		})
	})
}

