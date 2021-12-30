package chain

import (
	"encoding/json"
	"testing"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1alpha1"
	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/util/common"
)

const (
	gitServer      = "gitServer"
	library        = "library-name"
	deploymentType = "auto"
	cdPipelineName = "stub-cdPipeline-name"
	branch         = "branch"
)

func putJenkinsJobSchemeInit(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.Stage{}, &v1alpha1.CDPipeline{}, &codebaseApi.Codebase{}, &codebaseApi.GitServer{}, &jenkinsApi.JenkinsJob{})
	return scheme
}

func putJenkinsJobCreateCdPipeline(t *testing.T) *v1alpha1.CDPipeline {
	t.Helper()
	return &v1alpha1.CDPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdPipelineName,
			Namespace: namespace,
		},
		Spec: v1alpha1.CDPipelineSpec{
			DeploymentType: deploymentType,
		},
		Status: v1alpha1.CDPipelineStatus{},
	}
}

func putJenkinsJobCreateStage(t *testing.T) v1alpha1.Stage {
	t.Helper()
	return v1alpha1.Stage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.StageSpec{
			QualityGates: []v1alpha1.QualityGate{
				{
					QualityGateType: "autotests",
				},
			},
			Source: v1alpha1.Source{
				Type: "library",
				Library: v1alpha1.Library{
					Name:   library,
					Branch: branch,
				},
			},
			CdPipeline: cdPipelineName,
		},
	}
}

func putJenkinsJobCreateCodebase(t *testing.T) *codebaseApi.Codebase {
	t.Helper()
	return &codebaseApi.Codebase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      library,
			Namespace: namespace,
		},
		Spec: codebaseApi.CodebaseSpec{
			GitServer: gitServer,
		},
	}
}

func putJenkinsJobCreateGitServer(t *testing.T) *codebaseApi.GitServer {
	t.Helper()
	return &codebaseApi.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gitServer,
			Namespace: namespace,
		},
	}
}

