.PHONY: default test test-cover dev

defalt: dev

# for dev
dev: export CONFIG=./configs
dev:
	fresh