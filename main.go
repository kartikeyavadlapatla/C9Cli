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

type List struct {
	OrgList []string `yaml:"OrgList"`
}
type Orglist struct {
	Org struct {
		Name     string `yaml:"Name"`
		Quota    string `yaml:"Quota"`
		OrgUsers struct {
			LDAP struct {
				OrgManagers []string `yaml:"OrgManagers"`
				OrgAuditors []string `yaml:"OrgAuditors"`
			} `yaml:"LDAP"`
			SSO struct {
				OrgManagers []string `yaml:"OrgManagers"`
				OrgAuditors []string `yaml:"OrgAuditors"`
			} `yaml:"SSO"`
			UAA struct {
				OrgManagers []string `yaml:"OrgManagers"`
				OrgAuditors []string `yaml:"OrgAuditors"`
			} `yaml:"UAA"`
		} `yaml:"OrgUsers"`
		Spaces []struct {
			Name         string `yaml:"Name"`
			IsolationSeg string `yaml:"IsolationSeg"`
			SpaceUsers struct {
				LDAP struct {
					SpaceManagers   []string `yaml:"SpaceManagers"`
					SpaceDevelopers []string `yaml:"SpaceDevelopers"`
					SpaceAuditors   []string `yaml:"SpaceAuditors"`
				} `yaml:"LDAP"`
				UAA struct {
					SpaceManagers   []string `yaml:"SpaceManagers"`
					SpaceDevelopers []string `yaml:"SpaceDevelopers"`
					SpaceAuditors   []string `yaml:"SpaceAuditors"`
				} `yaml:"UAA"`
				SSO struct {
					SpaceManagers   []string `yaml:"SpaceManagers"`
					SpaceDevelopers []string `yaml:"SpaceDevelopers"`
					SpaceAuditors   []string `yaml:"SpaceAuditors"`
				} `yaml:"SSO"`
			} `yaml:"SpaceUsers,omitempty"`
		} `yaml:"Spaces"`
	} `yaml:"Org"`
}
type Quotalist struct {
	Quota []struct {
		Name        string `yaml:"Name"`
		MemoryLimit string `yaml:"memory_limit"`
		AllowPaidPlans       bool `yaml:"allow_paid_plans"`
		AppInstanceLimit     string `yaml:"app_instance_limit"`
		ServiceInstanceLimit string `yaml:"service_instance_limit"`
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
	flag.StringVar(&operation, "i", "", "Provide Operation to be performed: init, create-{org,space,org-user,space-user,quota, ")
	flag.StringVar(&cpath, "k", ".", "Provide path to configs, i.e, to config folder, use with all operations")
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
		fmt.Printf("Org: %v\n", org)
		fmt.Printf("Space: %v\n", space)
		fmt.Printf("EnableASG: %v\n", asg)
		fmt.Printf("Path: %v\n", cpath)
		Init(ClusterName, endpoint, user, org, space, asg, cpath)
	} else if operation == "org-init" {

		fmt.Printf("ClusterName: %v\n", ClusterName)
		OrgsInit(ClusterName, cpath)

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

	} else if operation == "create-space-asg"{

		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateSpacesASGs (ClusterName, cpath, ostype)

	}else if operation == "create-all" {
		fmt.Printf("ClusterName: %v\n", ClusterName)
		SetupConnection (ClusterName,  pwd, cpath)
		CreateOrUpdateProtOrgAsg (ClusterName, cpath, ostype)
		CreateOrUpdateQuotas(ClusterName, cpath)
		CreateOrUpdateOrgs (ClusterName, cpath)
		CreateOrUpdateOrgUsers(ClusterName, cpath)
		CreateOrUpdateSpaces (ClusterName, cpath, ostype)
		CreateOrUpdateSpacesASGs (ClusterName, cpath, ostype)
		CreateOrUpdateSpaceUsers (ClusterName, cpath)
	} else {
		fmt.Println("Provide Valid input operation")
	}
}
func CreateOrUpdateProtOrgAsg(clustername string, cpath string, ostype string) {

	var ProtectedOrgs ProtectedList
	var ASGpath string
	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/"+clustername+"/config.yml"

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
		ASGpath = cpath+"\\"+clustername+"\\ProtectedOrgsASGs\\"
	} else {
		ASGpath = cpath+"/"+clustername+"/ProtectedOrgsASGs/"
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
	ConfigFile := cpath+"/"+clustername+"/config.yml"

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
	var list List
	var ProtectedOrgs ProtectedList

	ListYml := cpath+"/"+clustername+"/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenList := len(list.OrgList)
	LenProtectedOrgs := len(ProtectedOrgs.Org)

	for i := 0; i < LenList; i++ {
		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")

			OrgsYml := cpath+"/"+clustername+"/"+list.OrgList[i]+"/Org.yml"
			fileOrgYml, err := ioutil.ReadFile(OrgsYml)

			if err != nil {
				fmt.Println(err)
			}

			err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
			if err != nil {
				panic(err)
			}
			if list.OrgList[i] == Orgs.Org.Name {
				guid := exec.Command("cf", "org", Orgs.Org.Name, "--guid")
				if _, err := guid.Output(); err == nil{

					fmt.Println("command: ", guid)
					fmt.Println("Org exists: ", guid.Stdout)
					fmt.Println("Updating Org quota")
					SetQuota := exec.Command("cf", "set-quota", Orgs.Org.Name, Orgs.Org.Quota)
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
					createorg := exec.Command("cf", "create-org", Orgs.Org.Name)
					if _, err := createorg.Output(); err != nil{
						fmt.Println("command: ", createorg)
						fmt.Println("Err: ", createorg.Stdout)
						fmt.Println("Err Code: ", err)
					} else {
						fmt.Println("command: ", createorg)
						fmt.Println(createorg.Stdout)
					}
					attachquota := exec.Command("cf", "set-quota", Orgs.Org.Name, Orgs.Org.Quota)
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
				fmt.Println("Org Name does't match with folder name")
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
	var list List

	ListYml := cpath+"/"+clustername+"/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/"+clustername+"/config.yml"

	fileConfigYml, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileConfigYml), &InitClusterConfigVals)
	if err != nil {
		panic(err)
	}
	var OrgsYml string
	//var ASGPath, OrgsYml string

	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenList := len(list.OrgList)
	LenProtectedOrgs := len(ProtectedOrgs.Org)


	for i := 0; i < LenList; i++ {

		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")

			if ostype == "windows" {
				//ASGPath = cpath+"\\"+clustername+"\\"+list.OrgList[i]+"\\ASGs\\"
				OrgsYml = cpath+"\\"+clustername+"\\"+list.OrgList[i]+"\\Org.yml"
			} else {
				//ASGPath = cpath+"/"+clustername+"/"+list.OrgList[i]+"/ASGs/"
				OrgsYml = cpath+"/"+clustername+"/"+list.OrgList[i]+"/Org.yml"
			}


			fileOrgYml, err := ioutil.ReadFile(OrgsYml)

			if err != nil {
				fmt.Println(err)
			}

			err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
			if err != nil {
				panic(err)
			}
			if list.OrgList[i] == Orgs.Org.Name {
				guid := exec.Command("cf", "org", Orgs.Org.Name, "--guid")

				if _, err := guid.Output(); err == nil {

					fmt.Println("command: ", guid)
					fmt.Println("Org exists: ", guid.Stdout)
					SpaceLen := len(Orgs.Org.Spaces)

					TargetOrg := exec.Command("cf", "t", "-o", Orgs.Org.Name)
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

						guid = exec.Command("cf", "space", Orgs.Org.Spaces[j].Name, "--guid")

						if _, err := guid.Output(); err == nil{

							fmt.Println("command: ", guid)
							fmt.Println("Space exists: ", guid.Stdout)

							fmt.Println("Enabling Space Isolation Segment")
							fmt.Println(Orgs.Org)
							fmt.Println(Orgs.Org.Spaces[j].Name)
							fmt.Println("SegName: ", Orgs.Org.Spaces[j].IsolationSeg)
							if Orgs.Org.Spaces[j].IsolationSeg != "" {
								iso := exec.Command("cf", "enable-org-isolation", Orgs.Org.Name, Orgs.Org.Spaces[j].IsolationSeg)
								if _, err := iso.Output(); err != nil {
									fmt.Println("command: ", iso)
									fmt.Println("Err: ", iso.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", iso)
									fmt.Println(iso.Stdout)
								}
								isospace := exec.Command("cf", "set-space-isolation-segment", Orgs.Org.Spaces[j].Name, Orgs.Org.Spaces[j].IsolationSeg)

								if _, err := isospace.Output(); err != nil {
									fmt.Println("command: ", isospace)
									fmt.Println("Err: ", isospace.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", isospace)
									fmt.Println(isospace.Stdout)
								}
							} else {
								fmt.Println("No Isolation Segment Provided, Will be attached to Default")
							}


							//fmt.Println("Creating or updating ASGs")
							//if InitClusterConfigVals.ClusterDetails.EnableASG == true {
							//	fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
							//	CreateOrUpdateASGs(Orgs.Org.Name, Orgs.Org.Spaces[j].Name, ASGPath, ostype)
							//} else {
							//	fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
							//	fmt.Println("ASGs not enabled")
							//}
						} else {
							fmt.Println("command: ", guid)
							fmt.Println("Pulling Space Guid ID: ", guid.Stdout )
							fmt.Println("Creating Space")

							CreateSpace := exec.Command("cf", "create-space", Orgs.Org.Spaces[j].Name, "-o", Orgs.Org.Name)

							if _, err := CreateSpace.Output(); err != nil {
								fmt.Println("command: ", CreateSpace)
								fmt.Println("Err: ", CreateSpace.Stdout)
								fmt.Println("Err Code: ", err)
							} else {
								fmt.Println("command: ", CreateSpace)
								fmt.Println(CreateSpace.Stdout)

								fmt.Println("Enabling Space Isolation Segment")
								fmt.Println("SegName: ", Orgs.Org.Spaces[j].IsolationSeg)
								if Orgs.Org.Spaces[j].IsolationSeg != "" {
									iso := exec.Command("cf", "enable-org-isolation", Orgs.Org.Name, Orgs.Org.Spaces[j].IsolationSeg)

									if _, err := iso.Output(); err != nil {
										fmt.Println("command: ", iso)
										fmt.Println("Err: ", iso.Stdout)
										fmt.Println("Err Code: ", err)
									} else {
										fmt.Println("command: ", iso)
										fmt.Println(iso.Stdout)
									}

									isospace := exec.Command("cf", "set-space-isolation-segment", Orgs.Org.Spaces[j].Name, Orgs.Org.Spaces[j].IsolationSeg)
									if _, err := isospace.Output(); err != nil {
										fmt.Println("command: ", isospace)
										fmt.Println("Err: ", isospace.Stdout)
										fmt.Println("Err Code: ", err)
									} else {
										fmt.Println("command: ", isospace)
										fmt.Println(isospace.Stdout)
									}

								} else {
									fmt.Println("No Isolation Segment Provided, Will be attached to Default")
								}

				//				fmt.Println("Creating ASGs")
				//				if InitClusterConfigVals.ClusterDetails.EnableASG == true {
				//					fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
				//					CreateOrUpdateASGs(Orgs.Org.Name, Orgs.Org.Spaces[j].Name, ASGPath, ostype)
				//				} else {
				//					fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
				//					fmt.Println("ASGs not enabled")
				//				}
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
				fmt.Println("Org Name does't match with folder name")
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


	QuotaYml := cpath+"/"+clustername+"/Quota.yml"
	fileQuotaYml, err := ioutil.ReadFile(QuotaYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileQuotaYml), &Quotas)
	if err != nil {
		panic(err)
	}

	ProtectedQuotasYml := cpath+"/"+clustername+"/ProtectedResources.yml"
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

		SerLimit := Quotas.Quota[i].ServiceInstanceLimit
		AppLimt  := Quotas.Quota[i].AppInstanceLimit
		MemLimit := Quotas.Quota[i].MemoryLimit

		if Quotas.Quota[i].ServiceInstanceLimit == ""{
			SerLimit = "0"
		}else {
		}

		if string(Quotas.Quota[i].AppInstanceLimit) == "" {
			AppLimt = "25"
		} else {
		}

		if Quotas.Quota[i].MemoryLimit == "" {
			MemLimit = "1024M"
		} else {
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
				fmt.Println("Updating Quota")

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

	var list List
	var Orgs Orglist
	var ProtectedOrgs ProtectedList

	ListYml := cpath+"/"+clustername+"/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}


	LenProtectedOrgs := len(ProtectedOrgs.Org)
	LenList := len(list.OrgList)
	for i := 0; i < LenList; i++ {

		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {

			fmt.Println("This is not Protected Org")

			OrgsYml := cpath+"/"+clustername+"/"+list.OrgList[i]+"/Org.yml"
			fileOrgYml, err := ioutil.ReadFile(OrgsYml)

			if err != nil {
				fmt.Println(err)
			}

			err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
			if err != nil {
				panic(err)
			}

			if list.OrgList[i] == Orgs.Org.Name {
				guid := exec.Command("cf", "org", Orgs.Org.Name, "--guid")
				if _, err := guid.Output(); err == nil{

					fmt.Println("command: ", guid)
					fmt.Println("Org exists: ", guid.Stdout)
					fmt.Println("Updating Org Users")
					fmt.Println("Updating LDAP Users")

					LDAPOrgManLen := len(Orgs.Org.OrgUsers.LDAP.OrgManagers)

					for j := 0; j < LDAPOrgManLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.LDAP.OrgManagers[j], Orgs.Org.Name, "OrgManager")

						if _, err := cmd.Output(); err != nil{
							fmt.Println("command: ", cmd)
							fmt.Println("Err: ", cmd.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", cmd)
							fmt.Println(cmd.Stdout)
						}
					}

					LDAPOrgAudLen := len(Orgs.Org.OrgUsers.LDAP.OrgAuditors)

					for j := 0; j < LDAPOrgAudLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.LDAP.OrgAuditors[j], Orgs.Org.Name, "OrgAuditor")

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

					UAAOrgManLen := len(Orgs.Org.OrgUsers.UAA.OrgManagers)

					for j := 0; j < UAAOrgManLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.UAA.OrgManagers[j], Orgs.Org.Name, "OrgManager")

						if _, err := cmd.Output(); err != nil{
							fmt.Println("command: ", cmd)
							fmt.Println("Err: ", cmd.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", cmd)
							fmt.Println(cmd.Stdout)
						}
					}

					UAAOrgAudLen := len(Orgs.Org.OrgUsers.UAA.OrgAuditors)

					for j := 0; j < UAAOrgAudLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.UAA.OrgAuditors[j], Orgs.Org.Name, "OrgAuditor")

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

					SSOOrgManLen := len(Orgs.Org.OrgUsers.SSO.OrgManagers)

					for j := 0; j < SSOOrgManLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.SSO.OrgManagers[j], Orgs.Org.Name, "OrgManager")

						if _, err := cmd.Output(); err != nil{
							fmt.Println("command: ", cmd)
							fmt.Println("Err: ", cmd.Stdout)
							fmt.Println("Err Code: ", err)
						} else {
							fmt.Println("command: ", cmd)
							fmt.Println(cmd.Stdout)
						}
					}

					SSOOrgAudLen := len(Orgs.Org.OrgUsers.SSO.OrgAuditors)

					for j := 0; j < SSOOrgAudLen; j++ {

						cmd := exec.Command("cf", "set-org-role", Orgs.Org.OrgUsers.SSO.OrgAuditors[j], Orgs.Org.Name, "OrgAuditor")

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
				fmt.Println("Org Name does't match with folder name")
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
	var list List

	ListYml := cpath+"/"+clustername+"/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenProtectedOrgs := len(ProtectedOrgs.Org)
	LenList := len(list.OrgList)

	for i := 0; i < LenList; i++ {

		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {


			OrgsYml := cpath+"/"+clustername+"/"+list.OrgList[i]+"/Org.yml"
			fileOrgYml, err := ioutil.ReadFile(OrgsYml)

			if err != nil {
				fmt.Println(err)
			}

			err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
			if err != nil {
				panic(err)
			}
			if list.OrgList[i] == Orgs.Org.Name {
				guid := exec.Command("cf", "org", Orgs.Org.Name, "--guid")

				if _, err := guid.Output(); err == nil {

					fmt.Println("command: ", guid)
					fmt.Println("Org exists: ", guid.Stdout)
					targetOrg := exec.Command("cf", "t", "-o", Orgs.Org.Name)
					if _, err := targetOrg.Output(); err == nil {
						fmt.Println("command: ", targetOrg)
						fmt.Println("Targeted Org: ", targetOrg.Stdout)
					} else {
						fmt.Println("command: ", targetOrg)
						fmt.Println("Err: ", targetOrg.Stdout)
						fmt.Println("Err Code: ", targetOrg.Stderr)
					}
					SpaceLen := len(Orgs.Org.Spaces)

					for j := 0; j < SpaceLen; j++ {

						guid = exec.Command("cf", "space", Orgs.Org.Spaces[j].Name, "--guid")
						if _, err := guid.Output(); err == nil {
							fmt.Println("command: ", guid)
							fmt.Println("Space exists: ", guid.Stdout)
							fmt.Println("Creating Space users")

							fmt.Println("Updating LDAP Users")

							LDAPSpaceManLen := len(Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceManagers)

							for k := 0; k < LDAPSpaceManLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceManager")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							LDAPSpaceDevLen := len(Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceAuditors)

							for k := 0; k < LDAPSpaceDevLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceDeveloper")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							LDAPSpaceAuditLen := len(Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceDevelopers)

							for k := 0; k < LDAPSpaceAuditLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.LDAP.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceAuditor")
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

							UAASpaceManLen := len(Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceManagers)

							for k := 0; k < UAASpaceManLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceManager")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							UAASpaceDevLen := len(Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceDevelopers)

							for k := 0; k < UAASpaceDevLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceDeveloper")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							UAASpaceAuditLen := len(Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceAuditors)

							for k := 0; k < UAASpaceAuditLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.UAA.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceAuditor")
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

							SSOSpaceManLen := len(Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceManagers)

							for k := 0; k < SSOSpaceManLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceManager")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							SSOSpaceDevLen := len(Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceDevelopers)

							for k := 0; k < SSOSpaceDevLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceDeveloper")
								if _, err := cmd.Output(); err != nil{
									fmt.Println("command: ", cmd)
									fmt.Println("Err: ", cmd.Stdout)
									fmt.Println("Err Code: ", err)
								} else {
									fmt.Println("command: ", cmd)
									fmt.Println(cmd.Stdout)
								}
							}

							SSOSpaceAuditLen := len(Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceAuditors)

							for k := 0; k < SSOSpaceAuditLen; k++ {
								cmd := exec.Command("cf", "set-space-role", Orgs.Org.Spaces[j].SpaceUsers.SSO.SpaceManagers[k], Orgs.Org.Name, Orgs.Org.Spaces[j].Name, "SpaceAuditor")
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
			} else {
				fmt.Println("Org Name does't match with folder name")
			}
		}
	}
	return err
}
func CreateOrUpdateSpacesASGs(clustername string, cpath string, ostype string) error {

	var Orgs Orglist
	var ProtectedOrgs ProtectedList
	var list List

	ListYml := cpath+"/"+clustername+"/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	var InitClusterConfigVals InitClusterConfigVals
	ConfigFile := cpath+"/"+clustername+"/config.yml"

	fileConfigYml, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileConfigYml), &InitClusterConfigVals)
	if err != nil {
		panic(err)
	}
	var ASGPath, OrgsYml string

	ProtectedOrgsYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenList := len(list.OrgList)
	LenProtectedOrgs := len(ProtectedOrgs.Org)


	for i := 0; i < LenList; i++ {

		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")

			if ostype == "windows" {
				ASGPath = cpath+"\\"+clustername+"\\"+list.OrgList[i]+"\\ASGs\\"
				OrgsYml = cpath+"\\"+clustername+"\\"+list.OrgList[i]+"\\Org.yml"
			} else {
				ASGPath = cpath+"/"+clustername+"/"+list.OrgList[i]+"/ASGs/"
				OrgsYml = cpath+"/"+clustername+"/"+list.OrgList[i]+"/Org.yml"
			}


			fileOrgYml, err := ioutil.ReadFile(OrgsYml)

			if err != nil {
				fmt.Println(err)
			}

			err = yaml.Unmarshal([]byte(fileOrgYml), &Orgs)
			if err != nil {
				panic(err)
			}
			if list.OrgList[i] == Orgs.Org.Name {
				guid := exec.Command("cf", "org", Orgs.Org.Name, "--guid")

				if _, err := guid.Output(); err == nil {

					fmt.Println("command: ", guid)
					fmt.Println("Org exists: ", guid.Stdout)
					SpaceLen := len(Orgs.Org.Spaces)

					TargetOrg := exec.Command("cf", "t", "-o", Orgs.Org.Name)
					if _, err := TargetOrg.Output(); err == nil {
						fmt.Println("command: ", TargetOrg)
						fmt.Println("Targeting: ", TargetOrg.Stdout)
					} else {
						fmt.Println("command: ", TargetOrg)
						fmt.Println("Err: ", TargetOrg.Stdout)
						fmt.Println("Err Code: ", err)
					}

					for j := 0; j < SpaceLen; j++ {

						fmt.Println("Creating Spaces ASGs")
						guid = exec.Command("cf", "space", Orgs.Org.Spaces[j].Name, "--guid")

						if _, err := guid.Output(); err == nil{

							fmt.Println("command: ", guid)
							fmt.Println("Space exists: ", guid.Stdout)
							fmt.Println("Creating or updating ASGs")
							if InitClusterConfigVals.ClusterDetails.EnableASG == true {
								fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
								CreateOrUpdateASGs(Orgs.Org.Name, Orgs.Org.Spaces[j].Name, ASGPath, ostype)
							} else {
								fmt.Println("Enable ASGs: ", InitClusterConfigVals.ClusterDetails.EnableASG)
								fmt.Println("ASGs not enabled")
							}
						} else {
							fmt.Println("command: ", guid)
							fmt.Println("Pulling Space Guid ID: ", guid.Stdout )
							fmt.Println("Space doesn't exist, please create space")
						}
					}
				} else {
					fmt.Println("command: ", guid )
					fmt.Println("Err: ", guid.Stdout)
					fmt.Println("Err Code: ", err)
					fmt.Println("Org doesn't exists, Please create Org")
				}
			} else {
				fmt.Println("Org Name does't match with folder name")
			}
		} else {
			fmt.Println("This is a protected Org")
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
func OrgsInit(clustername string, cpath string) error {

	var list List
	var ProtectedOrgs ProtectedList

	ListYml := cpath + "/" + clustername + "/OrgsList.yml"
	fileOrgYml, err := ioutil.ReadFile(ListYml)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileOrgYml), &list)
	if err != nil {
		panic(err)
	}

	ProtectedOrgsYml := cpath + "/" + clustername + "/ProtectedResources.yml"
	fileProtectedYml, err := ioutil.ReadFile(ProtectedOrgsYml)

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal([]byte(fileProtectedYml), &ProtectedOrgs)
	if err != nil {
		panic(err)
	}

	LenList := len(list.OrgList)
	LenProtectedOrgs := len(ProtectedOrgs.Org)

	for i := 0; i < LenList; i++ {
		var count, totalcount int
		fmt.Println("Org: ", list.OrgList[i])
		for p := 0; p < LenProtectedOrgs; p++ {
			fmt.Println("Protected Org: ", ProtectedOrgs.Org[p])
			if ProtectedOrgs.Org[p] == list.OrgList[i] {
				count = 1
			} else {
				count = 0
			}
		}
		totalcount = totalcount + count

		if totalcount == 0 {
			fmt.Println("This is not Protected Org")

			mgmtpath := cpath + "/" + clustername + "/" + list.OrgList[i]
			ASGPath := cpath + "/" + clustername + "/" + list.OrgList[i] + "/ASGs/"
			OrgsYml := cpath + "/" + clustername + "/" + list.OrgList[i] +"/Org.yml"
			JsonPath := cpath + "/" + clustername + "/" + list.OrgList[i] + "/ASGs/" + "test_test.json"

			_, err = os.Stat(mgmtpath)
			if os.IsNotExist(err) {

				fmt.Println("Creating <cluster>/<Org> folder")
				errDir := os.MkdirAll(mgmtpath, 0755)
				if errDir != nil {
					log.Fatal(err)
				}

				var OrgTmp = `---
Org:
  Name: ""
  Quota: ""
  OrgUsers:
    LDAP:
      OrgManagers:
        - User1
        - User2
        - User3
      OrgAuditors:
        - User1
        - User2
    SSO:
      OrgManagers:
        - User1
        - User2
        - User3
      OrgAuditors:
        - User1
        - User2
    UAA:
      OrgManagers:
        - User1
        - User2
        - User3
      OrgAuditors:
        - User1
        - User2
  Spaces:
    - Name: Space1
      IsolationSeg: "test-segment-1"
      SpaceUsers:
        LDAP:
          SpaceManagers:
            - User1
            - User2
            - User3
          SpaceDevelopers:
            - User1
            - User2
            - User3
          SpaceAuditors:
            - User1
            - User2
            - User3
    - Name: Space2
      IsolationSeg: "test-segment-2"
      SpaceUsers:
        LDAP:
          SpaceManagers:
            - User1
            - User2
            - User3
          SpaceDevelopers:
            - User1
            - User2
            - User3
          SpaceAuditors:
            - User1
            - User2
            - User3`

				fmt.Println("Creating <cluster>/<Org> sample yaml files")
				err = ioutil.WriteFile(OrgsYml, []byte(OrgTmp), 0644)
				check(err)
			} else {
				fmt.Println("<cluster>/<Org> exists, please manually edit file to make changes or provide new cluster name")
			}
			_, err = os.Stat(ASGPath)
			if os.IsNotExist(err) {
				errDir := os.MkdirAll(ASGPath, 0755)
				if errDir != nil {
					log.Fatal(err)
					fmt.Println("<cluster>/<Org>/ASGs exist, please manually edit file to make changes or provide new cluster name")
				} else {
					fmt.Println("Creating <cluster>/<Org>/ASGs")
					var AsgTmp = `---
[
  {
    "protocol": "tcp",
    "destination": "10.x.x.88",
    "ports": "1443",
	"log": true,
	"description": "Allow DNS lookup by default."
  }
]`

					fmt.Println("Creating <cluster>/<Org>/ASGs sample json file")
					err = ioutil.WriteFile(JsonPath, []byte(AsgTmp), 0644)
					check(err)
				}
			}
		}
	}
	return nil
}
func Init(clustername string, endpoint string, user string, org string, space string, asg string, cpath string) (err error) {

	type ClusterDetails struct {
		EndPoint         string `yaml:"EndPoint"`
		User         string `yaml:"User"`
		Org            string `yaml:"Org"`
		Space string  `yaml:"Space"`
		EnableASG     string `yaml:"EnableASG"`
	}


	mgmtpath := cpath+"/"+clustername
	ASGPath := cpath+"/"+clustername+"/ProtectedOrgsASGs/"
	QuotasYml := cpath+"/"+clustername+"/Quota.yml"
	ProtectedResourcesYml := cpath+"/"+clustername+"/ProtectedResources.yml"
	ListOrgsYml := cpath+"/"+clustername+"/OrgsList.yml"


	_, err = os.Stat(mgmtpath)
	if os.IsNotExist(err) {

		fmt.Println("Creating <cluster> folder")
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

		fmt.Println("Initializing folder and config files")

		err = s1.Execute(f, values)
		defer f.Close() // don't forget to close the file when finished.
		if err != nil {
			panic(err)
		}

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

		var ListTmp = `---
OrgList:
  - Org-1
  - Org-2
  - Org-3`

		fmt.Println("Creating <cluster>/ sample yaml files")
		err = ioutil.WriteFile(QuotasYml, []byte(QuotasTmp), 0644)
		check(err)
		err = ioutil.WriteFile(ProtectedResourcesYml, []byte(ProtectedListTmp), 0644)
		check(err)
		err = ioutil.WriteFile(ListOrgsYml, []byte(ListTmp), 0644)
		check(err)

		if errDir != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("<cluster> exists, please manually edit file to make changes or provide new cluster name")
	}

	_, err = os.Stat(ASGPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(ASGPath, 0755)
		if errDir != nil {
			log.Fatal(err)
			fmt.Println("<cluster>/ASGs exist, please manually edit file to make changes or provide new cluster name")
		} else {
			fmt.Println("Creating <cluster>/ASGs")
		}
	}

	return
}
func check(e error) {
	if e != nil {
		fmt.Println("<cluster>/ yamls exists, please manually edit file to make changes or provide new cluster name")
		panic(e)
	}
}
