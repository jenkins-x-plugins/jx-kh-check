package main

import (
	"os"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/jenkins-x/jx-secret/pkg/extsecrets"
)

const envVarTargetNamespace = "TARGET_NAMESPACE"

type Options struct {
	client extsecrets.Interface
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

	namespace := os.Getenv(envVarTargetNamespace)
	if namespace == "" {
		// it is the same value but we are being explicit that we are listing pods in all namespaces
		namespace = v1.NamespaceAll
	}

	// lookup all source repositories and error if any do not have the webhook created annotation
	externalSecrets, err := o.client.List(namespace)
	if err != nil {
		return kherrors, errors.Wrapf(err, "failed to list external secrets")
	}

	for _, es := range externalSecrets {
		// ignore external secrets with no status
		if es.Status == nil {
			continue
		}
		if es.Status.Status == "SUCCESS" {
			continue
		}

		kherrors = append(kherrors, es.Status.Status)
	}
	return kherrors, nil
}

func newOptions() (*Options, error) {
	o := Options{}
	var err error
	if o.client == nil {
		o.client, err = extsecrets.NewClient(nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create an extsecret Client")
		}
	}

	return &o, nil
}
