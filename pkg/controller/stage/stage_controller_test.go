package stage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	k8sApi "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1alpha1"
	componentApi "github.com/epam/edp-component-operator/pkg/apis/v1/v1"
	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"

	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/util/cluster"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/util/consts"
)

const (
	cdPipeline      = "stub-cdPipeline-name"
	dockerImageName = "docker-image-name"
	name            = "stub-name"
	namespace       = "stub-namespace"
	labelValue      = "stub-data"
	dockerRegistry  = "docker-registry"
)

func getStage(t *testing.T, client client.Client, name string) *v1alpha1.Stage {
	t.Helper()
	stage := &v1alpha1.Stage{}
	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, stage)
	if err != nil {
		t.Fatal(err)
	}
	return stage
}

func createLabelName(pipeName, stageName string) string {
	return fmt.Sprintf("%s/%s", pipeName, stageName)
}

func TestNewReconcileStage_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().Build()
	log := logr.DiscardLogger{}

	expectedReconcileStage := &ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    log.WithName("cd-stage"),
	}

	reconcileStage := NewReconcileStage(fakeClient, scheme, log)
	assert.Equal(t, expectedReconcileStage, reconcileStage)
}

func TestTryToDeleteCDStage_DeletionTimestampIsZero(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{})

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Finalizers: []string{},
		},
		Spec: v1alpha1.StageSpec{
			TriggerType: consts.AutoDeployTriggerType,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	_, err := reconcileStage.tryToDeleteCDStage(context.Background(), stage)
	assert.NoError(t, err)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Equal(t, stageAfterReconcile.Finalizers, []string{foregroundDeletionFinalizerName, envLabelDeletionFinalizer})
}

func TestTryToDeleteCDStage_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.CodebaseImageStream{}, &v1.Namespace{})

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 0,
			DeletionTimestamp: &metaV1.Time{
				Time: time.Now().UTC(),
			},
			Finalizers: []string{envLabelDeletionFinalizer},
		},
		Spec: v1alpha1.StageSpec{
			Name:        name,
			CdPipeline:  cdPipeline,
			TriggerType: consts.AutoDeployTriggerType,
			Order:       0,
		},
	}

	labels := make(map[string]string)
	labels[createLabelName(name, name)] = labelValue

	cdPipeline := &v1alpha1.CDPipeline{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      cdPipeline,
			Namespace: namespace,
		},
		Spec: v1alpha1.CDPipelineSpec{
			InputDockerStreams: []string{dockerImageName},
			Name:               name,
		},
	}

	image := &codebaseApi.CodebaseImageStream{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      dockerImageName,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cdPipeline, image, stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	_, err := reconcileStage.tryToDeleteCDStage(context.Background(), stage)
	assert.NoError(t, err)

	previousImageStream, err := cluster.GetCodebaseImageStream(reconcileStage.client, dockerImageName, namespace)
	assert.NoError(t, err)
	assert.Empty(t, previousImageStream.Labels)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Empty(t, stageAfterReconcile.Finalizers)
}

func TestSetCDPipelineOwnerRef_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.CodebaseImageStream{}, &v1.Namespace{})

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			DeletionTimestamp: &metaV1.Time{
				Time: time.Now().UTC(),
			},
		},
		Spec: v1alpha1.StageSpec{
			Name:       name,
			CdPipeline: cdPipeline,
		},
	}

	cdPipeline := &v1alpha1.CDPipeline{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      cdPipeline,
			Namespace: namespace,
		},
		Spec: v1alpha1.CDPipelineSpec{
			Name: name,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cdPipeline, stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	err := reconcileStage.setCDPipelineOwnerRef(context.Background(), stage)
	assert.NoError(t, err)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Equal(t, cdPipeline.Name, stageAfterReconcile.OwnerReferences[0].Name)
}

