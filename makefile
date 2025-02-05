all: runlocal
	
runlocal: build
	./run.sh

run-in-cluster: restart-deploy

podman-build:
	podman build -t registry.gitlab.com/itspacchu/zigram-images:aws-key-rotation-reconciler .

podman-push: podman-build
	podman push registry.gitlab.com/itspacchu/zigram-images:aws-key-rotation-reconciler

restart-deploy: podman-push
	kubectl rollout restart -n kube-system 	deploy aws-key-rotation-reconciler

build:
	echo "Building.."
	CGOENABLED=false go build -o bin/aws-key-rotation .