func TestGetQualityGateStagesMethod_ShouldReturnParsedStagesScenarioFirst(t *testing.T) {

	qualityGates := []v1alpha1.QualityGate{
		{
			QualityGateType: "autotests",
			StepName:        "aut1",
			AutotestName:    common.GetStringP("aut1"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut2",
			AutotestName:    common.GetStringP("aut2"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut3",
			AutotestName:    common.GetStringP("aut3"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut4",
			AutotestName:    common.GetStringP("aut4"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "manual",
			StepName:        "man1",
			AutotestName:    common.GetStringP("man1"),
			BranchName:      common.GetStringP("master"),
		},
	}
	stages, err := getQualityGateStages(qualityGates)
	assert.NoError(t, err)
	assert.NotNil(t, stages)
	expected := "[{\"name\":\"autotests\",\"step_name\":\"aut1\"},{\"name\":\"autotests\",\"step_name\":\"aut2\"},{\"name\":\"autotests\",\"step_name\":\"aut3\"},{\"name\":\"autotests\",\"step_name\":\"aut4\"}],{\"name\":\"manual\",\"step_name\":\"man1\"}"
	assert.Equal(t, expected, *stages)
}

func TestGetQualityGateStagesMethod_ShouldReturnParsedStagesScenarioSecond(t *testing.T) {
	qualityGates := []v1alpha1.QualityGate{
		{
			QualityGateType: "autotests",
			StepName:        "aut1",
			AutotestName:    common.GetStringP("aut1"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut3",
			AutotestName:    common.GetStringP("aut3"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut4",
			AutotestName:    common.GetStringP("aut4"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "manual",
			StepName:        "man1",
			AutotestName:    common.GetStringP("man1"),
			BranchName:      common.GetStringP("master"),
		},
		{
			QualityGateType: "autotests",
			StepName:        "aut2",
			AutotestName:    common.GetStringP("aut2"),
			BranchName:      common.GetStringP("master"),
		},
	}
	stages, err := getQualityGateStages(qualityGates)
	assert.NoError(t, err)
	assert.NotNil(t, stages)
	expected := "[{\"name\":\"autotests\",\"step_name\":\"aut1\"},{\"name\":\"autotests\",\"step_name\":\"aut3\"},{\"name\":\"autotests\",\"step_name\":\"aut4\"}],{\"name\":\"manual\",\"step_name\":\"man1\"},{\"name\":\"autotests\",\"step_name\":\"aut2\"}"
	assert.Equal(t, expected, *stages)
}

func TestGetQualityGateStagesMethod_ShouldReturnParsedStagesAsNilScenarioFirst(t *testing.T) {
	stages, err := getQualityGateStages(nil)
	assert.NoError(t, err)
	assert.Nil(t, stages)
}

func TestGetQualityGateStagesMethod_ShouldReturnParsedStagesAsNilScenarioSecond(t *testing.T) {
	var qualityGates []v1alpha1.QualityGate
	stages, err := getQualityGateStages(qualityGates)
	assert.NoError(t, err)
	assert.Nil(t, stages)
}

func TestCreateJenkinsJobConfig_Success(t *testing.T) {
	cdPipeline := &v1alpha1.CDPipeline{
		Spec: v1alpha1.CDPipelineSpec{
			DeploymentType: deploymentType,
		},
		Status: v1alpha1.CDPipelineStatus{},
	}

	stage := v1alpha1.Stage{
		Spec: v1alpha1.StageSpec{
			QualityGates: []v1alpha1.QualityGate{
				{
					QualityGateType: "autotests",
				},
			},
		},
	}

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(cdPipeline, &stage).Build(),
		log:    logr.DiscardLogger{},
	}

	resultJson, err := putJenkinsJob.createJenkinsJobConfig(stage)
	assert.NoError(t, err)

	result := make(map[string]string)
	err = json.Unmarshal(resultJson, &result)
	assert.NoError(t, err)
	assert.Equal(t, deploymentType, result["DEPLOYMENT_TYPE"])
}

func TestCreateJenkinsJobConfig_WithLibraryParams(t *testing.T) {
	cdPipeline := putJenkinsJobCreateCdPipeline(t)
	stage := putJenkinsJobCreateStage(t)
	codeBase := putJenkinsJobCreateCodebase(t)
	gitServer := putJenkinsJobCreateGitServer(t)

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(cdPipeline, &stage, codeBase, gitServer).Build(),
		log:    logr.DiscardLogger{},
	}

	resultJson, err := putJenkinsJob.createJenkinsJobConfig(stage)
	assert.NoError(t, err)

	result := make(map[string]string)
	err = json.Unmarshal(resultJson, &result)
	assert.NoError(t, err)
	assert.Equal(t, branch, result["LIBRARY_BRANCH"])
}

func TestTryToUpdateJenkinsJob_Success(t *testing.T) {
	cdPipeline := putJenkinsJobCreateCdPipeline(t)
	stage := putJenkinsJobCreateStage(t)
	codeBase := putJenkinsJobCreateCodebase(t)
	gitServer := putJenkinsJobCreateGitServer(t)

	jenkinsJob := jenkinsApi.JenkinsJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(cdPipeline, &stage, codeBase, gitServer, &jenkinsJob).Build(),
		log:    logr.DiscardLogger{},
	}

	err := putJenkinsJob.tryToUpdateJenkinsJobConfig(stage)
	assert.NoError(t, err)

	jenkinsJobAfterUpdate, err := putJenkinsJob.getJenkinsJob(name, namespace)
	assert.NoError(t, err)
	assert.NotNil(t, jenkinsJobAfterUpdate.Spec.Job.Config)
	assert.NotEmpty(t, jenkinsJobAfterUpdate.Spec.Job.Config)
}

func TestTryToUpdateJenkinsJob_NotFound(t *testing.T) {
	stage := v1alpha1.Stage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	jenkinsJob := jenkinsApi.JenkinsJob{}

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(&stage, &jenkinsJob).Build(),
		log:    logr.DiscardLogger{},
	}

	err := putJenkinsJob.tryToUpdateJenkinsJobConfig(stage)
	assert.Nil(t, err)
}

func TestTryToCreateJenkinsJobConfig_Success(t *testing.T) {
	cdPipeline := putJenkinsJobCreateCdPipeline(t)
	stage := putJenkinsJobCreateStage(t)
	codeBase := putJenkinsJobCreateCodebase(t)
	gitServer := putJenkinsJobCreateGitServer(t)

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(cdPipeline, &stage, codeBase, gitServer).Build(),
		log:    logr.DiscardLogger{},
	}

	err := putJenkinsJob.tryToCreateJenkinsJob(stage)
	jenkinsJobAfterUpdate, err := putJenkinsJob.getJenkinsJob(name, namespace)
	assert.NoError(t, err)
	assert.NotNil(t, jenkinsJobAfterUpdate.Spec.Job.Config)
	assert.NotEmpty(t, jenkinsJobAfterUpdate.Spec.Job.Config)
}

func TestPutJenkinsJob_ServeRequest_Success(t *testing.T) {
	cdPipeline := putJenkinsJobCreateCdPipeline(t)
	stage := putJenkinsJobCreateStage(t)
	codeBase := putJenkinsJobCreateCodebase(t)
	gitServer := putJenkinsJobCreateGitServer(t)

	putJenkinsJob := PutJenkinsJob{
		client: fake.NewClientBuilder().WithScheme(putJenkinsJobSchemeInit(t)).WithObjects(cdPipeline, &stage, codeBase, gitServer).Build(),
		log:    logr.DiscardLogger{},
	}

	err := putJenkinsJob.ServeRequest(&stage)
	jenkinsJobAfterUpdate, err := putJenkinsJob.getJenkinsJob(name, namespace)
	assert.NoError(t, err)
	assert.NotNil(t, jenkinsJobAfterUpdate.Spec.Job.Config)
	assert.NotEmpty(t, jenkinsJobAfterUpdate.Spec.Job.Config)
}
