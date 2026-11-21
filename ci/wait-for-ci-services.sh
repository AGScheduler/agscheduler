#!/usr/bin/env bash

set -euo pipefail

ports=(3306 6379 27017 2379 9200 4150 5672 1883 9092)

for port in "${ports[@]}"; do
  ./ci/wait-for-it.sh "127.0.0.1:$port" -s -t 120 >/dev/null
done

wait_for_command() {
  local description=$1
  shift
  local attempts=60
  for ((attempt = 1; attempt <= attempts; attempt++)); do
    if "$@" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for $description" >&2
  return 1
}

wait_for_command "MySQL" \
  docker compose -f ci/docker-compose.store.ci.yml exec -T mysql \
  mysqladmin ping -h 127.0.0.1 -uroot -p123456 --silent

wait_for_command "Redis" \
  docker compose -f ci/docker-compose.store.ci.yml exec -T redis redis-cli ping

wait_for_command "Elasticsearch" \
  curl -fsS "http://127.0.0.1:9200/_cluster/health?wait_for_status=yellow&timeout=1s"

wait_for_command "RabbitMQ" \
  docker compose -f ci/docker-compose.queue.ci.yml exec -T rabbitmq \
  rabbitmq-diagnostics -q ping
