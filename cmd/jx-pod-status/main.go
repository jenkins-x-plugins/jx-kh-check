// Package podStatus implements a pod health checker for Kuberhealthy.  Pods are checked
// to ensure they are not restarting too much and are in a healthy lifecycle phase.
package main

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	checkclient "github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	envVarTargetNamespace = "TARGET_NAMESPACE"
	envVarSkipDuration    = "SKIP_DURATION"
)

type Options struct {
	client kubernetes.Interface
}

func newOptions() (*Options, error) {
	o := Options{}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		log.Logger().Fatalf("failed to get kubernetes config: %v", err)
	}

	if o.client == nil {
		o.client, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Logger().Fatalf("error building kubernetes client: %v", err)
		}
	}

	return &o, nil
}

func init() {
	checkclient.Debug = true
}

func main() {

	log.Logger().Infof("starting jx-install health checks")

	o, err := newOptions()
	if err != nil {
		log.Logger().Fatalf("failed to validate options: %v", err)
		return
	}

	kherrors, err := o.findErrors()
	if err != nil {
		log.Logger().Fatalf("failed to find errors: %v", err)
	}

	if len(kherrors) == 0 {
		err = checkclient.ReportSuccess()
		if err != nil {
			log.Logger().Fatalf("failed to report success status %v", err)
		}
	} else {
		err = checkclient.ReportFailure(kherrors)
		if err != nil {
			log.Logger().Fatalf("failed to report failure status %v", err)
		}
	}

	log.Logger().Infof("successfully reported")
}

// finds pods that are older than 10 minutes and are in an unhealthy lifecycle phase
func (o Options) findErrors() ([]string, error) {

	skipDurationEnv := os.Getenv(envVarSkipDuration)
	if skipDurationEnv == "" {
		skipDurationEnv = "10m"
	}

	namespace := os.Getenv(envVarTargetNamespace)

	var kherrors []string

	if namespace == "" {
		// it is the same value but we are being explicit that we are listing pods in all namespaces
		namespace = v1.NamespaceAll
	}

	pods, err := o.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app!=kuberhealthy-check,source!=kuberhealthy"})
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to list pods in namespace '%s'", namespace)
	}

	// calculate acceptable times for pods to be skipped in
	skipDuration, err := time.ParseDuration(skipDurationEnv)
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to parse skipDurationEnv %s", skipDurationEnv)
	}

	checkTime := time.Now()
	skipBarrier := checkTime.Add(-skipDuration)

	// start iteration over pods
	for _, pod := range pods.Items {
		// check if the pod age is over 10 minutes
		if pod.CreationTimestamp.Time.After(skipBarrier) {
			log.Logger().Infof("skipping checks on pod because it is too young: %s/%s", pod.Namespace, pod.Name)
			continue
		}

		// pods that are in phase Running/Succeeded are healthy
		// pods that are in phase Pending/Failed/Unknown are unhealthy and added to our list of failed pods
		// log if there is no match to the 5 possible pod status phases
		switch {
		case pod.Status.Phase == v1.PodRunning:
			continue
		case pod.Status.Phase == v1.PodSucceeded:
			continue
		case pod.Status.Phase == v1.PodPending:
			kherrors = append(kherrors, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
		case pod.Status.Phase == v1.PodFailed:
			// lets not report if this is a build pod because because failing builds have a pod failed status
			if !isBuildPod(pod) {
				kherrors = append(kherrors, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
			}
		case pod.Status.Phase == v1.PodUnknown:
			kherrors = append(kherrors, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
		default:
			log.Logger().Info("pod: "+pod.Name+" in namespace: "+pod.Namespace+" is not in one of the five possible pod status phases " + string(pod.Status.Phase) + " ")
		}
	}

	return kherrors, nil

}

func isBuildPod(pod v1.Pod) bool {
	switch {
	case pod.Labels["created-by-lighthouse"] == "true":
		return true
	case pod.Annotations["tekton.dev/ready"] != "READY":
		return true
	case pod.Labels["jenkins.io/pipelineType"] == "build":
		return true
	default:
		return false
	}
}
