proto:
	@bash ./scripts/genproto.sh

copyright:
	@bash ./scripts/copyright/update-copyright.sh

build:
	@go build