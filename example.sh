#!/bin/bash

# Example call: ./example.sh https://checkmk.myserver.com/example/lql-api "-u myuser:SuperSecretPw!#"

# Configuration
SERVER="http://localhost:8080"
if [ -n "$1" ]; then
    SERVER=$1
fi
ARGS=$2



# Execute
set -ex

CURL="/usr/bin/curl $2 -s -f"

# GET Hosts
# $CURL -X POST -d '{"method": "GET", "table": "hosts", "columns": ["name", "address"]}' $SERVER/v1/raw

# Same Request but with table endpoint
$CURL "$SERVER/v1/table/hosts?columns=name,address&limit=1"

# GET Hosts with limit
# $CURL -X POST -d '{"method": "GET", "table": "hosts", "columns": ["name", "address", "groups"], "limit": 3}' $SERVER/v1/raw

# Same Request but with table endpoint
$CURL "$SERVER/v1/table/hosts?columns=name&column=address&column=groups&limit=3"

# host stats from the tactical_overview widget
# $CURL -X POST -d '{"method": "GET", "table": "hosts", "query": [["Stats", "state >= 0"], ["Stats", "state > 0"], ["Stats", "scheduled_downtime_depth = 0"], ["StatsAnd", "2"], ["Stats", "state > 0"], ["Stats", "scheduled_downtime_depth = 0"], ["Stats", "acknowledged = 0"], ["StatsAnd", "3"], ["Stats", "host_staleness >= 1.5"], ["Stats", "host_scheduled_downtime_depth = 0"], ["StatsAnd", "2"]]}' $SERVER/v1/raw

# Tactical overview data :)
$CURL "$SERVER/v1/stats/tactical_overview"

# Services
$CURL "$SERVER/v1/table/services?limit=1&filter=service_unhandled"

# Services by hostname
$CURL "$SERVER/v1/table/services?filter=Filter%3A%20host_name%20%3D%20checkmk01.%2A"