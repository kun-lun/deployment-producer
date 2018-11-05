package apis

import (
	"github.com/kun-lun/artifacts/pkg/apis"
	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	"github.com/kun-lun/common/errors"
	"github.com/kun-lun/common/storage"
)

type DeploymentProducer struct {
	stateStore storage.Store
}

func NewDeploymentProducer(
	stateStore storage.Store,
) DeploymentProducer {
	return DeploymentProducer{
		stateStore: stateStore,
	}
}

type deploymentItem struct {
	hostGroup  deployments.HostGroup
	deployment deployments.Deployment
}

func (dp DeploymentProducer) Produce(
	manifest apis.Manifest,
) error {
	// generate the deployments
	deployment_items := []deploymentItem{}

	for _, item := range manifest.VMGroups {
		if item.Roles != nil && len(item.Roles) > 0 {
			deployment_items = append(deployment_items, deploymentItem{})
		}
	}
	// generate the ansible scripts based on the deployments.

	return &errors.NotImplementedError{}
}
