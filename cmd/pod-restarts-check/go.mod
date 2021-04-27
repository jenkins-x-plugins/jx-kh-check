module github.com/jenkins-x-plugins/jx-kh-check/cmd/pod-restarts-check

go 1.15

require (
	github.com/Comcast/kuberhealthy/v2 v2.2.1-0.20201008204401-47f4cf834e6e
	github.com/sirupsen/logrus v1.8.1
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
)
