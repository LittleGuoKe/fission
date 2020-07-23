FISSION_NAMESPACE="fission"
CONTEXT_NAME="minikube"
TEMPLATE_NAME="fission"

FISSION_IMAGE_NAME="jingtao/fission-bundle"
FISSION_IMAGE_TAG="2020-07-14-2"
FISSION_FETCHER_IMAGE_NAME="jingtao/fission-fetcher"
FISSION_FETCHER_IMAGE_TAG="2020-07-14-2"
FISSION_PREUPGRADECHECKS_IMAGE_NAME="jingtao/fission-preupgradechecks"
FISSION_PREUPGRADECHECKS_IMAGE_TAG="2020-07-14-2"

## fission 集群初始化
init_fission_cluster: minikube_start minikube_environment_init install_fission


## minikube的运维命令
minikube_start:
	https_proxy=http://hp.jingtao.fun:1080 \
		minikube start \
			--docker-env http_proxy=http://hp.jingtao.fun:1080 \
			--docker-env https_proxy=http://hp.jingtao.fun:1080 \
			--docker-env no_proxy=localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24 \
			--kubernetes-version v1.18.3 \
			--driver=virtualbox \
			--cpus=6 \
			--memory=8g \
			--disk-size=40g \
			--container-runtime=docker

minikube_dashboard:
	minikube dashboard

minikube_stop:
	minikube stop

minikube_delete:
	minikube delete


## minikube的基础设施环境
minikube_environment_init: install_localVolume_provisioner rancher_import export_docker_env

install_localVolume_provisioner:
	kubectl --context=$(CONTEXT_NAME) apply -f rancher-local-path-storage.yaml

rancher_import:
	kubectl apply -f https://rancher.huxiang.pro/v3/import/98qghf4j58vrk746qv24wp5b2v8bnz9f4pj8st67tlqzgknznqz7dn.yaml

export_docker_env:
	minikube -p minikube docker-env > .docker_env


## fission的安装
install_fission: # build_all
	kubectl --context=$(CONTEXT_NAME) create namespace $(FISSION_NAMESPACE)
	cd fission-all && https_proxy=http://hp.jingtao.fun:1080 helm dependency update && cd -
	https_proxy=http://hp.jingtao.fun:1080 \
	helm --kube-context $(CONTEXT_NAME) install --namespace $(FISSION_NAMESPACE) --name-template $(TEMPLATE_NAME) \
		 -f minikube-fission-values.yaml \
		 --set repository= \
		 --set image=$(FISSION_IMAGE_NAME),imageTag=$(FISSION_IMAGE_TAG) \
		 --set fetcher.image=$(FISSION_FETCHER_IMAGE_NAME),fetcher.imageTag=$(FISSION_FETCHER_IMAGE_TAG) \
		 --set preUpgradeChecksImage=$(FISSION_PREUPGRADECHECKS_IMAGE_NAME):$(FISSION_PREUPGRADECHECKS_IMAGE_TAG) \
		./fission-all
	kubectl --context=$(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo

update_fission_bundle_image: build_all
	helm --kube-context $(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) upgrade $(TEMPLATE_NAME) \
		-f minikube-fission-values.yaml \
		--set repository= \
		--set image=$(FISSION_IMAGE_NAME),imageTag=$(FISSION_IMAGE_TAG) \
		--set fetcher.image=$(FISSION_FETCHER_IMAGE_NAME),fetcher.imageTag=$(FISSION_FETCHER_IMAGE_TAG) \
		--set preUpgradeChecksImage=$(FISSION_PREUPGRADECHECKS_IMAGE_NAME):$(FISSION_PREUPGRADECHECKS_IMAGE_TAG) \
		./fission-all

## fission镜像管理
# build_fission_bundle build_fission_fetcher build_fission_cli build_fission_preupgradechecks
build_all: build_fission_bundle build_fission_fetcher build_fission_cli

build_fission_bundle: export_docker_env
	# docker environment from `minikube docker-env`
	bash .docker_env && \
		docker build -t $(FISSION_IMAGE_NAME):$(FISSION_IMAGE_TAG) -f ../cmd/fission-bundle/Dockerfile.fission-bundle ../

build_fission_fetcher: export_docker_env
	bash .docker_env && \
		docker build -t $(FISSION_FETCHER_IMAGE_NAME):$(FISSION_FETCHER_IMAGE_TAG) -f ../cmd/fetcher/Dockerfile.fission-fetcher ../

build_fission_preupgradechecks: export_docker_env
	bash .docker_env && \
		docker build -t $(FISSION_PREUPGRADECHECKS_IMAGE_NAME):$(FISSION_PREUPGRADECHECKS_IMAGE_TAG) -f ../cmd/preupgradechecks/Dockerfile.fission-preupgradechecks ../

build_fission_cli:
	cd ../cmd/fission-cli && go build -o ~/go/bin/fission


## 代理相关配置
proxy_all: proxy_fission_grafana proxy_fission_prometheus proxy_fission_pushgateway proxy_fission_router

proxy_fission_router:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/router 9080:80 --address 0.0.0.0 &
	echo "fission router map to 0.0.0.0:9080"

proxy_fission_grafana:
	echo "grafana admin password:"
	kubectl --context=$(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-grafana 3000 --address 0.0.0.0 &
	echo "fission grafana map to 0.0.0.0:3000"

proxy_fission_prometheus:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-server 9090:80 --address 0.0.0.0 &
	echo "fission prometheus map to 0.0.0.0:9090"

proxy_fission_pushgateway:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-pushgateway 9091 --address 0.0.0.0 &
	echo "fission pushgateway map to 0.0.0.0:9090"


## 卸载
uninstall_fission:
	helm --kube-context $(CONTEXT_NAME) uninstall $(TEMPLATE_NAME) --namespace $(FISSION_NAMESPACE)
	#kubectl --kube-context $(CONTEXT_NAME) delete namespace $(FISSION_NAMESPACE)

