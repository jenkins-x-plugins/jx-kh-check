package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

const (
	oauthToken  = "OAUTH_TOKEN"
	gitProvider = "GIT_PROVIDER"
)

type Options struct {
	httpClient *http.Client
}

func newOptions() (*Options, error) {
	return &Options{httpClient: &http.Client{}}, nil
}

func main() {

	if os.Getenv(gitProvider) == "" {
		log.Logger().Fatalf("%s not set", gitProvider)
		return
	}
	if os.Getenv(oauthToken) == "" {
		log.Logger().Fatalf("%s not set", oauthToken)
		return
	}

	log.Logger().Infof("starting jx-bot-token health checks")

	o, err := newOptions()
	if err != nil {
		log.Logger().Fatalf("failed to validate options: %v", err)
	}

	kherrors := o.findErrors()

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

func (o Options) findErrors() []string {

	var kherrors []string
	var checkURL string

	gitProvider := os.Getenv(gitProvider)
	oauthToken := os.Getenv(oauthToken)

	switch gitProvider {
	case "https://github.com":
		checkURL = "https://api.github.com/user"
	default:
		return append(kherrors, fmt.Sprintf("verifying bot token not yet supported for git provider %s", gitProvider))
	}

	req, err := http.NewRequestWithContext(context.TODO(), "GET", checkURL, nil)
	if err != nil {
		return append(kherrors, fmt.Sprintf("failed to create new request with URL: %s: %v", checkURL, err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", oauthToken))

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return append(kherrors, fmt.Sprintf("failed to reach URL: %s: %v", checkURL, err))
	}

	if resp == nil {
		return append(kherrors, fmt.Sprintf("no response from URL: %s", checkURL))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return append(kherrors, fmt.Sprintf("failed to verify bot account with %s, response status %s code %d", checkURL, resp.Status, resp.StatusCode))
	}

	log.Logger().Infof("received %d", resp.StatusCode)
	return kherrors
}
