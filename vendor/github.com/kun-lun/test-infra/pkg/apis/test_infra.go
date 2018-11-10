package apis

import (
	"fmt"
	"io/ioutil"
	"path"

	yaml "gopkg.in/yaml.v2"

	artifacts "github.com/kun-lun/artifacts/pkg/apis"
	"github.com/kun-lun/common/storage"

	"github.com/spf13/afero"
)

type TestInfra struct{}

func (ti TestInfra) BuildSampleManifest() *artifacts.Manifest {
	platform := artifacts.Platform{
		Type: "php",
	}

	networks := []artifacts.VirtualNetwork{
		{
			Name: "vnet-1",
			Subnets: []artifacts.Subnet{
				{
					Range:   "10.10.0.0/24",
					Gateway: "10.10.0.1",
					Name:    "snet-1",
				},
			},
		}}

	loadBalancers := []artifacts.LoadBalancer{
		{
			Name: "kunlun-wenserver-lb",
			SKU:  "standard",
		},
	}

	networkSecurityGroups := []artifacts.NetworkSecurityGroup{
		{
			Name: "nsg_1",
			NetworkSecurityRules: []artifacts.NetworkSecurityRule{
				{
					Name:                     "allow-ssh",
					Priority:                 100,
					Direction:                "Inbound",
					Access:                   "Allow",
					Protocol:                 "Tcp",
					SourcePortRange:          "*",
					DestinationPortRange:     "22",
					SourceAddressPrefix:      "*",
					DestinationAddressPrefix: "*",
				},
			},
		},
	}

	vmGroups := []artifacts.VMGroup{
		{
			Name: "jumpbox",
			Meta: yaml.MapSlice{
				{
					Key:   "group_type",
					Value: "jumpbox",
				},
			},
			SKU:   artifacts.VMStandardDS1V2,
			Count: 1,
			Type:  "VM",
			Storage: &artifacts.VMStorage{
				Image: &artifacts.Image{
					Offer:     "offer1",
					Publisher: "ubuntu",
					SKU:       "sku1",
					Version:   "latest",
				},
				OSDisk: &artifacts.OSDisk{},
				DataDisks: []artifacts.DataDisk{
					{
						DiskSizeGB: 10,
					},
				},
				AzureFiles: []artifacts.AzureFile{
					{
						StorageAccount: "storage_account_1",
						Name:           "azure_file_1",
						MountPoint:     "/mnt/azurefile_1",
					},
				},
			},
			OSProfile: artifacts.VMOSProfile{
				AdminName: "kunlun",
			},
			NetworkInfos: []artifacts.VMNetworkInfo{
				{
					SubnetName:               networks[0].Subnets[0].Name,
					LoadBalancerName:         loadBalancers[0].Name,
					NetworkSecurityGroupName: networkSecurityGroups[0].Name,
					PublicIP:                 "dynamic",
					Outputs: []artifacts.VMNetworkOutput{
						{
							IP:       "172.16.8.4",
							PublicIP: "13.75.71.162",
							Host:     "andliuubuntu.eastasia.cloudapp.azure.com",
						},
					},
				},
			},
			Roles: []artifacts.Role{
				{
					Name: "builtin/jumpbox",
				},
			},
		},
		{
			Name:  "d2v3_group",
			SKU:   artifacts.VMStandardDS1V2,
			Count: 2,
			Type:  "VM",
			OSProfile: artifacts.VMOSProfile{
				AdminName: "kunlun",
			},
			Storage: &artifacts.VMStorage{
				OSDisk: &artifacts.OSDisk{},
				DataDisks: []artifacts.DataDisk{
					{
						DiskSizeGB: 10,
					},
				},
				AzureFiles: []artifacts.AzureFile{},
			},
			NetworkInfos: []artifacts.VMNetworkInfo{
				{
					SubnetName:       networks[0].Subnets[0].Name,
					LoadBalancerName: loadBalancers[0].Name,
					Outputs: []artifacts.VMNetworkOutput{
						{
							IP: "172.16.8.4",
						},
						{
							IP: "172.16.8.4",
						},
					},
				},
			},
			Roles: []artifacts.Role{
				{
					Name: "builtin/php_web_role",
				},
			},
		},
	}

	storageAccounts := []artifacts.StorageAccount{
		{
			Name:     "storage_account_1",
			SKU:      "standard",
			Location: "eastus",
		},
	}

	databases := []artifacts.MysqlDatabase{
		{
			MigrationInformation: &artifacts.MigrationInformation{
				OriginHost:     "asd",
				OriginDatabase: "asd",
				OriginUsername: "asd",
				OriginPassword: "asd",
			},
			Cores:               2,
			Storage:             5,
			BackupRetentionDays: 35,
			Username:            "dbuser",
			Password:            "abcd1234!",
		},
	}

	// The checker add needed resource to manifest
	m := &artifacts.Manifest{
		Schema:                "v0.1",
		IaaS:                  "azure",
		Location:              "eastus",
		Platform:              &platform,
		VMGroups:              vmGroups,
		VNets:                 networks,
		LoadBalancers:         loadBalancers,
		StorageAccounts:       storageAccounts,
		NetworkSecurityGroups: networkSecurityGroups,
		MysqlDatabases:        databases,
	}
	return m
}

func (ti TestInfra) PrepareForDeploymentCmd() storage.Store {

	// File IO
	fs := afero.NewOsFs()
	afs := &afero.Afero{Fs: fs}

	// Configuration
	tempDir, _ := ioutil.TempDir("", "")
	fmt.Printf("store root dir is %s\n", tempDir)
	stateStore := storage.NewStore(tempDir, afs)

	sampleManifest := ti.BuildSampleManifest()
	content, _ := sampleManifest.ToYAML()
	mainArtifactPath, _ := stateStore.GetMainArtifactFilePath()
	afs.WriteFile(mainArtifactPath, content, 0644)

	// write the ops files.
	opsFileContent := `- type: replace
  path: /vm_groups/name=jumpbox/networks/0/outputs?
  value:
    - ip: 192.168.1.3
    - public_ip: 202.120.30.102
    - host: jumpbpbx.xx.com
`
	patchDir, _ := stateStore.GetArtifactsPatchDir()
	opsFilePath := path.Join(patchDir, "ops1.yml")
	fmt.Printf("writing the ops file: %s\n", opsFilePath)
	afs.WriteFile(opsFilePath, []byte(opsFileContent), 0644)
	return stateStore
}
