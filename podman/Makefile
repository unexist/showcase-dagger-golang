PODNAME := showcase
PG_USER := postgres
PG_PASS := postgres

COUNTRY := DE
STATE := NRW
LOCALITY := NRW
ORGNAME := unexist.dev
ORGUNIT := showcase
FQDN := unexist.dev

# SSL
gen-ssl:
	openssl req -newkey rsa:4096  -x509  -sha512  -days 365 -nodes \
		-out gitlab.crt -keyout gitlab.key \
		-subj "/C=$(COUNTRY)/ST=$(STATE)/L=$(LOCALITY)/O=$(ORGNAME)/OU=$(ORGUNIT)/CN=$(FQDN)"

# Podman
pd-machine-init:
	podman machine init --memory=8192 --cpus=4

pd-machine-start:
	podman machine start

pd-machine-stop:
	podman machine stop

pd-machine-rm:
	@podman machine rm

pd-machine-recreate: pd-machine-rm pd-machine-init pd-machine-start

pd-pod-create:
	@podman pod create -n $(PODNAME) --network bridge \
		-p 10022:22 -p 10080:80 -p 10443:443 \
		-p 4567:4567

pd-pod-rm:
	podman pod rm -f $(PODNAME)

pd-pod-recreate: pd-pod-rm pd-pod-create

# gitlab:
# image: 'gitlab/gitlab-ce:latest'
# container_name: 'gitlab'
# restart: always
# ports:
#   - '22:22'
#   - '80:80'
#   - '443:443'
#   - '4567:4567'
# volumes:
#   - './gitlab.rb:/etc/gitlab/gitlab.rb'
#   - './ssl/gitlab.crt:/etc/gitlab/ssl/gitlab.crt'
#   - './ssl/gitlab.key:/etc/gitlab/ssl/gitlab.key'
pd-gitlab: #gen-ssl
	@podman run -dit --name gitlab --pod=$(PODNAME) \
		--memory=4096m --cpus=4 \
		-v ./gitlab.crt:/etc/gitlab/ssl/gitlab.crt \
		-v ./gitlab.key:/etc/gitlab/ssl/gitlab.key \
		-v ./gitlab.rb:/etc/gitlab/gitlab.rb \
		-e GITLAB_ROOT_EMAIL="root@local" -e GITLAB_ROOT_PASSWORD="todo_showcase!42" \
		docker.io/gitlab/gitlab-ce:latest

#  runner:
#    image: 'gitlab/gitlab-runner:latest'
#    container_name: 'runner'
#    restart: always
#    environment:
#      - DOCKER_HOST=tcp://dind:2375
#    volumes:
#      - './persistent-data/runner/config:/etc/gitlab-runner'
#      - './ssl/root.pem:/etc/ssl/certs/ca-certificates.crt'
pd-runner:
	@podman run -dit --name runner --pod=$(PODNAME) \
		--memory=2048m --cpus=2 \
		docker.io/gitlab/gitlab-runner:latest

pd-postgres:
	@podman run -dit --name postgres --pod=$(PODNAME) \
		-e POSTGRES_USER=$(PG_USER) \
		-e POSTGRES_PASSWORD=$(PG_PASS) \
		docker.io/postgres:latest

pd-init: pd-machine-init pd-machine-start pd-pod-create pd-images-build

pd-start: pd-gitlab pd-runner pd-dind
