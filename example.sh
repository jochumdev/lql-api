#!/bin/bash

set -ex

SERVER="http://localhost:8080"

# GET Hosts
curl -X POST -d '{"method": "GET", "table": "hosts", "columns": ["name", "address", "groups"]}' $SERVER/v1/raw

# GET Hosts with limit
curl -X POST -d '{"method": "GET", "table": "hosts", "columns": ["name", "address", "groups"], "limit": 3}' $SERVER/v1/raw

# host stats from the tactical_overview widget
curl -X POST -d '{"method": "GET", "table": "hosts", "query": [["Stats", "state >= 0"], ["Stats", "state > 0"], ["Stats", "scheduled_downtime_depth = 0"], ["StatsAnd", "2"], ["Stats", "state > 0"], ["Stats", "scheduled_downtime_depth = 0"], ["Stats", "acknowledged = 0"], ["StatsAnd", "3"], ["Stats", "host_staleness >= 1.5"], ["Stats", "host_scheduled_downtime_depth = 0"], ["StatsAnd", "2"]]}' $SERVER/v1/raw