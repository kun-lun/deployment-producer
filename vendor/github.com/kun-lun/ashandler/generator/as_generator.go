package generator

import (
	"io/ioutil"
	"path"

	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	clogger "github.com/kun-lun/common/logger"
	"github.com/kun-lun/common/storage"
	yaml "gopkg.in/yaml.v2"
)

type ASGenerator struct {
	stateStore storage.Store
	logger     *clogger.Logger
}

func NewASGenerator(
	stateStore storage.Store,
	logger *clogger.Logger,
) ASGenerator {
	return ASGenerator{
		stateStore: stateStore,
		logger:     logger,
	}
}

// https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html?highlight=roles
func (a ASGenerator) Generate(hostGroups []deployments.HostGroup, deployments []deployments.Deployment) error {
	// generate the hosts files.
	hostsFileContent := a.generateHostsFile(hostGroups)
	ansibleInventoriesDir, _ := a.stateStore.GetAnsibleInventoriesDir()
	hostsFile := path.Join(ansibleInventoriesDir, "hosts.yml")
	a.logger.Printf("writting hosts file to %s\n", hostsFile)
	err := ioutil.WriteFile(hostsFile, hostsFileContent, 0644)
	if err != nil {
		a.logger.Printf("write file failed: %s\n", err.Error())
		return err
	}

	// generate the roles files.
	playbookContent := a.generatePlaybookFile(deployments)
	ansibleDir, _ := a.stateStore.GetAnsibleDir()
	playbookFile := path.Join(ansibleDir, "kunlun.yml")

	a.logger.Printf("writting playbook file to %s\n", playbookFile)
	err = ioutil.WriteFile(playbookFile, playbookContent, 0644)
	if err != nil {
		a.logger.Printf("write file failed: %s\n", err.Error())
		return err
	}
	return nil
}

// TODO error handling.
func (a ASGenerator) generateHostsFile(hostGroups []deployments.HostGroup) []byte {
	// ---
	// sample_server:
	// 	 hosts:
	// 	   172.16.8.4:
	// 	     ansible_ssh_user: andy
	// 	     ansible_ssh_common_args: '-o ProxyCommand="ssh -W %h:%p -q andy@65.52.176.243"'
	hostGroupsSlices := yaml.MapSlice{}

	for _, hostGroup := range hostGroups {
		hosts := yaml.MapSlice{}

		for _, host := range hostGroup.Hosts {
			hostSlice := yaml.MapItem{
				Key: host.Alias,
				Value: AnsibleHost{
					Host:          host.Host,
					SSHUser:       host.User,
					SSHCommonArgs: host.SSHCommonArgs,
				},
			}
			hosts = append(hosts, hostSlice)
		}

		hostGroupSlice := yaml.MapItem{
			Key: hostGroup.Name,
			Value: yaml.MapSlice{
				{
					Key:   "hosts",
					Value: hosts,
				},
			},
		}
		hostGroupsSlices = append(hostGroupsSlices, hostGroupSlice)
	}
	content, _ := yaml.Marshal(hostGroupsSlices)
	return content
}

type AnsibleHost struct {
	Host          string `yaml:"ansible_host"`
	SSHUser       string `yaml:"ansible_ssh_user"`
	SSHCommonArgs string `yaml:"ansible_ssh_common_args"`
}

type role struct {
	Role       string `yaml:"role"`
	Become     string `yaml:"become"`
	BecomeUser string `yaml:"become_user"`
}
type depItem struct {
	Hosts    string   `yaml:"hosts"`
	VarsFile []string `yaml:"var_files"`
	Roles    []role   `yaml:"roles"`
}

// TODO error handling.
func (a ASGenerator) generatePlaybookFile(deployments []deployments.Deployment) []byte {
	// ---
	// - hosts: sample_server
	//   vars_files:
	// 	   - vars/sample.yml
	//   roles:
	// 	   - role: 'geerlingguy.composer'
	// 	     become: true
	// 	     become_user: root
	// - hosts: sample_server2
	//   vars_files:
	// 	   - vars/sample.yml
	//   roles:
	// 	   - role: 'geerlingguy.php'
	// 	     become: true
	// 	     become_user: root

	depItems := []depItem{}
	// write the vars files
	for _, dep := range deployments {
		// write the files
		varsDir, _ := a.stateStore.GetAnsibleVarsDir()
		varsFile := path.Join(varsDir, dep.HostGroupName+".yml")
		varsContent, _ := yaml.Marshal(dep.Vars)

		a.logger.Printf("writting vars file to %s\n", varsFile)
		err := ioutil.WriteFile(varsFile, varsContent, 0644)
		if err != nil {
			a.logger.Printf("write vars file failed: %s\n", err.Error())
		}
		depItem := depItem{
			Hosts:    dep.HostGroupName,
			VarsFile: []string{varsFile},
		}
		depItems = append(depItems, depItem)
	}
	content, _ := yaml.Marshal(depItems)
	return content
}
