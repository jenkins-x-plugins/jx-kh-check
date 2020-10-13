package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

const (
	operatorDeployment     = "jx-git-operator"
	bootJobhHealthExceeded = "BOOT_JOB_HEALTH_TIME_EXCEEDED"
)

type Options struct {
	client kubernetes.Interface
	clock  clock.Clock
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

	if o.clock == nil {
		o.clock = clock.RealClock{}
		if err != nil {
			log.Logger().Fatalf("error creating clock: %v", err)
		}
	}

	return &o, nil
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

func (o Options) findErrors() ([]string, error) {
	var kherrors []string

	currentNamespace, err := kubeclient.CurrentNamespace()
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to find current currentNamespace")
	}

	// first check the git operator is running ok
	kherrors = o.checkGitOperator(currentNamespace)
	if len(kherrors) > 0 {
		return kherrors, nil
	}

	// check boot jobs are running
	return o.checkBootJob(currentNamespace)

}

func (o Options) checkBootJob(currentNamespace string) ([]string, error) {
	var kherrors []string

	jobs, err := o.client.BatchV1().Jobs(currentNamespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=boot",
	})

	if err != nil {
		// don't carry on if there are no boot jobs
		return kherrors, errors.Wrapf(err, "failed to list jobs in namespace %s", currentNamespace)
	}

	if jobs == nil || len(jobs.Items) == 0 {
		kherrors = append(kherrors, fmt.Sprintf("failed to find any boot jobs in namespace %s", currentNamespace))
		// don't carry on if there are no boot jobs
		return kherrors, nil
	}

	// sort jobs so we check the most recent one
	sort.Slice(jobs.Items, func(i, j int) bool {
		return jobs.Items[j].Status.StartTime.Before(jobs.Items[i].Status.StartTime)
	})

	// lets check the most recent job
	jobToCheck := jobs.Items[0]

	// first lest see if it seems stuck
	if jobToCheck.Status.CompletionTime == nil {

		jobTimeExceeded := 30 * time.Minute
		overrideDelay := os.Getenv(bootJobhHealthExceeded)
		if overrideDelay != "" {
			delay, err := time.ParseDuration(overrideDelay)
			if err != nil {
				return kherrors, errors.Wrapf(err, "failed to parse %s into an integer", overrideDelay)
			}
			jobTimeExceeded = delay * time.Minute
		}

		if jobToCheck.Status.StartTime == nil {
			return append(kherrors, fmt.Sprintf("latest boot job %s has not started, it could be stuck", jobToCheck.Name)), nil
		}

		if jobToCheck.Status.StartTime.Add(jobTimeExceeded).Before(o.clock.Now()) {
			kherrors = append(kherrors, fmt.Sprintf("latest boot job %s has been running for more than %s, it could be stuck", jobToCheck.Name, jobTimeExceeded.String()))
		}
	}

	// check if it failed
	if jobToCheck.Status.Failed > 0 {
		kherrors = append(kherrors, fmt.Sprintf("latest boot job %s has a failed run", jobToCheck.Name))
	}

	return kherrors, nil
}

func (o Options) checkGitOperator(currentNamespace string) []string {
	var kherrors []string
	deployment, err := o.client.AppsV1().Deployments(currentNamespace).Get(context.TODO(), operatorDeployment, metav1.GetOptions{})
	if err != nil {
		return append(kherrors, fmt.Sprintf("failed to find %s in namespace %s", operatorDeployment, currentNamespace))
	}

	if deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
		kherrors = append(kherrors, fmt.Sprintf("ready pods (%d) to not match the expected number (%d)", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas))
	}
	return kherrors
}
