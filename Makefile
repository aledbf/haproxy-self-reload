all: push

# 0.0.0 shouldn't clobber any released builds
TAG = 0.2
PREFIX = aledbf/haproxy-self-reload

binary: main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o haproxy-init

container: binary
	docker build -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	# remove haproxy images
	docker rmi -f $(PREFIX):$(TAG) || true
