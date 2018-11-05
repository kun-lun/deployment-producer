package deployments

import (
	"github.com/kun-lun/artifacts/pkg/apis"
	yaml "gopkg.in/yaml.v2"
)

type Deployment struct {
	HostGroupName string
	Vars          yaml.MapSlice
	Roles         []apis.Role
}
