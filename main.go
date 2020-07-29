package main

import (
	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/jenkins-x/jx-logging/pkg/log"
)

func main() {

	log.Logger().Infof("starting Jenkins X health checks")

	errors := []string{}
	// ingress

	// tls

	// dns

	var err error
	if len(errors) == 0 {
		err = checkclient.ReportSuccess()
		if err != nil {
			log.Logger().Fatalf("failed to report success status %v", err)
		}

	} else {
		err = checkclient.ReportFailure([]string{"a bad thing happened", "and another!"})
		if err != nil {
			log.Logger().Fatalf("failed to report failure status %v", err)
		}
	}

	log.Logger().Infof("successfully reported")
}