func TestSetCDPipelineOwnerRef_OwnerExists(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.CodebaseImageStream{}, &v1.Namespace{})

	ownerReference := metaV1.OwnerReference{
		Kind: consts.CDPipelineKind,
		Name: cdPipeline,
	}

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: []metaV1.OwnerReference{ownerReference},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	err := reconcileStage.setCDPipelineOwnerRef(context.Background(), stage)
	assert.NoError(t, err)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Equal(t, cdPipeline, stageAfterReconcile.OwnerReferences[0].Name)
}

func TestSetFinishStatus_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{})

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	err := reconcileStage.setFinishStatus(context.Background(), stage)
	assert.NoError(t, err)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Equal(t, consts.FinishedStatus, stageAfterReconcile.Status.Status)
}

func TestReconcileStage_Reconcile_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.CodebaseImageStream{}, &v1.Namespace{})

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 0,
			DeletionTimestamp: &metaV1.Time{
				Time: time.Now().UTC(),
			},
			Finalizers: []string{envLabelDeletionFinalizer},
		},
		Spec: v1alpha1.StageSpec{
			Name:        name,
			CdPipeline:  cdPipeline,
			TriggerType: consts.AutoDeployTriggerType,
			Order:       0,
		},
	}

	labels := make(map[string]string)
	labels[createLabelName(name, name)] = labelValue

	cdPipeline := &v1alpha1.CDPipeline{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      cdPipeline,
			Namespace: namespace,
		},
		Spec: v1alpha1.CDPipelineSpec{
			InputDockerStreams: []string{dockerImageName},
			Name:               name,
		},
	}

	image := &codebaseApi.CodebaseImageStream{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      dockerImageName,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cdPipeline, image, stage).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	_, err := reconcileStage.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}})
	assert.NoError(t, err)

	previousImageStream, err := cluster.GetCodebaseImageStream(reconcileStage.client, dockerImageName, namespace)
	assert.NoError(t, err)
	assert.Empty(t, previousImageStream.Labels)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Empty(t, stageAfterReconcile.Finalizers)
}

func TestReconcileStage_ReconcileReconcile_SetOwnerRef(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.CodebaseImageStream{}, &v1.Namespace{}, &componentApi.EDPComponent{}, &k8sApi.RoleBinding{}, &jenkinsApi.JenkinsJob{})

	edpComponent := &componentApi.EDPComponent{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      dockerRegistry,
			Namespace: namespace,
		},
	}

	qualityGate := v1alpha1.QualityGate{}

	stage := &v1alpha1.Stage{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Finalizers: []string{envLabelDeletionFinalizer},
		},
		Spec: v1alpha1.StageSpec{
			Name:         name,
			CdPipeline:   cdPipeline,
			TriggerType:  consts.AutoDeployTriggerType,
			Order:        0,
			QualityGates: []v1alpha1.QualityGate{qualityGate},
		},
	}

	labels := make(map[string]string)
	labels[createLabelName(name, name)] = labelValue

	cdPipeline := &v1alpha1.CDPipeline{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      cdPipeline,
			Namespace: namespace,
		},
		Spec: v1alpha1.CDPipelineSpec{
			InputDockerStreams: []string{dockerImageName},
			Name:               name,
		},
	}

	image := &codebaseApi.CodebaseImageStream{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      dockerImageName,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cdPipeline, image, stage, edpComponent).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	_, err := reconcileStage.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}})
	assert.NoError(t, err)

	stageAfterReconcile := getStage(t, reconcileStage.client, name)
	assert.Equal(t, cdPipeline.Name, stageAfterReconcile.OwnerReferences[0].Name)
	assert.Equal(t, consts.FinishedStatus, stageAfterReconcile.Status.Status)
}

func TestReconcileStage_Reconcile_StageIsNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8sApi.SchemeGroupVersion, &v1alpha1.Stage{})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	reconcileStage := ReconcileStage{
		client: fakeClient,
		scheme: scheme,
		log:    logr.DiscardLogger{},
	}

	stage := &v1alpha1.Stage{}
	err := reconcileStage.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, stage)
	assert.True(t, k8sErrors.IsNotFound(err))

	_, err = reconcileStage.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}})
	assert.NoError(t, err)
}