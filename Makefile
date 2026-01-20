.PHONY: install upgrade remove benchmark

NAMESPACE ?= benchmark
RELEASE_NAME ?= benchmark
CHART_PATH = charts/benchmark
KUBECONFIG=~/.kube/yacloud-k3s.yaml

install:
	helm install $(RELEASE_NAME) $(CHART_PATH) -n $(NAMESPACE) --create-namespace --kubeconfig $(KUBECONFIG)

upgrade:
	helm upgrade $(RELEASE_NAME) $(CHART_PATH) -n $(NAMESPACE) \
		--atomic \
		--install \
		--cleanup-on-fail \
		--timeout 4m \
		--debug \
		--kubeconfig $(KUBECONFIG)

remove:
	helm uninstall $(RELEASE_NAME) -n $(NAMESPACE) --kubeconfig $(KUBECONFIG)

benchmark:
	sh benchmark.sh
