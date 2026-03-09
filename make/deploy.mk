.PHONY: deploy deploy-api deploy-frontend

## deploy: deploy Go API to Fly.io and Next.js playground to Vercel.
##
## Usage:
##   make deploy
##
deploy: deploy-api deploy-frontend

## deploy-api: deploy Go API to Fly.io.
##
## Usage:
##   make deploy-api
##
deploy-api:
	@echo "Deploying API to Fly.io…"
	flyctl deploy --remote-only
	@echo "API deployed → https://drift-guard-api.fly.dev"

## deploy-frontend: deploy Next.js playground to Vercel.
##
## Usage:
##   make deploy-frontend
##
deploy-frontend:
	@echo "Deploying playground to Vercel…"
	cd playground && npx vercel --prod --yes
	@echo "Playground deployed → https://drift-guard-theta.vercel.app"
