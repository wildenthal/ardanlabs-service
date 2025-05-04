VERSION = 0.0.1
IMAGE = ardanlabs/app:$(VERSION)
DOCKER_IMAGE = localhost/$(IMAGE)
NAMESPACE = app-system
APP = app

run:
	go run cmd/app/main.go | go run pkg/logfmt/main.go

up:
	kind create cluster \
		--image kindest/node:v1.32.0 \
		--name ardan-starter-cluster \
		--config k8s/kind-config.yaml
	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner
	./k8s/kind-with-registry.sh

down:
	kind delete cluster --name ardan-starter-cluster

build:
	podman build \
		-f docker/dockerfile \
		-t $(DOCKER_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
		.

load:
	podman image push $(DOCKER_IMAGE) localhost:5001/$(IMAGE)

apply-dev:
	kustomize build k8s/dev | kubectl apply -f -

restart:
	kubectl rollout restart deployment app --namespace=app-system

update: build load restart

metrics:
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"
