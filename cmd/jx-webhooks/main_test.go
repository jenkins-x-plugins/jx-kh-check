package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v3/pkg/client/clientset/versioned/fake"
)

func TestOptions_findErrors(t *testing.T) {
	// getting the current namespace is found from a local kube config file
	err := os.Setenv("KUBECONFIG", filepath.Join("test_data", "test-config"))
	assert.NoError(t, err)

	type fields struct {
		sr *v1.SourceRepository
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{name: "no_webhook_annotation", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"cheese": "wine"},
				Namespace:   "cheese",
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "no_errors", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": "true"},
				Namespace:   "cheese",
			},
		},
		}, want: []string{}, wantErr: false},

		{name: "webhook_false", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": "false"},
				Namespace:   "cheese",
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "webhook_unknown", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": ""},
				Namespace:   "cheese",
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "error_with_webhook_message", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io/error": "something bad happened"},
				Namespace:   "cheese",
			},
		},
		}, want: []string{"no webhook registered for foo: something bad happened"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tt.fields.sr)
			o := Options{
				jxClient: client,
			}
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
