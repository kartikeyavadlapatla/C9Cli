# C9Cli
Cloud Foundry - cf management

C9Cli helps maintaining  Cloud Foundry - Orgs/Spaces/Users/ASGs

C9Cli support following operations: 

-i init -e {endpoint} -u {user} -o {org} -s {space} -a true {enable ASGs} 
  
-i create-org -e {endpoint} -p {pwd} -k {path to C9Cli Folder}

-i create-space -e {endpoint} -p {pwd} -k {path to C9Cli Folder}

-i create-org-user -e {endpoint} -p {pwd} -k {path to C9Cli Folder}

-i create-space-user -e {endpoint} -p {pwd} -k {path to C9Cli Folder}

-i create-quota -e {endpoint} -p {pwd} -k {path to C9Cli Folder}

-i create-protected-org-asg -e {endpoint} -p {pwd} -k {path to C9Cli Folder}


Build: go build main.go

To Initialize configurations: 

C9Cli -i init will create config and sample ymls for further operation
  
  cf login config:

~/C9Cli/endpoint/config.yml - holds all cf login details created during init operation
  
  Yamls for Org/Space management:

~/C9Cli/endpoint/Org.yml

~/C9Cli/endpoint/Quota.yml

~/C9Cli/endpoint/ProtectedResources.yml
  
Operators/Users are required to fill the the above Yaml files before executing create operations. The Orgs and Quotas listed ProtectedResources.yml are not touched during the  cli execution even if they are added to Org/Quota.yml


ASGs:
Inital ASG setup such as creating platform wide default ASG needed to performed by operator. But cli can create/update running ASG to protected Orgs during create-protected-org-asg operation. All the ASGs has to be created in ~/C9Cli/mgmt/endpoint/AGSs/ folder with name <org>_<space>.json. At the Space level, Whenever create-space operation is triggered all space with updated with new created or existing ASGs. This command will also update changes made to JSON files

Pending Items:

- Delete Orgs/Spaces/Users/ASGs
