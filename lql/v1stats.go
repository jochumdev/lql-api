package lql

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type V1StatsTacticalOverview struct {
	Hosts    *V1StatsTacticalOverviewEntry `json:"hosts" validate:"required" description:"Host stats"`
	Services *V1StatsTacticalOverviewEntry `json:"services" validate:"required" description:"Service stats"`
	Events   *V1StatsTacticalOverviewEntry `json:"events" validate:"required" description:"Event stats"`
}

type V1StatsTacticalOverviewEntry struct {
	All       float64 `json:"all" validate:"required" description:"all services/hosts"`
	Problems  float64 `json:"problems" validate:"require" description:"Num of problems"`
	Unhandled float64 `json:"unhandled" validate:"require" description:"Num of unhandled"`
	Stale     float64 `json:"stale" validate:"require" description:"Num of stale"`
}

func v1StatsGetTacticalOverview(c *gin.Context) (*V1StatsTacticalOverview, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")
	if client.IsAdmin(user) {
		user = ""
	}

	msg := `GET hosts
Stats: state >= 0
Stats: state > 0
Stats: scheduled_downtime_depth = 0
StatsAnd: 2
Stats: state > 0
Stats: scheduled_downtime_depth = 0
Stats: acknowledged = 0
StatsAnd: 3
Stats: host_staleness >= 1.5
Stats: host_scheduled_downtime_depth = 0
StatsAnd: 2`

	rsp, err := client.Request(c, msg, user, 0)
	if err != nil {
		return nil, err
	}

	if len(rsp) < 1 {
		return nil, errors.New("Received invalid host stats from socket")
	}

	host := &V1StatsTacticalOverviewEntry{}
	host.All = rsp[0]["stats_1"].(float64)
	host.Problems = rsp[0]["stats_2"].(float64)
	host.Unhandled = rsp[0]["stats_3"].(float64)
	host.Stale = rsp[0]["stats_4"].(float64)

	msg = `GET services
Stats: state >= 0
Stats: state > 0
Stats: scheduled_downtime_depth = 0
Stats: host_scheduled_downtime_depth = 0
Stats: host_state = 0
StatsAnd: 4
Stats: state > 0
Stats: scheduled_downtime_depth = 0
Stats: host_scheduled_downtime_depth = 0
Stats: acknowledged = 0
Stats: host_state = 0
StatsAnd: 5
Stats: service_staleness >= 1.5
Stats: host_scheduled_downtime_depth = 0
Stats: service_scheduled_downtime_depth = 0
StatsAnd: 3`

	rsp, err = client.Request(c, msg, user, 0)
	if rsp == nil {
		return nil, err
	}

	if len(rsp) < 1 {
		return nil, errors.New("Received invalid host stats from socket")
	}

	svc := &V1StatsTacticalOverviewEntry{}
	svc.All = rsp[0]["stats_1"].(float64)
	svc.Problems = rsp[0]["stats_2"].(float64)
	svc.Unhandled = rsp[0]["stats_3"].(float64)
	svc.Stale = rsp[0]["stats_4"].(float64)

	msg = `GET eventconsoleevents
Stats: event_phase = open
Stats: event_phase = ack
StatsOr: 2
Stats: event_phase = open
Stats: event_phase = ack
StatsOr: 2
Stats: event_state != 0
StatsAnd: 2
Stats: event_phase = open
Stats: event_state != 0
Stats: event_host_in_downtime != 1
StatsAnd: 3`

	rsp, err = client.Request(c, msg, user, 0)
	if err != nil {
		return nil, err
	}

	if len(rsp) < 1 {
		return nil, errors.New("Received invalid host stats from socket")
	}

	ev := &V1StatsTacticalOverviewEntry{}
	ev.All = rsp[0]["stats_1"].(float64)
	ev.Problems = rsp[0]["stats_2"].(float64)
	ev.Unhandled = rsp[0]["stats_3"].(float64)
	ev.Stale = 0

	return &V1StatsTacticalOverview{Hosts: host, Services: svc, Events: ev}, nil
}
