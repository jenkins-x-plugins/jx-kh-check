package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jenkins-x/jx-secret/pkg/extsecrets"
	"github.com/jenkins-x/jx-secret/pkg/extsecrets/testsecrets"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestOptions_findErrors(t *testing.T) {

	var err error

	// getting the current namespace is found from a local kube config file
	err = os.Setenv("KUBECONFIG", filepath.Join("test_data", "test-config"))
	assert.NoError(t, err)

	scheme := runtime.NewScheme()
	o := Options{}

	tests := []struct {
		name    string
		want    []string
		wantErr bool
	}{
		{name: "no_error", want: []string{}, wantErr: false},
		{name: "one_error", want: []string{"ERROR, 5 NOT_FOUND: Secret [projects/123/secrets/tf-foo-jenkins-npm-token] not found or has no versions."}, wantErr: false},
		{name: "two_errors", want: []string{"ERROR, 5 NOT_FOUND: Secret [projects/123/secrets/tf-foo-jenkins-maven-settings] not found or has no versions.", "ERROR, 5 NOT_FOUND: Secret [projects/123/secrets/tf-foo-jenkins-npm-token] not found or has no versions."}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dynObjects := testsecrets.LoadExtSecretDir(t, "cheese", filepath.Join("test_data", tt.name))
			fakeDynClient := testsecrets.NewFakeDynClient(scheme, dynObjects...)
			o.client, err = extsecrets.NewClient(fakeDynClient)
			require.NoError(t, err, "failed to create fake extsecrets Client")

			got, err := o.findErrors()
			if (err != nil) != tt.wantErr {
				t.Errorf("findErrors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findErrors() got = %v, want %v", got, tt.want)
			}
		})
	}
}
