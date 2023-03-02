package chain

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	rbacApi "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cdPipeApi "github.com/epam/edp-cd-pipeline-operator/v2/api/v1"
	"github.com/epam/edp-cd-pipeline-operator/v2/controllers/stage/chain/util"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/platform"
	"github.com/epam/edp-cd-pipeline-operator/v2/pkg/rbac"
)

func TestConfigureJenkinsRbac_ServeRequest(t *testing.T) {
	const namespace = "test-ns"

	scheme := runtime.NewScheme()
	require.NoError(t, cdPipeApi.AddToScheme(scheme))
	require.NoError(t, rbacApi.AddToScheme(scheme))

	tests := []struct {
		name      string
		stage     *cdPipeApi.Stage
		prepare   func(t *testing.T)
		objects   []runtime.Object
		wantErr   require.ErrorAssertionFunc
		wantCheck func(t *testing.T, stage *cdPipeApi.Stage, k8sClient client.Client)
	}{
		{
			name: "rbac configuration is successful",
			stage: &cdPipeApi.Stage{
				ObjectMeta: metaV1.ObjectMeta{
					Namespace: namespace,
					Name:      "test-stage",
				},
			},
			wantErr: require.NoError,
			wantCheck: func(t *testing.T, stage *cdPipeApi.Stage, k8sClient client.Client) {
				require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      jenkinsAdminRbName,
					Namespace: util.GenerateNamespaceName(stage),
				}, &rbacApi.RoleBinding{}))
			},
		},
		{
			name: "rbac already exists",
			stage: &cdPipeApi.Stage{
				ObjectMeta: metaV1.ObjectMeta{
					Namespace: namespace,
					Name:      "test-stage",
				},
			},
			objects: []runtime.Object{
				&rbacApi.RoleBinding{
					ObjectMeta: metaV1.ObjectMeta{
						Name: jenkinsAdminRbName,
						Namespace: util.GenerateNamespaceName(&cdPipeApi.Stage{
							ObjectMeta: metaV1.ObjectMeta{
								Namespace: namespace,
								Name:      "test-stage",
							},
						}),
					},
				},
			},
			wantErr: require.NoError,
			wantCheck: func(t *testing.T, stage *cdPipeApi.Stage, k8sClient client.Client) {
				require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      jenkinsAdminRbName,
					Namespace: util.GenerateNamespaceName(stage),
				}, &rbacApi.RoleBinding{}))
			},
		},
		{
			name: "rbac configuration is successful for kubernetes platform",
			stage: &cdPipeApi.Stage{
				ObjectMeta: metaV1.ObjectMeta{
					Namespace: namespace,
					Name:      "test-stage",
				},
			},
			prepare: func(t *testing.T) {
				t.Setenv(platform.TypeEnv, platform.Kubernetes)
			},
			wantErr: require.NoError,
			wantCheck: func(t *testing.T, stage *cdPipeApi.Stage, k8sClient client.Client) {
				require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      jenkinsAdminRbName,
					Namespace: util.GenerateNamespaceName(stage),
				}, &rbacApi.RoleBinding{}))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t)
			}

			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.objects...).Build()

			h := ConfigureJenkinsRbac{
				client: k8sClient,
				log:    logr.Discard(),
				rbac:   rbac.NewRbacManager(k8sClient, logr.Discard()),
			}

			err := h.ServeRequest(tt.stage)
			tt.wantErr(t, err)
			tt.wantCheck(t, tt.stage, k8sClient)
		})
	}
}