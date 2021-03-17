module github.com/crossplane/provider-alibaba

go 1.14

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.109
	github.com/aliyun/aliyun-log-go-sdk v0.1.19
	github.com/crossplane/crossplane-runtime v0.12.0
	github.com/crossplane/crossplane-tools v0.0.0-20201201125637-9ddc70edfd0d
	github.com/google/go-cmp v0.5.2
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.3.0
)
