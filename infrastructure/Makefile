# Config
PODNAME := showcase

SSL_COUNTRY := DE
SSL_STATE := DE
SSL_LOCALITY := DE
SSL_ORG_NAME := unexist.dev
SSL_ORG_UNIT := showcase
SSL_FQDN := gitlab

GITLAB_MAIL := "admin@local"
GITLAB_PASS := "todo_showcase!42"
GITLAB_URL := "gitlab:10443"
REGISTRY_URL := "localhost:4567"

RUNNER_IMAGE_NAME := "custom-pip-runner"
BUILDER_IMAGE_NAME := "custom-builder"

# Helper
--gen-ssl:
	@openssl req -newkey rsa:4096  -x509  -sha512  -days 365 -nodes \
		-out gitlab.crt -keyout gitlab.key \
		-addext "subjectAltName=DNS:gitlab" \
		-subj "/C=$(SSL_COUNTRY)/ST=$(SSL_STATE)/L=$(SSL_LOCALITY)/O=$(SSL_ORG_NAME)/OU=$(SSL_ORG_UNIT)/CN=$(SSL_FQDN)"

--guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set: $*=abc123 make $(MAKECMDGOALS)"; \
		exit 1; \
	fi

--check-token:
ifndef TOKEN
	$(error $$TOKEN must be set: TOKEN=abcd123 make $(MAKECMDGOALS))
endif

--get-token:
	@echo "# This requires Podman >=1.4"
	@echo "#podman secret exists ACCESS_TOKEN && podman secret rm ACCESS_TOKEN || true"

	@podman secret rm ACCESS_TOKEN || true
	@echo "-u $(GITLAB_MAIL):$(GITLAB_PASS)" | curl -s --insecure -X GET \
		-K- "https://$(GITLAB_URL)/jwt/auth?service=container_registry&scope=repository:$(BUILDER_IMAGE_NAME):pull,push" | jq .token | \
		podman secret create ACCESS_TOKEN -

list-repository:
	@curl -s --insecure -X GET -u "$(GITLAB_MAIL):$(GITLAB_PASS)" \
		"https://$(GITLAB_URL)/jwt/auth?service=container_registry&scope=registry:catalog:*" | jq .token | \
		xargs -i curl -s --insecure --header "Authorization: Bearer {}" "https://$(REGISTRY_URL)/v2/_catalog" | jq .

dump-secret: --guard-SECRET
	@podman run --rm --secret $(SECRET) --name dummy-cat-container alpine cat /run/secrets/$(SECRET)

show-ssl:
	@openssl x509 -in gitlab.crt -text -noout

# Podman
pd-machine-init:
	@podman machine init --memory=8192 --cpus=4

pd-machine-start:
	@podman machine start

pd-machine-stop:
	@podman machine stop

pd-machine-rm:
	@podman machine rm

pd-machine-recreate: pd-machine-rm pd-machine-init pd-machine-start

pd-pod-create:
	@podman pod create -n $(PODNAME) --network bridge \
		-p 10022:22 -p 10080:80 -p 10443:10443 \
		-p 4567:4567

pd-pod-rm:
	@podman pod rm -f $(PODNAME)

pd-pod-recreate: pd-pod-rm pd-pod-create

pd-gitlab: --gen-ssl
	@podman run -dit --name gitlab --pod=$(PODNAME) \
		--memory=4096m --cpus=4 \
		-v ./gitlab.crt:/etc/gitlab/ssl/gitlab.crt \
		-v ./gitlab.key:/etc/gitlab/ssl/gitlab.key \
		-v ./gitlab.rb:/etc/gitlab/gitlab.rb \
		-v ./gitlab-data:/var/opt/gitlab \
		-e GITLAB_ROOT_EMAIL=$(GITLAB_MAIL) -e GITLAB_ROOT_PASSWORD=$(GITLAB_PASS) \
		docker.io/gitlab/gitlab-ce:latest

pd-runner:
	@podman run -dit --name runner --pod=$(PODNAME) \
		--memory=2048m --cpus=2 \
		-v ./gitlab.crt:/etc/gitlab-runner/certs/gitlab.crt \
		docker.io/gitlab/gitlab-runner:latest

pd-runner-podman-build:
	@podman build -t $(RUNNER_IMAGE_NAME) -f runner/Containerfile \
		--build-arg=GITLAB_URL=$(GITLAB_URL) \
		--build-arg=REGISTRY_URL=$(REGISTRY_URL) \
		--build-arg=PODNAME=$(PODNAME)

pd-runner-builder-build:
	@podman build -t $(BUILDER_IMAGE_NAME) -f builder/Containerfile

pd-gitlab-prepare-cache: #pd-runner-builder-build
	@podman login -u $(GITLAB_MAIL) -p $(GITLAB_PASS) --tls-verify=false https://$(REGISTRY_URL)

	# Runner builder
	@podman push --tls-verify=false \
		$(BUILDER_IMAGE_NAME):latest $(REGISTRY_URL)/root/showcase-dagger-golang/$(BUILDER_IMAGE_NAME):latest

	# Dagger builder
	@podman pull docker.io/library/alpine:latest
	@podman push --tls-verify=false \
		docker.io/library/alpine:latest $(REGISTRY_URL)/root/showcase-dagger-golang/alpine:latest

	# Dagger runner
	@podman pull docker.io/library/golang:latest
	@podman push --tls-verify=false \
		docker.io/library/golang:latest $(REGISTRY_URL)/root/showcase-dagger-golang/golang:latest

# TOKEN=abcd123 make pd-runner-podman
pd-runner-podman: --guard-TOKEN
	@echo "# This requires Podman >=1.4"
	@echo "#podman secret exists REGISTRATION_TOKEN && podman secret rm REGISTRATION_TOKEN || true"
	@echo "#podman secret exists config.toml && podman secret rm config.toml || true"

	@podman secret rm REGISTRATION_TOKEN || true
	@podman secret rm config.toml || true
	@podman volume exists pipglr-storage && podman volume rm pipglr-storage || true
	@podman volume exists pipglr-cache && podman volume rm pipglr-cache || true

	@echo "$(TOKEN)" | podman secret create REGISTRATION_TOKEN -
	@truncate -s 0 ./config.toml || true
	@podman container runlabel register $(RUNNER_IMAGE_NAME)

	@echo "# Fix SSL config to contact Gitlab registry"
	@sed -i -e 's#volumes = \["/cache"\]#volumes = ["/cache", "/etc/gitlab-runner/certs/gitlab.crt:/etc/pki/ca-trust/source/anchors/ca.crt:ro"]#' ./config.toml

	@podman secret create config.toml ./config.toml

	@podman container runlabel setupstorage $(RUNNER_IMAGE_NAME)
	@podman container runlabel setupcache $(RUNNER_IMAGE_NAME)
	@podman container runlabel run $(RUNNER_IMAGE_NAME)

pd-init: pd-machine-init pd-machine-start pd-pod-create

pd-start: pd-gitlab
