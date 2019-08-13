##Setting up Terraform and Terratest to work with AWS roles

### AWS Cli setup
1. Request AWS credentials on `#ask-aws` Slack channel.
2. Change password
3. Create new access key, save it in a secure place on your hard drive and inactivate the old one (IAM->Users->your username->Security Credentials)
4. Setup MFA for your username
5. Along with your credentials to account `889772146711` you will be also permitted to use `OrganizationAccountAccessRole` role
6. To switch to the role in the web interface use after signing in use: https://signin.aws.amazon.com/switchrole
7. Install and set up AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/install-macos.html
8. Configure CLI to use the OrganizationAccountAccessRole  role with MFA like explain here: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html#cli-configure-role-mfa

### Asdf
Best way to manage Terraform versions for the different project is to use along with the version manager like `asdf`. Asdf will not only facilitate managing different versions of the tools in the different project but it also installs the tools itself. 

1. Install `asdf` as described: https://asdf-vm.com/#/core-manage-asdf-vm
2. Install terraform plugin: `asdf plugin-add terraform https://github.com/Banno/asdf-hashicorp.git`
3. Install golang plugin: `asdf plugin-add golang https://github.com/kennyp/asdf-golang.git`
4. Install dep plugin: `asdf plugin-add golang-dep https://github.com/mcdan/asdf-golang-dep.git`

### Terraform
Create a directory for the Terraform project. Terratest (and more precisely Go) needs all source code files to be included in $GOPATH/src directory ie.  $GOPATH/src/aws_examples. Enter the working directory - for now aws_examples and run:
1. `asdf list-all terraform` lists all available version of Terraform; choose the version you want to use 
2. `asdf install terraform 0.12.5` to install the tool
3. `asdf local terraform 0.12.5` set the version for the project (it should create the `.tool-versions` file)

Terraform doesn't support MFA enabled AWS roles. To make it work to install and setup AWS-VAULT ( https://github.com/99designs/aws-vault)

When aws-vault is set up all terraform command need to prefix with `aws-vault exec [profile] -- ` ie:
`aws-vault exec test -- terraform apply`

### Terratest
Terratest needs Go compiler installed and $GOPATH set.
1. Set the `$GOPATH` ie. go to your project's directory and `export GOPATH=$PWD`
2. Install  and setup desired version of Go:  
   3. `asdf install golang 1.12.7`
   2. `asdf local golang 1.12.7`
3. Install  and setup desired version of Golang-dep:
  1.  `asdf install golang-dep v0.5.4`
  2.  `asdf local golang-dep v0.5.4`
3. In the `test` folder, create a `Gopkg.toml` file with the following content:
```
[[constraint]]
  name = "github.com/gruntwork-io/terratest"
  version = "0.17.4"  
```
5. Run `dep ensure`. This should load or necessary Go dependecies.

Terratest also need to be run with the aws-vault wrapper ie: `aws-vault exec test -- go test -v -timeout 15m`

###Useful readings
 https://blog.gruntwork.io/authenticating-to-aws-with-environment-variables-e793d6f6d02e

## Isolating Terratest executions

### Terraform workspaces
Some backends support multiple workspaces. It means that the state file is separate for each workspace, so execution of terraform scripts in different workspaces doesn't overwrite the state files.
For example in S3 backed the default workspace will be created in a path defined by the _key_ setting. Any other workspace will be created in the same bucket but within /env:/[workspace_name] path.

Listing existing workspaces: `terraform workspace list`
Showing current workspace: `terraform worspace show`
Changin to new workspace: `terraform workspace selecet [workspace_name]`




