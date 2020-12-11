package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestOptions_findErrors(t *testing.T) {
	// getting the current namespace is found from a local kube config file
	err := os.Setenv("KUBECONFIG", filepath.Join("test_data", "test-config"))
	assert.NoError(t, err)
	ns := "jx"

	fakeHandler := &FakeHttpServer{}
	server := httptest.NewServer(fakeHandler)
	defer server.Close()

	t.Logf("got server at %s\n", server.URL)

	u, err := url.Parse(server.URL)
	require.NoError(t, err, "failed to parse server URL %s", server.URL)
	hostname := u.Hostname()
	port := u.Port()
	require.NotEmpty(t, t, hostname, "no test server host")
	require.NotEmpty(t, t, port, "no test server port")
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err, "failed to parse test server port %s", port)

	tests := []struct {
		name    string
		results string
		wantErr bool
		want    []string
	}{
		{
			name:    "several-metrics",
			results: "# TYPE lighthouse_webhook_counter counter\nlighthouse_webhook_counter{event_type=\"ping\"} 6\nlighthouse_webhook_counter{event_type=\"pull_request\"} 16\n",
			want:    nil,
		},
		{
			name:    "no-metrics",
			results: "promhttp_metric_handler_requests_total{code=\"200\"} 38\n",
			want:    []string{"no webhooks received by lighthouse"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeHandler.results = tt.results
			ep := &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hook",
					Namespace: ns,
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								Hostname: hostname,
							},
						},
						Ports: []corev1.EndpointPort{
							{
								Port: int32(portNumber),
							},
						},
					},
				},
			}
			client := fake.NewSimpleClientset(ep)
			o := Options{
				client:    client,
				namespace: ns,
				port:      port,
			}
			got, err := o.findErrors()
			if tt.wantErr {
				require.Error(t, err, "should have error for %s", tt.name)
			} else {
				require.NoError(t, err, "should have error for %s", tt.name)

				assert.Equal(t, tt.want, got, "results for %s", tt.name)
			}
		})
	}
}

type FakeHttpServer struct {
	results string
}

func (f *FakeHttpServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(f.results))
}
