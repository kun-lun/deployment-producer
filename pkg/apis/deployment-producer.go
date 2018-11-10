package apis

import (
	"github.com/kun-lun/artifacts/pkg/apis"

	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	ashandler "github.com/kun-lun/ashandler/pkg/apis"
	"github.com/kun-lun/common/fileio"
	"github.com/kun-lun/common/storage"
	"github.com/kun-lun/deployment-producer/dpbuilder"
)

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string) bool
}

type DeploymentProducer struct {
	stateStore storage.Store
	logger     logger
	fs         fileio.Fs
}

func NewDeploymentProducer(
	stateStore storage.Store,
	logger logger,
	fs fileio.Fs,
) DeploymentProducer {
	return DeploymentProducer{
		stateStore: stateStore,
		logger:     logger,
		fs:         fs,
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
	dpBuilder := dpbuilder.DeploymentBuilder{}
	hostGroups, deployments, err := dpBuilder.Produce(manifest)
	if err != nil {
		return err
	}

	// generate the ansible scripts based on the deployments.
	asHandler := ashandler.NewASHandler(dp.stateStore, dp.logger, dp.fs)
	err = asHandler.Handle(hostGroups, deployments)
	if err != nil {
		return err
	}
	return nil
}
