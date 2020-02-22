gobuild:
	go build cmd/registry.go
image:
	docker build -f build/Dockerfile -t barelyuseless/simple-service-registry:latest .
