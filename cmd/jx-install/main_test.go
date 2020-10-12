package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/clock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOptions_checkGitOperator(t *testing.T) {
	// getting the current namespace is found from a local kube config file
	err := os.Setenv("KUBECONFIG", filepath.Join("test_data", "test-config"))
	assert.NoError(t, err)

	type fields struct {
		sr *v1.Deployment
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{name: "no_errors", fields: fields{sr: &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      operatorDeployment,
				Namespace: "cheese",
			},
			Spec:   v1.DeploymentSpec{Replicas: int32Ptr(2)},
			Status: v1.DeploymentStatus{ReadyReplicas: 2},
		}}},
		{name: "no_git_operator", fields: fields{sr: &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
			},
		}}, want: []string{"failed to find jx-git-operator in namespace cheese"}},
		{name: "no_ready_pods", fields: fields{sr: &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      operatorDeployment,
				Namespace: "cheese",
			},
			Spec: v1.DeploymentSpec{Replicas: int32Ptr(2)},
		}}, want: []string{"ready pods (0) to not match the expected number (2)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tt.fields.sr)
			o := Options{
				client: client,
			}
			got := o.checkGitOperator("cheese")
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

func TestOptions_checkBootJob(t *testing.T) {
	// getting the current namespace is found from a local kube config file
	err := os.Setenv("KUBECONFIG", filepath.Join("test_data", "test-config"))
	assert.NoError(t, err)

	now := time.Now()
	fakeClock := clock.NewFakeClock(now)
	currentTime := metav1.NewTime(fakeClock.Now())

	time10MinsAgo := metav1.NewTime(now.Add(-10 * time.Minute))
	time20MinsAgo := metav1.NewTime(now.Add(-20 * time.Minute))

	type fields struct {
		objects         []runtime.Object
		minsInTheFuture time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{name: "no_boot_job", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
			},
		}}}, want: []string{"failed to find any boot jobs in namespace cheese"}},
		{name: "boot_job_not_started", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
		}}}, want: []string{"latest boot job foo has not started, it could be stuck"}},
		{name: "boot_job_running_more_than_default_exceeded_time", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &currentTime,
			},
		}}, minsInTheFuture: 35}, want: []string{"latest boot job foo has been running for more than 30m0s, it could be stuck"}},
		{name: "sort_get_latest_boot_job", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &time10MinsAgo,
				Failed:    1,
			},
		}, &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &time20MinsAgo,
				Failed:    1,
			},
		}}, minsInTheFuture: 10}, want: []string{"latest boot job foo has a failed run"}},
		{name: "sort_get_latest_boot_job_change_start_time_order", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &time20MinsAgo,
				Failed:    1,
			},
		}, &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &time10MinsAgo,
				Failed:    1,
			},
		}}, minsInTheFuture: 10}, want: []string{"latest boot job bar has a failed run"}},
		{name: "no_errors", fields: fields{objects: []runtime.Object{&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "cheese",
				Labels:    map[string]string{"app": "boot"},
			},
			Status: batchv1.JobStatus{
				StartTime: &currentTime,
			},
		}}, minsInTheFuture: 10}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.fields.minsInTheFuture != 0 {
				t2 := now.Add(tt.fields.minsInTheFuture * time.Minute)
				fakeClock.SetTime(t2)
			}

			client := fake.NewSimpleClientset(tt.fields.objects...)
			o := Options{
				client: client,
				clock:  fakeClock,
			}
			got, err := o.checkBootJob("cheese")
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

func int32Ptr(i int32) *int32 { return &i }
