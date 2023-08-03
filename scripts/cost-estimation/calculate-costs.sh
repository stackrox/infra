#!/bin/bash

set +xeuo pipefail

WORKING_DIR=$(dirname "$0")

bq query \
    --nouse_legacy_sql \
    --project_id stackrox-infra \
    --format prettyjson \
< "${WORKING_DIR}/total-time-consumed.sql" \
| python3 "${WORKING_DIR}/render_costs.py"
