module github.com/jenkins-x/jx-kh-check/cmd/jx-secrets

go 1.15

require (
	github.com/Comcast/kuberhealthy/v2 v2.2.1-0.20201008204401-47f4cf834e6e
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/jenkins-x/jx-api/v3 v3.0.1
	github.com/jenkins-x/jx-kube-client/v3 v3.0.1
	github.com/jenkins-x/jx-logging/v3 v3.0.2
	github.com/jenkins-x/jx-secret v0.0.170
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
)
