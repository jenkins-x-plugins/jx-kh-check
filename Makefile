
.PHONY: release
release: test

.PHONY: linux
linux:

.PHONY: goreleaser
linux:

.PHONY: test
test:
	cd cmd/jx-bot-token; make build test
	cd cmd/jx-install; make build test
	cd cmd/jx-pod-status; make build test
	cd cmd/jx-secrets; make build test
	cd cmd/jx-webhook-events; make build test
	cd cmd/jx-webhooks; make build test
