package apis

import (
	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	"github.com/kun-lun/ashandler/generator"
	"github.com/kun-lun/common/logger"
	"github.com/kun-lun/common/storage"
)

type ASHandler struct {
	asGenerator generator.ASGenerator
}

func NewASHandler(
	stateStore storage.Store,
	logger *logger.Logger,
) ASHandler {
	return ASHandler{
		asGenerator: generator.NewASGenerator(stateStore, logger),
	}
}
func (a ASHandler) Handle(hostGroups []deployments.HostGroup, deployments []deployments.Deployment) error {
	return a.asGenerator.Generate(hostGroups, deployments)
}
