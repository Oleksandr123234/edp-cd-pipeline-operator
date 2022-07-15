package chain

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8sApi "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cdPipeApi "github.com/epam/edp-cd-pipeline-operator/v2/pkg/apis/edp/v1"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/controller/stage/chain/handler"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/controller/stage/rbac"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/platform/helper"
)

const (
	roleKind                       = "Role"
	clusterRoleKind                = "ClusterRole"
	serviceAccountKind             = "ServiceAccount"
	adminConsoleServiceAccountName = "edp-admin-console"
	jenkinsServiceAccountName      = "jenkins"
	adminClusterRoleName           = "admin"
	viewClusterRoleName            = "view"
	clusterOpenshiftType           = "openshift"
	groupKind                      = "Group"
	acViewRoleName                 = "admin-console-view-deployments"
	acViewRbName                   = "ac-deployments-viewer"
	jenkinsAdminRbName             = "jenkins-admin"
)

type ConfigureRbac struct {
	next   handler.CdStageHandler
	client client.Client
	log    logr.Logger
	rbac   rbac.RbacManager
}

type options struct {
	subjects []k8sApi.Subject
	rf       k8sApi.RoleRef
	pr       []k8sApi.PolicyRule
}

func (h ConfigureRbac) ServeRequest(stage *cdPipeApi.Stage) error {
	targetNamespace := generateTargetNamespaceName(stage)
	log := h.log.WithValues("namespace", targetNamespace)
	log.Info("configuring rbac for newly created namespace")

	acViewOpts := buildAcViewRoleOptions(stage.Namespace)
	if err := h.createRole(acViewRoleName, targetNamespace, acViewOpts); err != nil {
		return err
	}
	if err := h.createRoleBinding(acViewRbName, targetNamespace, acViewOpts); err != nil {
		return err
	}

	jenkinsAdminOpts := buildJenkinsAdminRoleOptions(stage.Namespace)
	if err := h.createRoleBinding(jenkinsAdminRbName, targetNamespace, jenkinsAdminOpts); err != nil {
		return err
	}

	if helper.GetPlatformTypeEnv() == clusterOpenshiftType {
		viewGroupRbName := generateViewGroupRbName(stage.Namespace)
		viewGroupOpts := buildViewGroupRoleOptions(stage.Namespace)
		if err := h.createRoleBinding(viewGroupRbName, targetNamespace, viewGroupOpts); err != nil {
			return err
		}
	}
	log.Info("rbac has been configured.")
	return nextServeOrNil(h.next, stage)
}

func (h ConfigureRbac) roleBindingExists(name, namespace string) (bool, error) {
	log := h.log.WithValues("name", name, "namespace", namespace)
	log.Info("check existence of rolebinding")
	if _, err := h.rbac.GetRoleBinding(name, namespace); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("rolebinding doesn't exist")
			return false, nil
		}
		return false, err
	}
	log.Info("rolebinding exists")
	return true, nil
}

func (h ConfigureRbac) createRoleBinding(rbName, namespace string, opts options) error {
	exists, err := h.roleBindingExists(rbName, namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to check existence of %s rolebinding", rbName)
	}

	if exists {
		log.Info("skip creating rolebinding as it does exist", "name", rbName, "namespace", namespace)
		return nil
	}

	if err := h.rbac.CreateRoleBinding(rbName, namespace, opts.subjects, opts.rf); err != nil {
		return errors.Wrapf(err, "unable to create %s rolebinding", rbName)
	}
	return nil
}

func (h ConfigureRbac) roleExists(name, namespace string) (bool, error) {
	log := h.log.WithValues("name", name, "namespace", namespace)
	log.Info("check existence of role")
	if _, err := h.rbac.GetRole(name, namespace); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("role doesn't exist")
			return false, nil
		}
		return false, err
	}
	log.Info("rolebinding exists")
	return true, nil
}

func (h ConfigureRbac) createRole(rName, namespace string, opts options) error {
	exists, err := h.roleExists(rName, namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to check existence of %s role", rName)
	}

	if exists {
		log.Info("skip creating role as it does exist", "name", rName, "namespace", namespace)
		return nil
	}

	if err := h.rbac.CreateRole(rName, namespace, opts.pr); err != nil {
		return errors.Wrapf(err, "unable to create %s role", rName)
	}
	return nil
}

func buildAcViewRoleOptions(sourceNamespace string) options {
	return options{
		subjects: []k8sApi.Subject{
			{
				Kind:      serviceAccountKind,
				Name:      adminConsoleServiceAccountName,
				Namespace: sourceNamespace,
			},
		},
		rf: k8sApi.RoleRef{
			Name:     acViewRoleName,
			APIGroup: k8sApi.GroupName,
			Kind:     roleKind,
		},
		pr: []k8sApi.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
			},
		},
	}
}

func buildJenkinsAdminRoleOptions(sourceNamespace string) options {
	return options{
		subjects: getJenkinsAdminRoleSubjects(sourceNamespace),
		rf: k8sApi.RoleRef{
			Name:     adminClusterRoleName,
			APIGroup: k8sApi.GroupName,
			Kind:     clusterRoleKind,
		},
	}
}

func getJenkinsAdminRoleSubjects(sourceNamespace string) []k8sApi.Subject {
	if helper.GetPlatformTypeEnv() != clusterOpenshiftType {
		return []k8sApi.Subject{
			{
				Kind:      serviceAccountKind,
				Name:      jenkinsServiceAccountName,
				Namespace: sourceNamespace,
			},
		}
	}
	return []k8sApi.Subject{
		{
			Kind: groupKind,
			Name: fmt.Sprintf("%v-edp-super-admin", sourceNamespace),
		},
		{
			Kind: groupKind,
			Name: fmt.Sprintf("%v-edp-admin", sourceNamespace),
		},
		{
			Kind:      serviceAccountKind,
			Name:      jenkinsServiceAccountName,
			Namespace: sourceNamespace,
		},
	}
}

func buildViewGroupRoleOptions(sourceNamespace string) options {
	return options{
		subjects: []k8sApi.Subject{
			{
				Kind: groupKind,
				Name: fmt.Sprintf("%v-edp-view", sourceNamespace),
			},
		},
		rf: k8sApi.RoleRef{
			Name:     viewClusterRoleName,
			APIGroup: k8sApi.GroupName,
			Kind:     clusterRoleKind,
		},
	}
}

func generateTargetNamespaceName(stage *cdPipeApi.Stage) string {
	return fmt.Sprintf("%s-%s", stage.Namespace, stage.Name)
}

func generateViewGroupRbName(namespace string) string {
	return fmt.Sprintf("%s-view", namespace)
}
