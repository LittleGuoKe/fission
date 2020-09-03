FISSION_NAMESPACE="fission"
KAFKA_NAMESPACE="development"
CONTEXT_NAME="minikube"
TEMPLATE_NAME="fission"
PROXY=http://hp.jingtao.fun:1080
TODAY=$(shell date "+%Y_%m_%d")

# docker environment from `minikube docker-env`
DOCKER_TLS_VERIFY="1"
DOCKER_HOST="tcp://192.168.39.6:2376"
DOCKER_CERT_PATH="/home/jingtao/.minikube/certs"
MINIKUBE_ACTIVE_DOCKERD="minikube"

DEFAULT_TAG="$(TODAY)-1"
FISSION_IMAGE_NAME="jingtao/fission-bundle"
FISSION_IMAGE_TAG=$(DEFAULT_TAG)  # preUpgradeChecksImage 也使用这个tag
FISSION_FETCHER_IMAGE_NAME="jingtao/fission-fetcher"
FISSION_FETCHER_IMAGE_TAG=$(DEFAULT_TAG)
FISSION_PREUPGRADECHECKS_IMAGE_NAME="jingtao/fission-preupgradechecks"
FISSION_PREUPGRADECHECKS_IMAGE_TAG=$(DEFAULT_TAG)
FISSION_FLUENTD_WRAPPER_IMAGE_NAME="jingtao/fission-fluentd-wrapper"
FISSION_FLUENTD_WRAPPER_IMAGE_TAG=$(DEFAULT_TAG)


## minikube的运维命令
minikube_start:
	HTTPS_PROXY=$(PROXY) HTTP_PROXY=$(PROXY) NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24,192.168.50.0/24 \
		minikube start \
			--profile $(CONTEXT_NAME) \
			--docker-env http_proxy=$(PROXY) \
			--docker-env https_proxy=$(PROXY) \
			--docker-env no_proxy=localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24,192.168.50.0/24 \
			--addons metrics-server \
			--addons storage-provisioner-gluster \
			--kubernetes-version v1.18.3 \
			--driver kvm2 \
			--cpus 8 \
			--memory 8g \
			--disk-size 40g \
			--container-runtime docker

minikube_dashboard:
	minikube dashboard --profile $(CONTEXT_NAME)

minikube_stop:
	minikube stop --profile $(CONTEXT_NAME)

minikube_delete:
	minikube delete --profile $(CONTEXT_NAME)


## fission 集群初始化
init_fission_cluster: minikube_environment_init install_fission fission_env_init


## minikube的基础设施环境
minikube_environment_init: install_localVolume_provisioner install_nfs_provisioner rancher_import install_kafka

install_localVolume_provisioner:
	kubectl --context $(CONTEXT_NAME) apply -f infrastructures/rancher-local-path-storage.yaml

install_nfs_provisioner:
	# https://github.com/helm/charts/tree/master/stable/nfs-client-provisioner
	kubectl --context $(CONTEXT_NAME) create namespace storage-nfs
	https_proxy=$(PROXY) \
		helm repo add stable https://kubernetes-charts.storage.googleapis.com && \
		helm repo update && \
		helm --kube-context $(CONTEXT_NAME) install --namespace storage-nfs --name-template helm-nfs-provisioner --set nfs.server=hp.jingtao.fun --set nfs.path=/k8s \
		--set nfs.mountOptions="{vers=4,port=2049}" \
		stable/nfs-client-provisioner

rancher_import:
	kubectl apply -f https://rancher.huxiang.pro/v3/import/qrthth9tbmjxnn4wdhg5jlhgmv6j4xf2drqkksskzq25xl7ht4d64x.yaml

install_kafka:
	kubectl --context $(CONTEXT_NAME) create namespace $(KAFKA_NAMESPACE)
	https_proxy=$(PROXY) helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
	https_proxy=$(PROXY) helm --kube-context $(CONTEXT_NAME) install --namespace $(KAFKA_NAMESPACE) -f infrastructures/kafka-values.yaml \
		--name-template my-kafka incubator/kafka


