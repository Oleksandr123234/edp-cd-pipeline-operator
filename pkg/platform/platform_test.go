package platform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlatformTypeEnv_Success(t *testing.T) {
	stubPlatformType := "stubPlatformType"

	err := os.Setenv(TypeEnv, stubPlatformType)
	if err != nil {
		t.Fatalf("cannot set an env variable: %v", err)
	}

	defer func() {
		err := os.Unsetenv(TypeEnv)
		if err != nil {
			t.Fatalf("cannot unset an env variable: %v", err)
		}
	}()

	platformType := GetPlatformTypeEnv()
	assert.Equal(t, stubPlatformType, platformType)
}

func TestGetPlatformTypeEnv_PlatformTypeIsNotSet(t *testing.T) {
	err := os.Unsetenv(TypeEnv)
	if err != nil {
		t.Fatalf("cannot unset an env variable: %v", err)
	}

	platformType := GetPlatformTypeEnv()
	assert.Equal(t, Openshift, platformType)
}

func TestIsKubernetes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setEnv func(t *testing.T)
		want   bool
	}{
		{
			name: "platform type is kubernetes",
			setEnv: func(t *testing.T) {
				t.Setenv(TypeEnv, Kubernetes)
			},
			want: true,
		},
		{
			name: "platform type is openshift",
			setEnv: func(t *testing.T) {
				t.Setenv(TypeEnv, Openshift)
			},
			want: false,
		},
		{
			name: "platform type is not set",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != nil {
				tt.setEnv(t)
			}

			assert.Equal(t, IsKubernetes(), tt.want)
		})
	}
}

func TestKioskEnabled(t *testing.T) {
	tests := []struct {
		name   string
		setEnv func(t *testing.T)
		want   bool
	}{
		{
			name: "kiosk is enabled",
			setEnv: func(t *testing.T) {
				t.Setenv(KioskEnabledEnv, "true")
			},
			want: true,
		},
		{
			name: "kiosk is disabled",
			setEnv: func(t *testing.T) {
				t.Setenv(KioskEnabledEnv, "false")
			},
			want: false,
		},
		{
			name: "kiosk is not set",
			setEnv: func(t *testing.T) {
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnv(t)

			assert.Equal(t, tt.want, KioskEnabled())
		})
	}
}
