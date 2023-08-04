#!/bin/bash

WORKING_DIR=$(dirname "$0")

# Hack to avoid bq initialization spamming the stdout
touch "${HOME}/.bigqueryrc"

bq query \
    --nouse_legacy_sql \
    --project_id stackrox-infra \
    --format prettyjson \
< "${WORKING_DIR}/total-time-consumed.sql" \
| python3 "${WORKING_DIR}/render_costs.py"
