module github.com/stackrox/infra

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/argoproj/argo v2.4.3+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c
	google.golang.org/grpc v1.24.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190816222004-e3a6b8045b0b // indirect
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/colinmarc/hdfs => github.com/colinmarc/hdfs v1.1.4-0.20180805212432-9746310a4d31
	k8s.io/api => k8s.io/api v0.0.0-20190816222004-e3a6b8045b0b
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190704094733-8f6ac2502e51
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
)
