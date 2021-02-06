package main

//import "C"
import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Orglist struct {
	Org []struct {
		Name     string `yaml:"Name"`
		Quota    string `yaml:"Quota"`
		OrgUsers struct {
			LDAP []struct {
				Name string `yaml:"Name"`
				Role string `yaml:"Role"`
			} `yaml:"LDAP"`
			SSO []struct {
				Name string `yaml:"Name"`
				Role string `yaml:"Role"`
			} `yaml:"SSO"`
			UAA []struct {
				Name string `yaml:"Name"`
				Role string `yaml:"Role"`
			} `yaml:"UAA"`
		} `yaml:"OrgUsers"`
		Spaces []struct {
			Name       string `yaml:"Name"`
			IsolationSeg string `"IsolationSeg"`
			SpaceUsers struct {
				LDAP []struct {
					Name string `yaml:"Name"`
					Role string `yaml:"Role"`
				} `yaml:"LDAP"`
				SSO []struct {
					Name string `yaml:"Name"`
					Role string `yaml:"Role"`
				} `yaml:"SSO"`
				UAA []struct {
					Name string `yaml:"Name"`
					Role string `yaml:"Role"`
				} `yaml:"UAA"`
			} `yaml:"SpaceUsers"`
		} `yaml:"Spaces"`
	} `yaml:"Org"`
}
type Quotalist struct {
	Quota []struct {
		Name        string `yaml:"Name"`
		MemoryLimit string `yaml:"memory_limit"`
		AllowPaidPlans       bool `yaml:"allow_paid_plans"`
		AppInstanceLimit     int `yaml:"app_instance_limit"`
		ServiceInstanceLimit int `yaml:"service_instance_limit"`
	} `yaml:"quota"`
}
type ProtectedList struct {
	Org   []string `yaml:"Org"`
	Quota []string `yaml:"quota"`
	DefaultRunningSecurityGroup string   `yaml:"DefaultRunningSecurityGroup"`
}
type InitClusterConfigVals struct {
	ClusterDetails struct {
		EndPoint  string `yaml:"EndPoint"`
		User      string `yaml:"User"`
		Org       string `yaml:"Org"`
		Space     string `yaml:"Space"`
		EnableASG bool   `yaml:"EnableASG"`
	} `yaml:"ClusterDetails"`
}

