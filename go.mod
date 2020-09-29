module github.com/jenkins-x-plugins/jx-khcheck

go 1.14

require (
	github.com/Comcast/kuberhealthy/v2 v2.1.2
	github.com/jenkins-x/jx-logging v0.0.3
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/client-go v10.0.0+incompatible // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77

)
