default: up

help:
	@echo 'Management commands:'
	@echo
	@echo 'Usage:'
	@echo '    make run                     Start the service.'
	@echo '    make upload                  Upload program to S3.'
	@echo '    make merge/dev               Merge branch into dev.'
	@echo '    make test                    Run tests on a compiled project.'
	@echo '    make clean                   Clean project.'
	@echo '    make logs/dev                Print and follow dev server logs.'
	@echo '    make logs/prod               Print and follow prod server logs.'
	@echo '    make logs-svc/accountsvc        Print and follow user service logs.'
	@echo

build:
	docker-compose build

up:
	docker-compose up

up-fast:
	docker-sync start
	docker-compose -f docker-compose.yml -f docker-compose.docker-sync.yml up

down:
	docker-compose down --remove-orphans

test:
	docker-compose -f docker-compose.test.yml run test $(file)

test-shell:
	docker-compose -f docker-compose.test.yml run test bash

test-fast:
	docker-sync start
	docker-compose -f docker-compose.test.yml -f docker-compose.test.docker-sync.yml run test $(file)

test-shell-fast:
	docker-sync start
	docker-compose -f docker-compose.test.yml -f docker-compose.test.docker-sync.yml run test bash

debug:
	docker-compose -f docker-compose.yml -f docker-compose.debug.yml up

tunnel:
	@echo 'EKS env should be set properly'
	@echo 'This command requires kubefwd (https://github.com/txn2/kubefwd)'
	@echo
	sudo -E kubefwd svc -n authsvc -n accountsvc -n installedappsvc

print:
	@echo $(call args,default)

logs/%:
	$(eval ENV := $(subst logs/,,$@))
	kubectl logs -l app.kubernetes.io/instance=buzzscreen-api-$(ENV) -c buzzscreen-api -n=buzzscreen -f

logs-svc/%:
	$(eval SERVICE := $(subst logs-svc/,,$@))
	kubectl logs -l app.kubernetes.io/name=$(SERVICE) -c $(SERVICE) -n $(SERVICE) -f