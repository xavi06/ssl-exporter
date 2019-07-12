app := ssl-exporter
imagename := zj/golang/ssl-exporter
imagetag := 1.0.0
registry := jxregistry.docker.fccs.cn:28888
cluster := c1
namespace := ops

default: build

build:
	go build -o ~/go/bin/$(app)

docker:
	docker build -t $(imagename):$(imagetag) .

dockerrun:
	docker run -tid --rm --name $(app) $(imagename):$(imagetag) sh

dockercopy:
	docker cp $(app):/usr/bin/$(app) bin/linux/$(app)

dockerstop:
	docker stop $(app)

dockertag:
	docker tag $(imagename):$(imagetag) $(registry)/$(imagename):$(imagetag)

dockerpush:
	docker push $(registry)/$(imagename):$(imagetag)

linux: docker dockerrun dockercopy dockerstop
macos:
	go build -o bin/macos/$(app)

release: docker dockertag dockerpush

install:
	kubectl apply -f deploy/k8s
	#helm install --name $(app) --namespace $(namespace) --set image.tag=$(imagetag),cluster=$(cluster) deploy/$(app)

upgrade:
	kubectl apply -f deploy/k8s
	#helm upgrade $(app) --set image.tag=$(imagetag),cluster=$(cluster) deploy/$(app)
