package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/jenkins-x/jx-api/v3/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	jxClient versioned.Interface
}

func main() {

	log.Logger().Infof("starting jx-webhooks health checks")

	o, err := newOptions()
	if err != nil {
		log.Logger().Fatalf("failed to validate options: %v", err)
		return
	}

	kherrors, err := o.findErrors()
	if err != nil {
		log.Logger().Fatalf("failed to list source repositories: %v", err)
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
	kherrors := []string{}

	namespace, err := kubeclient.CurrentNamespace()
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to find current namespace")
	}

	// lookup all source repositories and error if any do not have the webhook created annotation
	sourceRepositories, err := o.jxClient.JenkinsV1().SourceRepositories(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to list source repositories")
	}

	for _, sr := range sourceRepositories.Items {

		value := strings.ToLower(sr.Annotations["webhook.jenkins-x.io"])
		if value == "true" {
			continue
		}

		message := sr.Annotations["webhook.jenkins-x.io/error"]

		if message != "" {
			kherrors = append(kherrors, fmt.Sprintf("no webhook registered for %s: %s", sr.Name, message))
		} else {
			kherrors = append(kherrors, fmt.Sprintf("no webhook registered for %s", sr.Name))
		}
	}

	return kherrors, nil
}

func newOptions() (*Options, error) {
	o := Options{}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		log.Logger().Fatalf("failed to get kubernetes config: %v", err)
	}

	if o.jxClient == nil {
		o.jxClient, err = versioned.NewForConfig(cfg)
		if err != nil {
			log.Logger().Fatalf("error building jx client: %v", err)
		}
	}

	return &o, nil
}
