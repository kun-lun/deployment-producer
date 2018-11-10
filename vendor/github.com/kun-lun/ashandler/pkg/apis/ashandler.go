package apis

import (
	"github.com/kun-lun/artifacts/pkg/apis/deployments"
	"github.com/kun-lun/ashandler/generator"
	"github.com/kun-lun/common/fileio"
	"github.com/kun-lun/common/storage"
)

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string) bool
}

type ASHandler struct {
	asGenerator generator.ASGenerator
}

func NewASHandler(
	stateStore storage.Store,
	logger logger,
	fs fileio.Fs,
) ASHandler {
	return ASHandler{
		asGenerator: generator.NewASGenerator(stateStore, logger, fs),
	}
}
func (a ASHandler) Handle(hostGroups []deployments.HostGroup, deployments []deployments.Deployment) error {
	return a.asGenerator.Generate(hostGroups, deployments)
}
