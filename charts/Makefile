FISSION_NAMESPACE="fission"
CONTEXT_NAME="minikube"
TEMPLATE_NAME="fission"

minikube_start:
	https_proxy=http://hp.jingtao.fun:1080 \
		minikube start \
			--docker-env http_proxy=http://hp.jingtao.fun:1080 \
			--docker-env https_proxy=http://hp.jingtao.fun:1080 \
			--docker-env no_proxy=192.168.99.0/24 \
			--kubernetes-version v1.18.3 \
			--driver=virtualbox \
			--cpus=6 \
			--memory=8g \
			--disk-size=100g \
			--container-runtime=docker

minikube_dashboard:
	minikube dashboard

minikube_stop:
	minikube stop

minikube_delete:
	minikube delete

minikube_environment_init: install_localVolume_provisioner


test: uninstall_fission install_fission

install_localVolume_provisioner:
	kubectl --context=$(CONTEXT_NAME) apply -f rancher-local-path-storage.yaml

install_fission:
	#kubectl --context=$(CONTEXT_NAME) create namespace $(FISSION_NAMESPACE)
	cd fission-all && https_proxy=http://hp.jingtao.fun:1080 helm dependency update && cd -
	https_proxy=http://hp.jingtao.fun:1080 \
	helm --kube-context $(CONTEXT_NAME) install --namespace $(FISSION_NAMESPACE) --name-template $(TEMPLATE_NAME) \
		 -f minikube-fission-values.yaml \
		./fission-all
	kubectl --context=$(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo


proxy_fission_grafana:
	kubectl --context=$(CONTEXT_NAME) get secret --namespace $(FISSION_NAMESPACE) fission-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward \
		$(shell kubectl --context=$(CONTEXT_NAME) get pods --namespace $(FISSION_NAMESPACE) -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=fission-grafana" -o jsonpath="{.items[0].metadata.name}") \
		3000 --address 0.0.0.0 &

proxy_fission_prometheus:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-server 9090:80 --address 0.0.0.0 &

proxy_fission_pushgateway:
	kubectl --context=$(CONTEXT_NAME) --namespace $(FISSION_NAMESPACE) port-forward service/fission-prometheus-pushgateway 9091 --address 0.0.0.0 &


uninstall_fission:
	helm --kube-context $(CONTEXT_NAME) uninstall $(TEMPLATE_NAME) --namespace $(FISSION_NAMESPACE)
	#kubectl --kube-context $(CONTEXT_NAME) delete namespace $(FISSION_NAMESPACE)

