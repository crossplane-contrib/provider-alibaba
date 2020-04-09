module github.com/crossplane/provider-alibaba

go 1.14

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.109
	github.com/crossplane/crossplane v0.10.0-rc.0.20200410142608-84b1c08d1890
	github.com/crossplane/crossplane-runtime v0.7.0
	github.com/crossplane/crossplane-tools v0.0.0-20200303232609-b3831cbb446d
	github.com/pkg/errors v0.8.1
	golang.org/x/tools v0.0.0-20200410194907-79a7a3126eef // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.4
)
