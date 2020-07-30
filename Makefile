# Copyright 2017 The Fission Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.DEFAULT_GOAL := check

check: test-run build clean

# run basic check scripts
test-run:
	hack/verify-gofmt.sh
	hack/verify-govet.sh
	hack/verify-staticcheck.sh
	hack/runtests.sh
	@rm -f coverage.txt

# ensure the changes are buildable
build:
	go build -o cmd/fission-bundle/fission-bundle ./cmd/fission-bundle/
	go build -o cmd/fission-cli/fission ./cmd/fission-cli/
	go build -o cmd/fetcher/fetcher ./cmd/fetcher/
	go build -o cmd/fetcher/builder ./cmd/builder/

# install CLI binary to $PATH
install: build
	mv cmd/fission-cli/fission $(GOPATH)/bin

# build images (environment images are not included)
image:
	docker build -t fission-bundle -f cmd/fission-bundle/Dockerfile.fission-bundle .
	docker build -t fetcher -f cmd/fetcher/Dockerfile.fission-fetcher .
	docker build -t builder -f cmd/builder/Dockerfile.fission-builder .

clean:
	@rm -f cmd/fission-bundle/fission-bundle
	@rm -f cmd/fission-cli/fission
	@rm -f cmd/fetcher/fetcher
	@rm -f cmd/fetcher/builder


kt-vpn:
	sudo ktctl --kubeconfig /home/jingtao/.kube/config --debug connect --method=socks5

kt-exchange_executor:
	 sudo ktctl --kubeconfig /home/jingtao/.kube/config --namespace fission exchange executor --expose 8888