## fission的安装
install_fission: build_all
	kubectl --context $(CONTEXT_NAME) create namespace $(FISSION_NAMESPACE)
	cd fission-all && https_proxy=http://hp.jingtao.fun:1080 helm dependency update && cd -
	https_proxy=$(PROXY) \
		helm --kube-context $(CONTEXT_NAME) install --namespace $(FISSION_NAMESPACE) --name-template $(TEMPLATE_NAME) \
		 	-f minikube-fission-values.yaml \
		 	--set repository= \
		 	--set logger.fluentdImageRepository= \
		 	--set image=$(FISSION_IMAGE_NAME),imageTag=$(FISSION_IMAGE_TAG) \
		 	--set fetcher.image=$(FISSION_FETCHER_IMAGE_NAME),fetcher.imageTag=$(FISSION_FETCHER_IMAGE_TAG) \
		 	--set logger.fluentdImage=$(FISSION_FLUENTD_WRAPPER_IMAGE_NAME),logger.fluentdImageTag=$(FISSION_FLUENTD_WRAPPER_IMAGE_TAG) \
		 	--set preUpgradeChecksImage=$(FISSION_PREUPGRADECHECKS_IMAGE_NAME):$(FISSION_PREUPGRADECHECKS_IMAGE_TAG) \
			./fission-all
	kubectl --context $(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo

update_fission_image: build_all
	https_proxy=$(PROXY) \
		helm --kube-context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) upgrade $(TEMPLATE_NAME) \
			-f minikube-fission-values.yaml \
			--set repository= \
			--set logger.fluentdImageRepository= \
			--set image=$(FISSION_IMAGE_NAME),imageTag=$(FISSION_IMAGE_TAG) \
			--set fetcher.image=$(FISSION_FETCHER_IMAGE_NAME),fetcher.imageTag=$(FISSION_FETCHER_IMAGE_TAG) \
		 	--set logger.fluentdImage=$(FISSION_FLUENTD_WRAPPER_IMAGE_NAME),logger.fluentdImageTag=$(FISSION_FLUENTD_WRAPPER_IMAGE_TAG) \
			--set preUpgradeChecksImage=$(FISSION_PREUPGRADECHECKS_IMAGE_NAME) \
			./fission-all

## fission镜像管理
# build_fission_bundle build_fission_fetcher build_fission_cli build_fission_preupgradechecks
build_all: build_fission_bundle build_fission_fetcher build_fission_preupgradechecks build_fission_cli build_fluentd_wrapper

build_fission_bundle:
	DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
		docker build -t $(FISSION_IMAGE_NAME):$(FISSION_IMAGE_TAG) -f ../cmd/fission-bundle/Dockerfile.fission-bundle ../

build_fission_fetcher:
	DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
		docker build -t $(FISSION_FETCHER_IMAGE_NAME):$(FISSION_FETCHER_IMAGE_TAG) -f ../cmd/fetcher/Dockerfile.fission-fetcher ../

build_fission_preupgradechecks:
	DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
		docker build -t $(FISSION_PREUPGRADECHECKS_IMAGE_NAME):$(FISSION_PREUPGRADECHECKS_IMAGE_TAG) -f ../cmd/preupgradechecks/Dockerfile.fission-preupgradechecks ../

build_fluentd_wrapper:
		DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
    		docker build -t $(FISSION_FLUENTD_WRAPPER_IMAGE_NAME):$(FISSION_FLUENTD_WRAPPER_IMAGE_TAG) -f ../cmd/fluentd-wrapper/Dockerfile.fluentd-wrapper ../

build_fission_cli:
	cd ../cmd/fission-cli && go build -o ~/go/bin/fission


## fission 环境部署
fission_env_init:tag="$(TODAY)-1"
fission_env_init:
	cd /home/jingtao/Fuxing/fissions/fission-envs/env/python/python3-base && \
		DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
		docker build -t env/python:$(tag) . && \
		fission env create --name python --image env/python:$(tag) --builder fission/python-builder:latest --graceperiod 10

## fission 环境更新
fission_env_update:tag="$(TODAY)-1"
fission_env_update:
	cd /home/jingtao/Fuxing/fissions/fission-envs/env/python/python3-base && \
		DOCKER_TLS_VERIFY=$(DOCKER_TLS_VERIFY) DOCKER_HOST=$(DOCKER_HOST) DOCKER_CERT_PATH=$(DOCKER_CERT_PATH) MINIKUBE_ACTIVE_DOCKERD=$(MINIKUBE_ACTIVE_DOCKERD) \
		docker build -t env/python:$(tag) . && \
		fission env update --name python --image env/python:$(tag) --graceperiod 10

## 代理相关配置
proxy_all: proxy_fission_grafana proxy_fission_prometheus proxy_fission_pushgateway proxy_fission_router

proxy_fission_router:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/router 9080:80 --address 0.0.0.0 &
	echo "fission router map to 0.0.0.0:9080"

proxy_fission_grafana:
	echo "grafana admin password:"
	kubectl --context $(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
	kubectl --context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-grafana 3000 --address 0.0.0.0 &
	echo "fission grafana map to 0.0.0.0:3000"

proxy_fission_prometheus:
	kubectl --context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-server 9090:80 --address 0.0.0.0 &
	echo "fission prometheus map to 0.0.0.0:9090"

proxy_fission_pushgateway:
	kubectl --context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-pushgateway 9091 --address 0.0.0.0 &
	echo "fission pushgateway map to 0.0.0.0:9091"

proxy_fission_alertmanager:
	kubectl --context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-alertmanager 9093 --address 0.0.0.0 &
	echo "fission alertmanager map to 0.0.0.0:9093"

## 卸载
uninstall_fission:
	helm --kube-context $(CONTEXT_NAME) uninstall $(TEMPLATE_NAME) --namespace $(FISSION_NAMESPACE)
	#kubectl --kube-context $(CONTEXT_NAME) delete namespace $(FISSION_NAMESPACE)

