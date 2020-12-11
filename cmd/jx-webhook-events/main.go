package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/services"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	client            kubernetes.Interface
	namespace         string
	serviceName       string
	metricPath        string
	port              string
	totalWebHookCount int64
}

func main() {
	log.Logger().Infof("starting jx-webhook-events health checks")

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

func (o *Options) findErrors() ([]string, error) {
	kherrors := []string{}

	var err error
	if o.serviceName == "" {
		o.serviceName = "hook"
	}
	if o.port == "" {
		o.port = "2112"
	}
	namespace := o.namespace
	if namespace == "" {
		namespace, err = kubeclient.CurrentNamespace()
		if err != nil {
			return kherrors, errors.Wrapf(err, "failed to find current namespace")
		}
		if namespace == "" {
			namespace = "jx"
		}
	}

	var endpoints []string

	if os.Getenv("LOCAL_MODE") == "true" {
		endpoints, err = o.findIngressMetricEndpoint(namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find ingress endpoint")
		}
	} else {
		endpoints, err = o.findLocalMetricEndpoints(namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find local endpoints")
		}
	}
	if len(endpoints) == 0 {
		kherrors = append(kherrors, fmt.Sprintf("no webhook endpoints found with service name: %s in namespace %s", o.serviceName, namespace))
		return kherrors, nil
	}

	err = o.findMetrics(endpoints)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find metrics for endpoints %v", endpoints)
	}

	return o.processMetrics()
}

func (o *Options) findLocalMetricEndpoints(namespace string) ([]string, error) {
	eps, err := o.client.CoreV1().Endpoints(namespace).Get(context.TODO(), o.serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find endpoints %s in namespace %s", o.serviceName, namespace)
	}
	var answer []string
	for _, ep := range eps.Subsets {
		host := ""
		port := o.port
		for _, a := range ep.Addresses {
			if a.Hostname != "" {
				host = a.Hostname
				break
			}
			if a.IP != "" {
				host = a.IP
				break
			}
		}
		if port == "" {
			for _, p := range ep.Ports {
				if p.Port > 0 {
					port = strconv.Itoa(int(p.Port))
					break
				}
			}
		}
		if host != "" && port != "" {
			if o.metricPath != "" && !strings.HasPrefix(o.metricPath, "/") {
				o.metricPath = "/" + o.metricPath
			}
			answer = append(answer, fmt.Sprintf("http://%s:%s%s", host, port, o.metricPath))
		}
	}
	return answer, nil
}

func (o *Options) findIngressMetricEndpoint(namespace string) ([]string, error) {
	u, err := services.FindIngressURL(o.client, namespace, o.serviceName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find Ingress %s in namespace %s", o.serviceName, namespace)
	}
	if u == "" {
		return nil, nil
	}
	return []string{stringhelpers.UrlJoin(u, o.metricPath)}, nil
}

func (o *Options) processMetrics() ([]string, error) {
	if o.totalWebHookCount <= 0 {
		return []string{"no webhooks received by lighthouse"}, nil
	}
	return nil, nil
}

func (o *Options) findMetrics(endpoints []string) error {
	for _, e := range endpoints {
		log.Logger().Debugf("querying metrics from %s", e)

		client, err := api.NewClient(api.Config{
			Address: e,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to create client for %s", e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		u, err := url.Parse(e)
		if err != nil {
			return errors.Wrapf(err, "failed to parse url %s", e)
		}

		req := &http.Request{
			Method: http.MethodGet,
			URL:    u,
		}
		_, data, err := client.Do(ctx, req)
		if err != nil {
			return errors.Wrapf(err, "failed to invoke requests on %s", e)
		}

		d := expfmt.NewDecoder(bytes.NewReader(data), expfmt.FmtText)

		for {
			metric := &dto.MetricFamily{}
			err = d.Decode(metric)
			if err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrapf(err, "failed to parse metric")
			}
			if metric.Name != nil && *metric.Name == "lighthouse_webhook_counter" {
				var c int64
				for _, m := range metric.Metric {
					if m.Counter != nil && m.Counter.Value != nil {
						c += int64(*m.Counter.Value)
					}
				}
				log.Logger().Infof("endpoint %s has lighthouse_webhook_counter: %v", e, c)
				o.totalWebHookCount += c
				return nil
			}
		}
	}
	return nil
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
