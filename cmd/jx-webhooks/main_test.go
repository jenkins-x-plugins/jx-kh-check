package main

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v3/pkg/client/clientset/versioned/fake"
)

func TestOptions_findErrors(t *testing.T) {

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
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "no_errors", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": "true"},
			},
		},
		}, want: []string{}, wantErr: false},

		{name: "webhook_false", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": "false"},
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "webhook_unknown", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io": ""},
			},
		},
		}, want: []string{"no webhook registered for foo"}, wantErr: false},

		{name: "error_with_webhook_message", fields: fields{sr: &v1.SourceRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{"webhook.jenkins-x.io/error": "something bad happened"},
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
