APP=riuh
APP_EXECUTABLE=$(APP)

start: build
	out/$(APP_EXECUTABLE) start

copy-config:
	cp application.yml.sample application.yml

dep:
	go mod vendor

build:
	mkdir -p out/
	go build -o out/$(APP_EXECUTABLE)

test:
	go test ./... -v -cover

clean:
	rm -rf out
