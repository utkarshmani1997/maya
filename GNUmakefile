# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Specify the name for the binaries
MAYACTL=maya
APISERVER=maya-apiserver

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: test mayactl apiserver

dev: format
	@MAYACTL=${MAYACTL} MAYA_DEV=1 sh -c "'$(PWD)/buildscripts/ctl/build.sh'"

mayactl:
	@echo "----------------------------"
	@echo "--> maya                    "
	@echo "----------------------------"
	@MAYACTL=${MAYACTL} sh -c "'$(PWD)/buildscripts/ctl/build.sh'"

initialize: bootstrap

deps:
	dep ensure

clean:
	rm -rf bin
	rm -rf ${GOPATH}/bin/${MAYACTL}
	rm -rf ${GOPATH}/bin/${APISERVER}
	rm -rf ${GOPATH}/pkg/*

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "--> Running go tool vet ..."
	@go tool vet $(VETARGS) ${GOFILES_NOVENDOR} ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "[LINT] Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
	fi

	@git grep -n `echo "log"".Print"` | grep -v 'vendor/' ; if [ $$? -eq 0 ]; then \
		echo "[LINT] Found "log"".Printf" calls. These should use Maya's logger instead."; \
	fi

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get $$tool; \
	done

image:
	@cp bin/maya/${MAYACTL} buildscripts/ctl/
	@cd buildscripts/ctl && sudo docker build -t openebs/maya:ci --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/ctl/${MAYACTL}
	@sh buildscripts/ctl/push

# You might need to use sudo
install: bin/${MAYACTL}
	install -o root -g root -m 0755 ./bin/${MAYACTL} /usr/local/bin/${MAYACTL}

# Use this to build only the maya-agent.
maya-agent:
	GOOS=linux go build ./cmd/maya-agent

# Use this to build only the maya apiserver.
apiserver:
	@echo "----------------------------"
	@echo "--> maya-apiserver               "
	@echo "----------------------------"
	@CTLNAME=${APISERVER} sh -c "'$(PWD)/buildscripts/apiserver/build.sh'"

# Currently both mayactl & apiserver binaries are pushed into
# m-apiserver image. This is going to be decoupled soon.
apiserver-image: bin apiserver
	@echo "----------------------------"
	@echo "--> apiserver image         "
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	@cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t openebs/m-apiserver:ci --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/apiserver/${APISERVER}
	@rm buildscripts/apiserver/${MAYACTL}
	@sh buildscripts/apiserver/push

.PHONY: all bin cov integ test vet maya-agent test-nodep apiserver apiserver-image
