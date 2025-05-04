VERSION = 0.0.1
IMAGE = ardanlabs/app:$(VERSION)
DOCKER_IMAGE = localhost/$(IMAGE)
NAMESPACE = app-system
APP = app

run:
	go run cmd/app/main.go | go run pkg/logfmt/main.go

dev-up:
	kind create cluster \
		--image kindest/node:v1.32.0 \
		--name ardan-starter-cluster \
		--config k8s/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	./k8s/kind-with-registry.sh

dev-down:
	kind delete cluster --name ardan-starter-cluster

dev-status-all:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-status:
	watch -n 2 kubectl get pods -o wide --all-namespaces

build:
	podman build \
		-f docker/dockerfile \
		-t $(DOCKER_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
		.
	podman image prune -f

dev-load:
	podman image push $(DOCKER_IMAGE) localhost:5001/$(IMAGE)

dev-apply:
	kustomize build k8s/dev | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --for=condition=Ready --timeout=120s

dev-restart:
	kubectl rollout restart deployment app --namespace=app-system
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --for=condition=Ready --timeout=120s

dev-update: build dev-load dev-restart

dev-update-apply: build dev-load dev-apply

dev-describe-deployment:
	kubectl describe deployment app --namespace=app-system

dev-describe-pod:
	kubectl describe pod -l app=app --namespace=app-system

dev-logs:
	kubectl logs -l app=app --namespace=app-system --all-containers=true -f --tail=100 --max-log-requests=6 | go run pkg/logfmt/main.go

dev-debug-shell:
	kubectl run -i --tty --rm debug --image=localhost:5001/$(IMAGE) --namespace=app-system -- /bin/sh

metrics:
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"
