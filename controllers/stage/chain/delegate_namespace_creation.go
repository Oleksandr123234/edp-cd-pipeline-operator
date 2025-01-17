package chain

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cdPipeApi "github.com/epam/edp-cd-pipeline-operator/v2/api/v1"
	"github.com/epam/edp-cd-pipeline-operator/v2/controllers/stage/chain/handler"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/kiosk"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/platform"
)

// DelegateNamespaceCreation is a stage chain element that decides whether to create a namespace, kiosk space or project.
type DelegateNamespaceCreation struct {
	next   handler.CdStageHandler
	client client.Client
	log    logr.Logger
}

// ServeRequest creates for kubernetes platform PutNamespace or PutKioskSpace if the kiosk is enabled.
// For platform openshift it creates PutOpenshiftProject.
// By default, it creates PutOpenshiftProject.
// If the namespace is not managed by the operator, it creates CheckNamespaceExist.
func (c DelegateNamespaceCreation) ServeRequest(stage *cdPipeApi.Stage) error {
	logger := c.log.WithValues("stage name", stage.Name)

	if !platform.ManageNamespace() {
		logger.Info("Namespace is not managed by the operator")

		return nextServeOrNil(CheckNamespaceExist(c), stage)
	}

	if platform.IsKubernetes() {
		logger.Info("Platform is kubernetes")

		if platform.KioskEnabled() {
			logger.Info("Kiosk is enabled")

			return nextServeOrNil(PutKioskSpace{
				next:   c.next,
				space:  kiosk.InitSpace(c.client),
				client: c.client,
				log:    c.log,
			}, stage)
		}

		logger.Info("Kiosk is disabled")

		return nextServeOrNil(PutNamespace(c), stage)
	}

	logger.Info("Platform is openshift")

	return nextServeOrNil(PutOpenshiftProject(c), stage)
}
