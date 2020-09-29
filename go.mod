module github.com/crossplane/provider-alibaba

go 1.14

// TODO(negz): Do not merge.
replace (
	// https://github.com/crossplane/crossplane-runtime/pull/206
	github.com/crossplane/crossplane-runtime => github.com/negz/crossplane-runtime v0.0.0-20200929054832-abf875093883

	// https://github.com/crossplane/crossplane-tools/pull/24
	github.com/crossplane/crossplane-tools => github.com/negz/crossplane-tools v0.0.0-20200926065509-1c7266fde3b5
)

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.109
	github.com/crossplane/crossplane-runtime v0.9.1-0.20200924144923-240dbf0821e6
	github.com/crossplane/crossplane-tools v0.0.0-20200923030414-95b434323cd4
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	golang.org/x/tools v0.0.0-20200410194907-79a7a3126eef // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