func main()  {

	var endpoint, user, pwd, org, space, asg, operation, cpath, ostype string
	var ospath io.Writer

	flag.StringVar(&endpoint, "e", "api.sys-domain", "Use with init operation, Provide PCF Endpoint")
	flag.StringVar(&user, "u", "user", "Use with init operation, Provide UserName")
	flag.StringVar(&pwd, "p", "pwd", "Use with all operation, Provide Password")
	flag.StringVar(&org, "o", "org", "Use with init operation, Provide Org")
	flag.StringVar(&space, "s", "space", "Use with init operation, Provide Space")
	flag.StringVar(&asg, "a", "true", "Use with init operation, Enable ASGs ?.")
	flag.StringVar(&operation, "i", "init", "Provide Operation to be performed: init, create-{org,space,org-user,space-user,quota, ")
	flag.StringVar(&cpath, "k", ".", "Provide path to configs, i.e, to C9Cli folder, use with all operations")
	flag.Parse()

	ClusterName := strings.ReplaceAll(endpoint, ".", "-")

	fmt.Printf("Operation: %v\n", operation)

	oscmd := exec.Command("cmd", "/C","echo","%systemdrive%%homepath%")
	if _, err := oscmd.Output(); err != nil{
		fmt.Println("Checking OS")
		fmt.Println("command: ", oscmd)
		fmt.Println("command: ", oscmd.Stdout)
		fmt.Println("Err Code: ", err)
		oscmd = exec.Command("sh", "-c", "echo","$HOME")
		if _, err := oscmd.Output(); err != nil{
			fmt.Println("Checking OS failed - Can't find Underlying OS")
			fmt.Println("command: ", oscmd)
			fmt.Println("command: ", oscmd.Stdout)
			fmt.Println("Err Code: ", err)
			panic(err)
		} else {
			fmt.Println("command: ", oscmd)
			ospath = oscmd.Stdout
			fmt.Println("PATH: ", ospath)
			fmt.Println("Checking OS - Setting up for Mac/Linux/Ubuntu")
			ostype = "non-windows"
		}
	} else {
		fmt.Println("command: ", oscmd)
		ospath = oscmd.Stdout
		fmt.Println("PATH: ", ospath)
		fmt.Println("Checking OS - Setting up for Windows")
		ostype = "windows"
		//panic(err)
	}

	if operation == "init" {

		fmt.Println("Initializing C9Cli")

		fmt.Printf("ClusterName: %v\n", ClusterName)
		fmt.Printf("EndPoint: %v\n", endpoint)
		fmt.Printf("User: %v\n", user)
		//fmt.Printf("Pwd: %v\n", pwd)
		fmt.Printf("Org: %v\n", org)
		fmt.Printf("Space: %v\n", space)
		fmt.Printf("EnableASG: %v\n", asg)
		fmt.Printf("Path: %v\n", cpath)
		Init(ClusterName, endpoint, user, org, space, asg, cpath)

	} else if operation == "create-org"{

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName, pwd, cpath)
		CreateOrUpdateOrgs (ClusterName, cpath)

	} else if operation == "create-quota" {

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateQuotas(ClusterName, cpath)

	} else if operation == "create-org-user" {

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection(ClusterName,  pwd, cpath)
		CreateOrUpdateOrgUsers(ClusterName, cpath)
	} else if operation == "create-space"{

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateSpaces (ClusterName, cpath, ostype)

	} else if operation == "create-space-user"{

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateSpaceUsers (ClusterName, cpath)

	} else if operation == "create-protected-org-asg"{

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateProtOrgAsg (ClusterName, cpath, ostype)
	} else if operation == "create-all" {
		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateProtOrgAsg (ClusterName, cpath, ostype)
		CreateOrUpdateQuotas(ClusterName, cpath)
		CreateOrUpdateOrgs (ClusterName, cpath)
		CreateOrUpdateOrgUsers(ClusterName, cpath)
		CreateOrUpdateSpaces (ClusterName, cpath, ostype)
		CreateOrUpdateSpaceUsers (ClusterName, cpath)
	} else {
		fmt.Println("Provide Valid input operation")
	}
}
func CreateOrUpdateProtOrgAsg(clustername string, cpath string, ostype string) {

	var ProtectedOrgs ProtectedList
	var ASGpath string
	//ASGpath = cpath+"/C9Cli/"+clustername+"/ASGs/"
	ProtectedOrgsYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/C9Cli/"+clustername+"/config.yml"

	fileConfigYml, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileConfigYml), &InitClusterConfigVals)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	if ostype == "windows" {
		ASGpath = cpath+"\\C9Cli\\"+clustername+"\\ASGs\\"
		//fmt.Println("Hi :",ASGPath)
	} else {
		ASGpath = cpath+"/C9Cli/"+clustername+"/ASGs/"
	}

	LenProtectedOrgs := len(ProtectedOrgs.Org)
	var check *exec.Cmd
	ASGfile := ASGpath+ProtectedOrgs.DefaultRunningSecurityGroup+".json"
	if InitClusterConfigVals.ClusterDetails.EnableASG == true {
		fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)

		if ostype == "windows" {
			check = exec.Command("powershell", "-command","Get-Content", ASGfile)
		} else {
			check = exec.Command("cat", ASGfile)
		}

		if _, err := check.Output(); err != nil {
			fmt.Println("ASG for Protected Orgs: ", ProtectedOrgs.DefaultRunningSecurityGroup)
			fmt.Println("command: ", check)
			fmt.Println("Err: ", check.Stdout)
			fmt.Println("Err Code: ", err)
			fmt.Println("No Default ASG file provided in path for Protected Orgs")
		} else {
			fmt.Println("command: ", check)
			fmt.Println(check.Stdout)
			fmt.Println("ASG for Protected Orgs: ", ProtectedOrgs.DefaultRunningSecurityGroup)
			checkdasg := exec.Command("cf", "security-group", ProtectedOrgs.DefaultRunningSecurityGroup)
			if _, err := checkdasg.Output(); err != nil {
				fmt.Println("command: ", checkdasg)
				fmt.Println("Err: ", checkdasg.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Default ASG doesn't exist, Creating default ASG")
				createdasg := exec.Command("cf", "create-security-group", ProtectedOrgs.DefaultRunningSecurityGroup, ASGfile)
				if _, err := createdasg.Output(); err != nil {
					fmt.Println("command: ", createdasg)
					fmt.Println("Err: ", createdasg.Stdout)
					fmt.Println("Err Code: ", err)
					fmt.Println("Creating default ASG failed")
				} else {
					fmt.Println("command: ", createdasg)
					fmt.Println(createdasg.Stdout)
				}
			} else {
				fmt.Println("Default ASG exist, Updating default ASG")
				updatedefasg := exec.Command("cf", "update-security-group", ProtectedOrgs.DefaultRunningSecurityGroup, ASGfile)
				if _, err := updatedefasg.Output(); err != nil {
					fmt.Println("command: ", updatedefasg)
					fmt.Println("Err: ", updatedefasg.Stdout)
					fmt.Println("Err Code: ", err)
					fmt.Println("Default ASG not updated")
				} else {
					fmt.Println("command: ", updatedefasg)
					fmt.Println(updatedefasg.Stdout)
				}
			}
		}

		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			fmt.Println("ASG for Protected Orgs: ", ProtectedOrgs.DefaultRunningSecurityGroup)
			bindasg := exec.Command("cf", "bind-security-group", ProtectedOrgs.DefaultRunningSecurityGroup, ProtectedOrgs.Org[p], "--lifecycle", "running")
			if _, err := bindasg.Output(); err != nil{
				fmt.Println("command: ", bindasg)
				fmt.Println("Err: ", bindasg.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Failed to bind to protected Org")
			} else {
				fmt.Println("command: ", bindasg)
				fmt.Println(bindasg.Stdout)
			}
		}
	} else {
		fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
		fmt.Println("ASGs not enabled")
	}
}
func SetupConnection(clustername string, pwd string, cpath string) error {

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/C9Cli/"+clustername+"/config.yml"

	fileConfigYml, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileConfigYml), &InitClusterConfigVals)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Endpoint: %v\n", InitClusterConfigVals.ClusterDetails.EndPoint)
	fmt.Printf("User: %v\n", InitClusterConfigVals.ClusterDetails.User)
	fmt.Printf("Pwd: %v\n", pwd)
	fmt.Printf("Org: %v\n", InitClusterConfigVals.ClusterDetails.Org)
	fmt.Printf("Space: %v\n", InitClusterConfigVals.ClusterDetails.Space)
	//fmt.Println(InitClusterConfigVals.ClusterDetails.EndPoint)

	cmd := exec.Command("cf", "login", "-a", InitClusterConfigVals.ClusterDetails.EndPoint, "-u", InitClusterConfigVals.ClusterDetails.User, "-p", pwd, "-o", InitClusterConfigVals.ClusterDetails.Org, "-s", InitClusterConfigVals.ClusterDetails.Space, "--skip-ssl-validation")
	if _, err := cmd.Output(); err != nil{
		fmt.Println("Connection failed")
		fmt.Println("command: ", cmd)
		fmt.Println("Err: ", cmd.Stdout)
		fmt.Println("Err Code: ", err)
		panic(err)
	} else {
		fmt.Println("Connection Passed")
		fmt.Println("command: ", cmd)
		fmt.Println(cmd.Stdout)
	}
	return err
}
func CreateOrUpdateOrgs(clustername string, cpath string) error {

	var Orgs Orglist
	var ProtectedOrgs ProtectedList

	OrgsYml := cpath+"/C9Cli/"+clustername+"/Org.yml"
	fileOrgYml, err := ioutil.ReadFile(OrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
	if err != nil {
		panic(err)
	}


	ProtectedOrgsYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}


	LenOrgs := len(Orgs.Org)
	LenProtectedOrgs := len(ProtectedOrgs.Org)

	for i := 0; i < LenOrgs; i++ {
		var count, totalcount int
		fmt.Println("Org: ", Orgs.Org[i].Name)
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == Orgs.Org[i].Name {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count
	//	fmt.Println(totalcount)

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")
			guid := exec.Command("cf", "org", Orgs.Org[i].Name, "--guid")
			if _, err := guid.Output(); err == nil{

				fmt.Println("command: ", guid)
				fmt.Println("Org exists: ", guid.Stdout)
				fmt.Println("Updating Org quota")
				SetQuota := exec.Command("cf", "set-quota", Orgs.Org[i].Name, Orgs.Org[i].Quota)
				if _, err := SetQuota.Output(); err != nil{
					fmt.Println("command: ", SetQuota)
					fmt.Println("Err: ", SetQuota.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", SetQuota)
					fmt.Println(SetQuota.Stdout)
				}
			} else {
				fmt.Println("command: ", guid)
				fmt.Println("Err: ", guid.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Pulling Guid Id: ", guid.Stdout)
				fmt.Println("Org doesn't exists, Creating Org")
				createorg := exec.Command("cf", "create-org", Orgs.Org[i].Name)
				if _, err := createorg.Output(); err != nil{
					fmt.Println("command: ", createorg)
					fmt.Println("Err: ", createorg.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", createorg)
					fmt.Println(createorg.Stdout)
				}
				attachquota := exec.Command("cf", "set-quota", Orgs.Org[i].Name, Orgs.Org[i].Quota)
				if _, err := attachquota.Output(); err != nil{
					fmt.Println("command: ", attachquota)
					fmt.Println("Err: ", attachquota.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", attachquota)
					fmt.Println(attachquota.Stdout)
				}
			}
		} else {
			fmt.Println("This is a protected Org")
		}
	}
	return err
}
func CreateOrUpdateSpaces(clustername string, cpath string, ostype string) error {

	var Orgs Orglist
	var ProtectedOrgs ProtectedList

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/C9Cli/"+clustername+"/config.yml"

	fileConfigYml, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileConfigYml), &InitClusterConfigVals)
	if err != nil {
		panic(err)
	}
	var ASGPath, OrgsYml string

	//fmt.Printf("Endpoint: %v\n", InitClusterConfigVals.ClusterDetails.EndPoint)

	if ostype == "windows" {
		ASGPath = cpath+"\\C9Cli\\"+clustername+"\\ASGs\\"
		OrgsYml = cpath+"\\C9Cli\\"+clustername+"\\Org.yml"
		fmt.Println("Hi :",ASGPath)
	} else {
		ASGPath = cpath+"/C9Cli/"+clustername+"/ASGs/"
		OrgsYml = cpath+"/C9Cli/"+clustername+"/Org.yml"
	}

	fileOrgYml, err := ioutil.ReadFile(OrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenOrgs := len(Orgs.Org)
	LenProtectedOrgs := len(ProtectedOrgs.Org)


	for i := 0; i < LenOrgs; i++ {

		var count, totalcount int
		fmt.Println("Org: ", Orgs.Org[i].Name)
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == Orgs.Org[i].Name {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count
		//fmt.Println(totalcount)

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")
			guid := exec.Command("cf", "org", Orgs.Org[i].Name, "--guid")

			if _, err := guid.Output(); err == nil {

				fmt.Println("command: ", guid)
				fmt.Println("Org exists: ", guid.Stdout)
				SpaceLen := len(Orgs.Org[i].Spaces)

				TargetOrg := exec.Command("cf", "t", "-o", Orgs.Org[i].Name)
				if _, err := TargetOrg.Output(); err == nil {
					fmt.Println("command: ", TargetOrg)
					fmt.Println("Targeting: ", TargetOrg.Stdout)
				} else {
					fmt.Println("command: ", TargetOrg)
					fmt.Println("Err: ", TargetOrg.Stdout)
					fmt.Println("Err Code: ", err)
				}

				for j := 0; j < SpaceLen; j++ {

					fmt.Println("Creating Spaces")

					guid = exec.Command("cf", "space", Orgs.Org[i].Spaces[j].Name, "--guid")

					if _, err := guid.Output(); err == nil{

						fmt.Println("command: ", guid)
						fmt.Println("Space exists: ", guid.Stdout)

						fmt.Println("Enabling Space Isolation Segment")

						iso := exec.Command("cf", "enable-org-isolation", Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].IsolationSeg)

						if _, err := iso.Output(); err != nil {
							fmt.Println("command: ", iso)
							fmt.Println("Err: ", iso.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", iso)
							fmt.Println(iso.Stdout)
						}

						isospace := exec.Command("cf", "set-space-isolation-segment", Orgs.Org[i].Spaces[j].Name, Orgs.Org[i].Spaces[j].IsolationSeg)

						if _, err := isospace.Output(); err != nil {
							fmt.Println("command: ", isospace)
							fmt.Println("Err: ", isospace.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", isospace)
							fmt.Println(isospace.Stdout)
						}

						fmt.Println("Creating or updating ASGs")


//						cf enable-org-isolation ORG-NAME SEGMENT-NAME
//						cf set-space-isolation-segment SPACE-NAME SEGMENT-NAME

						if InitClusterConfigVals.ClusterDetails.EnableASG == true {
							fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
							CreateOrUpdateASGs(Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].Name, ASGPath, ostype)
						} else {
							fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
							fmt.Println("ASGs not enabled")
						}
					} else {
						fmt.Println("command: ", guid)
						fmt.Println("Pulling Space Guid ID: ", guid.Stdout )
						fmt.Println("Creating Space")

						CreateSpace := exec.Command("cf", "create-space", Orgs.Org[i].Spaces[j].Name, "-o", Orgs.Org[i].Name)

						if _, err := CreateSpace.Output(); err != nil {
							fmt.Println("command: ", CreateSpace)
							fmt.Println("Err: ", CreateSpace.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", CreateSpace)
							fmt.Println(CreateSpace.Stdout)

							fmt.Println("Enabling Space Isolation Segment")

							iso := exec.Command("cf", "enable-org-isolation", Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].IsolationSeg)

							if _, err := iso.Output(); err != nil {
								fmt.Println("command: ", iso)
								fmt.Println("Err: ", iso.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", iso)
								fmt.Println(iso.Stdout)
							}

							isospace := exec.Command("cf", "set-space-isolation-segment", Orgs.Org[i].Spaces[j].Name, Orgs.Org[i].Spaces[j].IsolationSeg)

							if _, err := isospace.Output(); err != nil {
								fmt.Println("command: ", isospace)
								fmt.Println("Err: ", isospace.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", isospace)
								fmt.Println(isospace.Stdout)
							}

							fmt.Println("Creating ASGs")

							if InitClusterConfigVals.ClusterDetails.EnableASG == true {
								fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
								CreateOrUpdateASGs(Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].Name, ASGPath, ostype)
							} else {
								fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
								fmt.Println("ASGs not enabled")
							}
						}
					}
				}
			} else {
				fmt.Println("command: ", guid )
				fmt.Println("Err: ", guid.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Org doesn't exists, Please create Org")
			}
		} else {
			fmt.Println("This is a protected Org")
		}
	}
	return err
}
func CreateOrUpdateQuotas(clustername string, cpath string) error {

	var Quotas Quotalist
	var ProtectedQuota ProtectedList
	var cmd *exec.Cmd


	QuotaYml := cpath+"/C9Cli/"+clustername+"/Quota.yml"
	fileQuotaYml, err := ioutil.ReadFile(QuotaYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileQuotaYml), &Quotas)
	if err != nil {
		panic(err)
	}

	ProtectedQuotasYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedQYml, err := ioutil.ReadFile(ProtectedQuotasYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedQYml), &ProtectedQuota)
	if err != nil {
		panic(err)
	}

	LenQuota := len(Quotas.Quota)
	LenProtectedQuota := len(ProtectedQuota.Quota)

	for i := 0; i < LenQuota; i++ {

		var count, totalcount int
		fmt.Println("Quota: ", Quotas.Quota[i].Name)

		SerLimit := string(Quotas.Quota[i].ServiceInstanceLimit)
		AppLimt  := string(Quotas.Quota[i].AppInstanceLimit)
		MemLimit := Quotas.Quota[i].MemoryLimit

		if string(Quotas.Quota[i].ServiceInstanceLimit) == "" {
			SerLimit = "25"
		}
		if string(Quotas.Quota[i].AppInstanceLimit) == "" {
			SerLimit = "25"
		}
		if Quotas.Quota[i].MemoryLimit == "" {
			MemLimit = "1024M"
		}
		for p := 0; p < LenProtectedQuota; p++ {
			fmt.Println("Protected Quota: ", ProtectedQuota.Quota[p])
			if strings.Trim(ProtectedQuota.Quota[p], "") == strings.Trim(Quotas.Quota[i].Name, "") {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count
		//fmt.Println(totalcount)

		if totalcount == 0 {

			fmt.Println("This is not Protected Quota")
			Quotadetails := exec.Command("cf", "quota", Quotas.Quota[i].Name)

			if _, err := Quotadetails.Output(); err != nil{
				fmt.Println("command: ", Quotadetails)
				fmt.Println("Err: ", Quotadetails.Stdout)
				fmt.Println("Err Code: ", err)
				//fmt.Println("Quota Doesn't exits: ", Quotadetails.Stdout)
				fmt.Println("Creating Quota")

				if Quotas.Quota[i].AllowPaidPlans == true {
					cmd = exec.Command("cf", "create-quota", Quotas.Quota[i].Name, "-m", MemLimit, "-i", "-1", "-r", "-1", "-s", SerLimit, "-a", AppLimt, "--allow-paid-service-plans")
				} else {
					cmd = exec.Command("cf", "create-quota", Quotas.Quota[i].Name, "-m", MemLimit, "-i", "-1", "-r", "-1", "-s", SerLimit, "-a", AppLimt, "--disallow-paid-service-plans")
				}

				if _, err := cmd.Output(); err != nil{
					fmt.Println("command: ", cmd)
					fmt.Println("Err: ", cmd.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", cmd)
					fmt.Println(cmd.Stdout)
				}
				QuotaGet := exec.Command("cf", "quota", Quotas.Quota[i].Name)
				if _, err := QuotaGet.Output(); err != nil{
					fmt.Println("command: ", QuotaGet)
					fmt.Println("Err: ", QuotaGet.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", QuotaGet)
					fmt.Println(QuotaGet.Stdout)
				}
			} else {
				fmt.Println("command: ", Quotadetails)
				fmt.Println("Quota exists: ", Quotadetails.Stdout)
				fmt.Println("Updating Org")

				if Quotas.Quota[i].AllowPaidPlans == true {
					cmd = exec.Command("cf", "update-quota", Quotas.Quota[i].Name, "-m", MemLimit, "-i", "-1", "-r", "-1", "-s",  SerLimit, "-a", AppLimt, "--allow-paid-service-plans")
				} else {
					cmd = exec.Command("cf", "update-quota", Quotas.Quota[i].Name, "-m", MemLimit, "-i", "-1", "-r", "-1", "-s",  SerLimit, "-a", AppLimt, "--disallow-paid-service-plans")
				}

				if _, err := cmd.Output(); err != nil{
					fmt.Println("command: ", cmd)
					fmt.Println("Err: ", cmd.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", cmd)
					fmt.Println(cmd.Stdout)
				}
				QuotaGet := exec.Command("cf", "quota", Quotas.Quota[i].Name)
				if _, err := QuotaGet.Output(); err != nil{
					fmt.Println("command: ", QuotaGet)
					fmt.Println("Err: ", QuotaGet.Stdout)
					fmt.Println("Err Code: ", err)
				} else {
					fmt.Println("command: ", QuotaGet)
					fmt.Println(QuotaGet.Stdout)
				}
			}
		} else {
			fmt.Println("This is a protected Org")
		}
	}
	return err
}
func CreateOrUpdateOrgUsers(clustername string, cpath string) error {

	var Orgs Orglist
	var ProtectedOrgs ProtectedList

	OrgsYml := cpath+"/C9Cli/"+clustername+"/Org.yml"
	fileOrgYml, err := ioutil.ReadFile(OrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}


	LenProtectedOrgs := len(ProtectedOrgs.Org)
	LenOrgs := len(Orgs.Org)

	for i := 0; i < LenOrgs; i++ {

		var count, totalcount int
		fmt.Println("Org: ", Orgs.Org[i].Name)
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == Orgs.Org[i].Name {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count
		//fmt.Println(totalcount)

		if totalcount == 0 {

			guid := exec.Command("cf", "org", Orgs.Org[i].Name, "--guid")
			if _, err := guid.Output(); err == nil{

				//fmt.Println("Err: ", err)

				fmt.Println("command: ", guid)
				fmt.Println("Org exists: ", guid.Stdout)
				fmt.Println("Updating Org Users")
				fmt.Println("Updating LDAP Users")

				LDAPOrgLen := len(Orgs.Org[i].OrgUsers.LDAP)

				for j := 0; j < LDAPOrgLen; j++ {

					//cf set-org-role USER ORGNAME ROLE(like OrgManager)

					cmd := exec.Command("cf", "set-org-role", Orgs.Org[i].OrgUsers.LDAP[j].Name, Orgs.Org[i].Name, Orgs.Org[i].OrgUsers.LDAP[j].Role)

					if _, err := cmd.Output(); err != nil{
						fmt.Println("command: ", cmd)
						fmt.Println("Err: ", cmd.Stdout)
						fmt.Println("Err Code: ", err)
					} else {
						fmt.Println("command: ", cmd)
						fmt.Println(cmd.Stdout)
					}
				}

				fmt.Println("Updating UAA Users")

				UAAOrgLen := len(Orgs.Org[i].OrgUsers.UAA)

				for k := 0; k < UAAOrgLen; k++ {

					//cf set-org-role USER ORGNAME ROLE(like OrgManager)

					cmd := exec.Command("cf", "set-org-role", Orgs.Org[i].OrgUsers.UAA[k].Name, Orgs.Org[i].Name, Orgs.Org[i].OrgUsers.UAA[k].Role)
					if _, err := cmd.Output(); err != nil{
						fmt.Println("command: ", cmd)
						fmt.Println("Err: ", cmd.Stdout)
						fmt.Println("Err Code: ", err)
					} else {
						fmt.Println("command: ", cmd)
						fmt.Println(cmd.Stdout)
					}
				}

				fmt.Println("Updating SSO Users")

				SSOOrgLen := len(Orgs.Org[i].OrgUsers.SSO)

				for l := 0; l < SSOOrgLen; l++ {
					//cf set-org-role USER ORGNAME ROLE(like OrgManager)
					cmd := exec.Command("cf", "set-org-role", Orgs.Org[i].OrgUsers.SSO[l].Name, Orgs.Org[i].Name, Orgs.Org[i].OrgUsers.SSO[l].Role)
					if _, err := cmd.Output(); err != nil{
						fmt.Println("command: ", cmd)
						fmt.Println("Err: ", cmd.Stdout)
						fmt.Println("Err Code: ", err)
					} else {
						fmt.Println("command: ", cmd)
						fmt.Println(cmd.Stdout)
					}
				}
			} else {
				fmt.Println("command: ", guid)
				fmt.Println("Err: ", guid.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Pulling Org Guid Id: ", guid.Stdout)
				fmt.Println("Please Create Org")
			}
		} else {
			fmt.Println("This is a protected Org")
		}
	}
	return err
}
func CreateOrUpdateSpaceUsers(clustername string, cpath string) error {

	var Orgs Orglist
	var ProtectedOrgs ProtectedList

	OrgsYml := cpath+"/C9Cli/"+clustername+"/Org.yml"
	fileOrgYml, err := ioutil.ReadFile(OrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenProtectedOrgs := len(ProtectedOrgs.Org)
	LenOrgs := len(Orgs.Org)

	for i := 0; i < LenOrgs; i++ {

		var count, totalcount int
		fmt.Println("Org: ", Orgs.Org[i].Name)
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == Orgs.Org[i].Name {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count
		//fmt.Println(totalcount)

		if totalcount == 0 {
			guid := exec.Command("cf", "org", Orgs.Org[i].Name, "--guid")

			if _, err := guid.Output(); err == nil {

				fmt.Println("command: ", guid)
				fmt.Println("Org exists: ", guid.Stdout)
				targetOrg := exec.Command("cf", "t", "-o", Orgs.Org[i].Name)
				if _, err := targetOrg.Output(); err == nil {
					fmt.Println("command: ", targetOrg)
					fmt.Println("Targeted Org: ", targetOrg.Stdout)
				} else {
					fmt.Println("command: ", targetOrg)
					fmt.Println("Err: ", targetOrg.Stdout)
					fmt.Println("Err Code: ", targetOrg.Stderr)
				}
				SpaceLen := len(Orgs.Org[i].Spaces)

				for j := 0; j < SpaceLen; j++ {

					guid = exec.Command("cf", "space", Orgs.Org[i].Spaces[j].Name, "--guid")
					if _, err := guid.Output(); err == nil {
						fmt.Println("command: ", guid)
						fmt.Println("Space exists: ", guid.Stdout)
						fmt.Println("Creating Space users")

						fmt.Println("Updating LDAP Users")

						LDAPSpaceUsersLen := len(Orgs.Org[i].Spaces[j].SpaceUsers.LDAP)

						for k := 0; k < LDAPSpaceUsersLen; k++ {
							cmd := exec.Command("cf", "set-space-role", Orgs.Org[i].Spaces[j].SpaceUsers.LDAP[k].Name, Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].Name, Orgs.Org[i].Spaces[j].SpaceUsers.LDAP[k].Role)
							if _, err := cmd.Output(); err != nil{
								fmt.Println("command: ", cmd)
								fmt.Println("Err: ", cmd.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", cmd)
								fmt.Println(cmd.Stdout)
							}
						}

						fmt.Println("Updating UAA Users")

						UAASpaceUsersLen := len(Orgs.Org[i].Spaces[j].SpaceUsers.UAA)

						for l := 0; l < UAASpaceUsersLen; l++ {
							cmd := exec.Command("cf", "set-space-role", Orgs.Org[i].Spaces[j].SpaceUsers.UAA[l].Name, Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].Name, Orgs.Org[i].Spaces[j].SpaceUsers.UAA[l].Role)
							if _, err := cmd.Output(); err != nil{
								fmt.Println("command: ", cmd)
								fmt.Println("Err: ", cmd.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", cmd)
								fmt.Println(cmd.Stdout)
							}
						}

						fmt.Println("Updating SSO Users")

						SSOSpaceUsersLen := len(Orgs.Org[i].Spaces[j].SpaceUsers.SSO)
						for m := 0; m < SSOSpaceUsersLen; m++ {
							cmd := exec.Command("cf", "set-space-role", Orgs.Org[i].Spaces[j].SpaceUsers.SSO[m].Name, Orgs.Org[i].Name, Orgs.Org[i].Spaces[j].Name, Orgs.Org[i].Spaces[j].SpaceUsers.SSO[m].Role)
							if _, err := cmd.Output(); err != nil{
								fmt.Println("command: ", cmd)
								fmt.Println("Err: ", cmd.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", cmd)
								fmt.Println(cmd.Stdout)
							}
						}

					} else {
						fmt.Println("command: ",guid)
						fmt.Println("Err: ", guid.Stdout)
						fmt.Println("Err Code: ", err)
						fmt.Println("Space doesn't exists, Please create Space")
					}
				}
			} else {
				fmt.Println("command: ", guid)
				fmt.Println("Err: ", guid.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("Org doesn't exists, Please create Org")
			}
		}
	}
	return err
}
func CreateOrUpdateASGs(Org string, Space string, asgpath string, ostype string) {

	ASGPath := asgpath
	ASGName := Org+"_"+Space+".json"
	path := ASGPath+ASGName

	//check := exec.Command("powershell", "-command","Get-Content", path)

	var check *exec.Cmd

	if ostype == "windows" {
		check = exec.Command("powershell", "-command","Get-Content", path)
		//check = exec.Command("type", path)
	} else {
		check = exec.Command("cat", path)
	}

	//check := exec.Command("cat", path)

	if _, err := check.Output(); err != nil {
		fmt.Println("command: ", check)
		fmt.Println("Err: ", check.Stdout)
		fmt.Println("Err Code: ", err)
		fmt.Println("No ASG defined for Org and Space combination")
	} else
	{
		fmt.Println("command: ", check)
		fmt.Println(check.Stdout)
		fmt.Println("Binding ASGs")

		checkcreate := exec.Command("cf", "security-group", ASGName)
		if _, err := checkcreate.Output(); err != nil {
			fmt.Println("command: ", checkcreate)
			fmt.Println("Err: ", checkcreate.Stdout)
			fmt.Println("Err Code: ", err)
			fmt.Println("ASG doesn't exist, Creating ASG")

			createasg := exec.Command("cf", "create-security-group", ASGName, path)
			if _, err := createasg.Output(); err != nil {
				fmt.Println("command: ", createasg)
				fmt.Println("Err: ", createasg.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("ASG creation failed")
			} else {
				fmt.Println("command: ", createasg)
				fmt.Println(createasg.Stdout)
			}
		} else {
			fmt.Println("command: ", checkcreate)
			fmt.Println(checkcreate.Stdout)
			fmt.Println("ASG exist, Updating ASG")
			updateasg := exec.Command("cf", "update-security-group", ASGName, path)
			if _, err := updateasg.Output(); err != nil {
				fmt.Println("command: ", updateasg)
				fmt.Println("Err: ", updateasg.Stdout)
				fmt.Println("Err Code: ", err)
				fmt.Println("ASG update failed")
			} else {
				fmt.Println("command: ", updateasg)
				fmt.Println(updateasg.Stdout)
			}
		}
		fmt.Println("Creating or Updating ASG finished, binding ASG")
		bindasg := exec.Command("cf", "bind-security-group", ASGName, Org, Space, "--lifecycle", "running")
		if _, err := bindasg.Output(); err != nil {
			fmt.Println("command: ", bindasg)
			fmt.Println("Err: ", bindasg.Stdout)
			fmt.Println("Err Code: ", err)
			fmt.Println("ASG binding failed")
		} else {
			fmt.Println("command: ", bindasg)
			fmt.Println(bindasg.Stdout)
		}
	}
	return
}
func Init(clustername string, endpoint string, user string, org string, space string, asg string, cpath string) (err error) {

	type ClusterDetails struct {
		EndPoint         string `yaml:"EndPoint"`
		User         string `yaml:"User"`
		Org            string `yaml:"Org"`
		Space string  `yaml:"Space"`
		EnableASG     string `yaml:"EnableASG"`
	}


	mgmtpath := cpath+"/C9Cli/"+clustername
	ASGPath := cpath+"/C9Cli/"+clustername+"/ASGs/"
	OrgsYml := cpath+"/C9Cli/"+clustername+"/Org.yml"
	QuotasYml := cpath+"/C9Cli/"+clustername+"/Quota.yml"
	ProtectedResourcesYml := cpath+"/C9Cli/"+clustername+"/ProtectedResources.yml"


	_, err = os.Stat(mgmtpath)
	if os.IsNotExist(err) {

		fmt.Println("Creating C9Cli/<cluster> folder")
		errDir := os.MkdirAll(mgmtpath, 0755)


		var data = `---
ClusterDetails:
  EndPoint: {{ .EndPoint }}
  User: {{ .User }}
  Org: {{ .Org }}
  Space: {{ .Space }}
  EnableASG: {{ .EnableASG }}`

		// Create the file:
		err = ioutil.WriteFile(mgmtpath+"/config.tmpl", []byte(data), 0644)
		check(err)

		values := ClusterDetails{EndPoint: endpoint, User: user, Org: org, Space: space, EnableASG: asg}

		var templates *template.Template
		var allFiles []string

		if err != nil {
			fmt.Println(err)
		}

		filename := "config.tmpl"
		fullPath := mgmtpath + "/config.tmpl"
		if strings.HasSuffix(filename, ".tmpl") {
			allFiles = append(allFiles, fullPath)
		}

		fmt.Println(allFiles)
		templates, err = template.ParseFiles(allFiles...)
		if err != nil {
			fmt.Println(err)
		}

		s1 := templates.Lookup("config.tmpl")
		f, err := os.Create(mgmtpath + "/config.yml")
		if err != nil {
			panic(err)
		}

		fmt.Println("Creating C9Cli folder and config files")

		err = s1.Execute(f, values)
		defer f.Close() // don't forget to close the file when finished.
		if err != nil {
			panic(err)
		}



		var OrgTmp = `---
Org:
  - Name: "org-test"
    Quota: "default"
    OrgUsers:
      LDAP:
        - Name: User1
          Role: OrgManager
        - Name: User2
          Role: OrgManager
        - Name: User3
          Role: OrgAudidor
      SSO:
        - Name: User1
          Role: OrgManager
        - Name: User2
          Role: OrgManager
        - Name: User3
          Role: OrgAudidor
      UAA:
        - Name: User1
          Role: OrgManager
        - Name: User2
          Role: OrgManager
        - Name: User3
          Role: OrgAudidor
    Spaces:
      - Name: Space1
        IsolationSeg: "test-segment-1"
        SpaceUsers:
          LDAP:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor
          SSO:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor
          UAA:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor
      - Name: Space2
		IsolationSeg: "test-segment-2"
        SpaceUsers:
          LDAP:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor
          SSO:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor
          UAA:
            - Name: User1
              Role: SpaceManager
            - Name: User2
              Role: SpaceManager
            - Name: User3
              Role: SpaceAudidor`

		var QuotasTmp = `---
quota:
  - Name: default
    memory_limit: 1024M
	allow_paid_plans: False
    app_instance_limit: 25
    service_instance_limit: 25
  - Name: small_quota
    memory_limit: 2048M
  - Name: medium_quota
    memory_limit: 2048M
  - Name: large_quota
    memory_limit: 2048M`

		var ProtectedListTmp = `---
  Org:
    - system
    - healthwatch
    - dynatrace
  quota:
    - default
  DefaultRunningSecurityGroup: default_security_group`

		fmt.Println("Creating C9Cli/<cluster>/ sample yaml files")
		err = ioutil.WriteFile(OrgsYml, []byte(OrgTmp), 0644)
		check(err)
		err = ioutil.WriteFile(QuotasYml, []byte(QuotasTmp), 0644)
		check(err)
		err = ioutil.WriteFile(ProtectedResourcesYml, []byte(ProtectedListTmp), 0644)
		check(err)

		if errDir != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("C9Cli/<cluster> exists, please manually edit file to make changes or provide new cluster name")
	}

	_, err = os.Stat(ASGPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(ASGPath, 0755)
		if errDir != nil {
			log.Fatal(err)
			fmt.Println("C9Cli/<cluster>/ASGs exist, please manually edit file to make changes or provide new cluster name")
		} else {
			fmt.Println("Creating C9Cli/<cluster>/ASGs")
		}
	}

	return
}
func check(e error) {
	if e != nil {
		fmt.Println("C9Cli/mgmt/<cluster>/ yamls exists, please manually edit file to make changes or provide new cluster name")
		panic(e)
	}
}
