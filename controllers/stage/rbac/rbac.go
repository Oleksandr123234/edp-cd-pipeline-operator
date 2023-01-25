package rbac

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	k8sApi "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	crNameLogKey = "name"
)

type RbacManager interface {
	GetRoleBinding(name, namespace string) (*k8sApi.RoleBinding, error)
	CreateRoleBinding(name, namespace string, subjects []k8sApi.Subject, roleRef k8sApi.RoleRef) error
	GetRole(name, namespace string) (*k8sApi.Role, error)
	CreateRole(name, namespace string, rules []k8sApi.PolicyRule) error
}

type KubernetesRbac struct {
	client client.Client
	log    logr.Logger
}

func InitRbacManager(c client.Client) RbacManager {
	return KubernetesRbac{
		client: c,
		log:    ctrl.Log.WithName("rbac-manager"),
	}
}

func (s KubernetesRbac) GetRoleBinding(name, namespace string) (*k8sApi.RoleBinding, error) {
	log := s.log.WithValues(crNameLogKey, name, "namespace", namespace)
	log.Info("getting role binding")

	rb := &k8sApi.RoleBinding{}
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, rb); err != nil {
		return nil, fmt.Errorf("failed to get role binding: %w", err)
	}

	return rb, nil
}

func (s KubernetesRbac) CreateRoleBinding(name, namespace string, subjects []k8sApi.Subject, roleRef k8sApi.RoleRef) error {
	log := s.log.WithValues(crNameLogKey, name)
	log.Info("creating rolebinding")

	rb := &k8sApi.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}
	if err := s.client.Create(context.Background(), rb); err != nil {
		return fmt.Errorf("failed to create role binding: %w", err)
	}

	log.Info("rolebinding has been created")

	return nil
}

func (s KubernetesRbac) GetRole(name, namespace string) (*k8sApi.Role, error) {
	log := s.log.WithValues(crNameLogKey, name, "namespace", namespace)
	log.Info("getting role binding")

	r := &k8sApi.Role{}
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, r); err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return r, nil
}

func (s KubernetesRbac) CreateRole(name, namespace string, rules []k8sApi.PolicyRule) error {
	log := s.log.WithValues(crNameLogKey, name)
	log.Info("creating role")

	r := &k8sApi.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: rules,
	}
	if err := s.client.Create(context.Background(), r); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	log.Info("role has been created")

	return nil
}