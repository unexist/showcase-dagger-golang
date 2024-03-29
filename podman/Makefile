# Config
PODNAME := showcase
PG_USER := postgres
PG_PASS := postgres

SSL_COUNTRY := DE
SSL_STATE := DE
SSL_LOCALITY := DE
SSL_ORGNAME := unexist.dev
SSL_ORGUNIT := showcase
SSL_FQDN := gitlab

GITLAB_MAIL := "admin@local"
GITLAB_PASS := "todo_showcase!42"
GITLAB_URL := "https://gitlab:443"

TAG_NAME := "custom-pip-runner"

# SSL
gen-ssl:
	@openssl req -newkey rsa:4096  -x509  -sha512  -days 365 -nodes \
		-out gitlab.crt -keyout gitlab.key \
		-addext "subjectAltName=DNS:gitlab" \
		-subj "/C=$(SSL_COUNTRY)/ST=$(SSL_STATE)/L=$(SSL_LOCALITY)/O=$(SSL_ORGNAME)/OU=$(SSL_ORGUNIT)/CN=$(SSL_FQDN)"

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
		-p 10022:22 -p 10080:80 -p 10443:443 \
		-p 4567:4567

pd-pod-rm:
	@podman pod rm -f $(PODNAME)

pd-pod-recreate: pd-pod-rm pd-pod-create

pd-gitlab: gen-ssl
	@podman run -dit --name gitlab --pod=$(PODNAME) \
		--memory=4096m --cpus=4 \
		-v ./gitlab.crt:/etc/gitlab/ssl/gitlab.crt \
		-v ./gitlab.key:/etc/gitlab/ssl/gitlab.key \
		-v ./gitlab.rb:/etc/gitlab/gitlab.rb \
		-e GITLAB_ROOT_EMAIL=$(GITLAB_MAIL) -e GITLAB_ROOT_PASSWORD=$(GITLAB_PASS) \
		docker.io/gitlab/gitlab-ce:latest

pd-runner:
	@podman run -dit --name runner --pod=$(PODNAME) \
		--memory=2048m --cpus=2 \
		-v ./gitlab.crt:/etc/gitlab-runner/certs/gitlab.crt \
		docker.io/gitlab/gitlab-runner:latest

pd-runner-podman-build:
	@podman build -t $(TAG_NAME) -f Containerfile \
		--build-arg=GITLAB_URL=$(GITLAB_URL) \
		--build-arg=PODNAME=$(PODNAME)

# TOKEN=abcd123 make pd-runner-podman
pd-runner-podman:
	#podman secret exists REGISTRATION_TOKEN && podman secret rm REGISTRATION_TOKEN || true
	#podman secret exists config.toml && podman secret rm config.toml || true
	@podman secret rm REGISTRATION_TOKEN || true
	@podman secret rm config.toml || true
	@podman volume exists pipglr-storage && podman volume rm pipglr-storage || true
	@podman volume exists pipglr-cache && podman volume rm pipglr-cache || true

	@echo "$(TOKEN)" | podman secret create REGISTRATION_TOKEN -
	@touch ./config.toml
	@podman container runlabel register $(TAG_NAME)

	@podman secret create config.toml ./config.toml

	@podman container runlabel setupstorage $(TAG_NAME)
	@podman container runlabel setupcache $(TAG_NAME)
	@podman container runlabel run $(TAG_NAME)

pd-postgres:
	@podman run -dit --name postgres --pod=$(PODNAME) \
		-e POSTGRES_USER=$(PG_USER) \
		-e POSTGRES_PASSWORD=$(PG_PASS) \
		docker.io/postgres:latest

pd-init: pd-machine-init pd-machine-start pd-pod-create

pd-start: pd-gitlab
