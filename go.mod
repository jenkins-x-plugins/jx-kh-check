module github.com/jenkins-x-plugins/jx-kh-check

go 1.15

require (
	github.com/Comcast/kuberhealthy/v2 v2.1.2
	github.com/jenkins-x/jx-api/v3 v3.0.1
	github.com/jenkins-x/jx-kube-client/v3 v3.0.1
	github.com/jenkins-x/jx-logging/v3 v3.0.2
	github.com/jenkins-x/jx-secret v0.0.170 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/metrics v0.19.2 // indirect
)
