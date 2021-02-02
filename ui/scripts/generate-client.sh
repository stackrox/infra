#!/usr/bin/env bash

# Generates TypeScript client based on the Swagger 2.0 definitions
# Should be invoked from infra/ui (grandparent) directory

OPENAPI_GENERATOR_CLI_IMAGE_TAG="v5.0.0"
OPENAPI_GENERATOR_CLI_IMAGE="openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_CLI_IMAGE_TAG}"

CLIENT_DIR="src/generated/client"

# paths below are relative to the git root "infra"
SWAGGER_FILE="generated/api/v1/service.swagger.json" 
GENERATOR_OUTPUT_DIR="ui/${CLIENT_DIR}"

set -x

docker run --rm -v "${PWD}/..:/local" "${OPENAPI_GENERATOR_CLI_IMAGE}" generate \
  -i "/local/${SWAGGER_FILE}" \
  -g typescript-axios \
  --skip-validate-spec \
  -o "/local/${GENERATOR_OUTPUT_DIR}"

yarn prettier --ignore-path "!${CLIENT_DIR}" --write "${CLIENT_DIR}"
