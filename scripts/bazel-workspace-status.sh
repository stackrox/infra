#!/bin/sh

# The URL for the CircleCI workflow that was run for the current commit.
if [ -n "$CIRCLE_WORKFLOW_ID" ]; then
    echo "STABLE_CIRCLECI_WORKFLOW_URL https://circleci.com/workflow-run/${CIRCLE_WORKFLOW_ID}"
fi

# The Git short SHA for the current commit.
echo "STABLE_GIT_SHORT_SHA $(git rev-parse --short HEAD)"

# The Git describe string for the current commit.
echo "STABLE_GIT_VERSION $(make --quiet tag)"
