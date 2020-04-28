.PHONY: all
all: deps lint-check test build

###########
## Build ##
###########

.PHONY: deps
deps: package.json
	yarn install --frozen-lockfile

SOURCES := $(shell ls *.json) tailwind.config.js yarn.lock
SOURCES += $(shell find public)
SOURCES += $(shell find src -name '*.ts' -print)
SOURCES += $(shell find src -name '*.tsx' -print)
SOURCES += $(shell find src -name '*.js' -print)
SOURCES += $(shell find src -name '*.css' -print)

# builds a production app
.PHONY: build
build: deps $(SOURCES)
	@echo "+ $@"
	yarn build

# runs unit tests in a non-interactive mode (w/o watching)
.PHONY: test
test: deps $(SOURCES)
	@echo "+ $@"
	yarn test --watchAll=false


#################
## Development ##
#################

# checks that files respect linting / formatting rules (primary used in CI)
.PHONY: lint-check
lint-check: deps $(SOURCES)
	@echo "+ $@"
	yarn lint-check

# attempts to lint / format files where it's possible to do automatically (it'll change the files)
.PHONY: lint
lint: deps $(SOURCES)
	@echo "+ $@"
	yarn lint

# starts UI development server and opens browser window with UI
.PHONY: start
start-dev-server: deps
	@echo "+ $@"
	yarn start

# runs unit tests in an interactive watch mode
.PHONY: test-watch
test-watch: deps $(SOURCES)
	@echo "+ $@"
	yarn test

##########
## Misc ##
##########

# prints the list of files that this makefile considers as sources
.PHONY: printsrcs
printsrcs:
	@echo "+ $@"
	@echo "$(SOURCES)"
