#!/usr/bin/env bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
source "$ROOT/scripts/lib.sh"

set -euo pipefail

debug_grpc_resets() {
    if [[ "$#" -ne 1 ]]; then
        die "missing args. usage: debug_grpc_resets <pcap file>"
    fi
    local pcap_file="$1"

    mkdir -p bin
    # curl --fail -sL https://infra.rox.systems/v1/cli/linux/amd64/upgrade \
    #     | jq -r ".result.fileChunk" \
    #     | base64 -d \
    #     > bin/infractl
    # chmod +x bin/infractl
    make cli
    ln -s infractl-linux-amd64 bin/infractl

    bin/infractl --version

    sudo apt-get update
    sudo apt install -y tshark
    sudo tshark --version
    sudo tshark -D
    sudo tshark -i any -a duration:10 -w "$pcap_file" &
    pid="$!"

    # Give tshark a moment to connect
    sleep 2

    (
        export GRPC_GO_LOG_SEVERITY_LEVEL=info
        export GRPC_GO_LOG_VERBOSITY_LEVEL=99
        bin/infractl whoami
        bin/infractl list --all
    ) || touch "FAIL"

    # Let packet capture complete
    wait "$pid"

    sudo chmod 0666 "$pcap_file"

    [[ ! -f FAIL ]] || die "gRPC test failed"
}

debug_grpc_resets "$@"
