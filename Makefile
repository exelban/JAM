.SILENT: ui
.PHONY: ui

ui:
	rm -rf ../backend/ui && cd frontend && yarn build && cp -r ./dist ../backend/ui