#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

NS_S="vlink-s"
NS_C="vlink-c"
VETH_S="veth-s"
VETH_C="veth-c"

LINK_S_IP="192.168.100.1/30"
LINK_C_IP="192.168.100.2/30"

CFG_S="/tmp/vlink-s.yaml"
CFG_C="/tmp/vlink-c.yaml"
LOG_S="/tmp/vlink-s.log"
LOG_C="/tmp/vlink-c.log"

RENDEZVOUS_SECRET="${RENDEZVOUS_SECRET:-vlink-local-123}"

SERVER_PID=""
CLIENT_PID=""

cleanup() {
  set +e
  if [[ -n "${CLIENT_PID}" ]]; then sudo kill "${CLIENT_PID}" 2>/dev/null || true; fi
  if [[ -n "${SERVER_PID}" ]]; then sudo kill "${SERVER_PID}" 2>/dev/null || true; fi

  sudo ip netns del "${NS_S}" 2>/dev/null || true
  sudo ip netns del "${NS_C}" 2>/dev/null || true
  sudo ip link del "${VETH_S}" 2>/dev/null || true
}

trap cleanup INT TERM

cd "${ROOT_DIR}"

echo "==> Resetting configurations"
sudo rm -f "${CFG_S}" "${CFG_C}"

echo "==> Building vlink"
go build -o vlink ./cmd/vlink

echo "==> Resetting namespaces + veth (if any)"
sudo pkill -f "./vlink up" 2>/dev/null || true
cleanup

echo "==> Creating namespaces"
sudo ip netns add "${NS_S}"
sudo ip netns add "${NS_C}"

echo "==> Creating veth pair"
sudo ip link add "${VETH_S}" type veth peer name "${VETH_C}"
sudo ip link set "${VETH_S}" netns "${NS_S}"
sudo ip link set "${VETH_C}" netns "${NS_C}"

echo "==> Bringing up loopback + veth"
sudo ip netns exec "${NS_S}" ip link set lo up
sudo ip netns exec "${NS_C}" ip link set lo up

sudo ip netns exec "${NS_S}" ip addr add "${LINK_S_IP}" dev "${VETH_S}"
sudo ip netns exec "${NS_C}" ip addr add "${LINK_C_IP}" dev "${VETH_C}"

sudo ip netns exec "${NS_S}" ip link set "${VETH_S}" up
sudo ip netns exec "${NS_C}" ip link set "${VETH_C}" up

echo "==> Verifying veth connectivity"
sudo ip netns exec "${NS_S}" ping -c 1 -W 1 192.168.100.2 >/dev/null
sudo ip netns exec "${NS_C}" ping -c 1 -W 1 192.168.100.1 >/dev/null
echo "    OK: ${NS_S}<->${NS_C} link up"

echo "==> Initializing configs"
sudo ip netns exec "${NS_S}" ./vlink init server --secret "${RENDEZVOUS_SECRET}" --config "${CFG_S}" >/dev/null
sudo ip netns exec "${NS_C}" ./vlink init client --secret "${RENDEZVOUS_SECRET}" --config "${CFG_C}" >/dev/null

SERVER_ID="$(sudo ip netns exec "${NS_S}" ./vlink status server --config "${CFG_S}" | awk '/Peer ID:/{print $3}')"
CLIENT_ID="$(sudo ip netns exec "${NS_C}" ./vlink status client --config "${CFG_C}" | awk '/Peer ID:/{print $3}')"

if [[ -z "${SERVER_ID}" || -z "${CLIENT_ID}" ]]; then
  echo "ERROR: failed to read Peer IDs"
  exit 1
fi

echo "    Server PeerID: ${SERVER_ID}"
echo "    Client PeerID: ${CLIENT_ID}"

echo "==> Whitelisting peers"
sudo ip netns exec "${NS_S}" ./vlink peer add server "${CLIENT_ID}" --name "client" --config "${CFG_S}" >/dev/null || true
sudo ip netns exec "${NS_C}" ./vlink peer add client "${SERVER_ID}" --name "server" --config "${CFG_C}" >/dev/null || true

echo "==> Starting server + client"
sudo ip netns exec "${NS_S}" ./vlink up server --config "${CFG_S}" >"${LOG_S}" 2>&1 &
SERVER_PID="$!"
sudo ip netns exec "${NS_C}" ./vlink up client --config "${CFG_C}" >"${LOG_C}" 2>&1 &
CLIENT_PID="$!"

echo "    Logs: ${LOG_S} (server), ${LOG_C} (client)"

echo "==> Waiting for tunnel to become reachable"
deadline=$((SECONDS + 60))
while (( SECONDS < deadline )); do
  if sudo ip netns exec "${NS_C}" ping -c 1 -W 1 10.1.1.1 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! sudo ip netns exec "${NS_C}" ping -c 1 -W 1 10.1.1.1 >/dev/null 2>&1; then
  echo "ERROR: tunnel did not come up within 60s"
  echo "---- client log (tail) ----"
  tail -n 60 "${LOG_C}" || true
  echo "---- server log (tail) ----"
  tail -n 60 "${LOG_S}" || true
  exit 1
fi

echo "==> Testing cross-pings over tunnel"
sudo ip netns exec "${NS_C}" ping -c 3 -W 1 10.1.1.1
sudo ip netns exec "${NS_S}" ping -c 3 -W 1 10.1.1.2

echo "==> SUCCESS"
echo "    vlink is still running. Press Ctrl+C to stop and cleanup."
echo
echo "    Helpful commands to run manually:"
echo "      sudo ip netns exec ${NS_S} ip -br addr"
echo "      sudo ip netns exec ${NS_C} ip -br addr"
echo "      sudo ip netns exec ${NS_S} ping -c 3 10.1.1.2"
echo "      sudo ip netns exec ${NS_C} ping -c 3 10.1.1.1"
echo "      tail -f ${LOG_S}   # server logs"
echo "      tail -f ${LOG_C}   # client logs"

echo
echo "==> Waiting (Ctrl+C to stop)"
wait
