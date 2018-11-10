package generator

import (
	"io/ioutil"
	"path"

	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	builtinroles "github.com/kun-lun/built-in-roles/pkg/apis"
	"github.com/kun-lun/common/fileio"
	"github.com/kun-lun/common/storage"
	yaml "gopkg.in/yaml.v2"
)

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string) bool
}

type ASGenerator struct {
	stateStore storage.Store
	logger     logger
	fs         fileio.Fs
}

func NewASGenerator(
	stateStore storage.Store,
	logger logger,
	fs fileio.Fs,
) ASGenerator {
	return ASGenerator{
		stateStore: stateStore,
		logger:     logger,
		fs:         fs,
	}
}

// https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html?highlight=roles
func (a ASGenerator) Generate(hostGroups []deployments.HostGroup, deployments []deployments.Deployment) error {
	// generate the ansible config file.
	builtInRolesFS, err := builtinroles.FSByte(false, "/ansible.cfg")
	if err != nil {
		return err
	}
	ansibleDir, err := a.stateStore.GetAnsibleDir()
	ansibleConfigFile := path.Join(ansibleDir, "ansible.cfg")
	a.fs.WriteFile(ansibleConfigFile, builtInRolesFS, 0644)

	// generate the hosts files.
	hostsFileContent := a.generateHostsFile(hostGroups)
	ansibleInventoriesDir, _ := a.stateStore.GetAnsibleInventoriesDir()
	hostsFile := path.Join(ansibleInventoriesDir, "hosts.yml")
	a.logger.Printf("writting hosts file to %s\n", hostsFile)
	err = a.fs.WriteFile(hostsFile, hostsFileContent, 0644)
	if err != nil {
		a.logger.Printf("write file failed: %s\n", err.Error())
		return err
	}

	// generate the roles files.
	playbookContent := a.generatePlaybookFile(deployments)
	ansibleMainFile, err := a.stateStore.GetAnsibleMainFile()

	a.logger.Printf("writting playbook file to %s\n", ansibleMainFile)
	err = ioutil.WriteFile(ansibleMainFile, playbookContent, 0644)
	if err != nil {
		a.logger.Printf("write file failed: %s\n", err.Error())
		return err
	}

	// generate the deployment script file.
	deploymentScriptFilePath, err := a.stateStore.GetDeploymentScriptFile()
	deploymentScriptContent := a.generateDeploymentScript()
	err = a.fs.WriteFile(deploymentScriptFilePath, deploymentScriptContent, 0744)
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
	VarsFile []string `yaml:"vars_files"`
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
		varsDir, _ := a.stateStore.GetAnsibleDir()
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

func (a ASGenerator) generateDeploymentScript() []byte {
	// varsDir, _ := a.stateStore.GetAnsibleDir()
	deploymentScript := (`#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export ANSIBLE_CONFIG=$DIR/ansible/ansible.cfg
ansible-playbook -i $DIR/ansible/inventories $DIR/ansible/main.yml -vv
`)
	return []byte(deploymentScript)
}
