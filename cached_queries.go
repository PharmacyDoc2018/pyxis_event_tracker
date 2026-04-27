package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
)

func getPyxisEvents(p *ProcessState, params database.GetPyxisEventsForDeviceByDateRangeParams) ([]database.PyxisEventResponse, error) {
	key := "GetPyxisEventsForDeviceByDateRange"
	key += params.Device
	key += params.Start.Format("2006-01-02 15:04")
	key += params.End.Format("2006-01-02 15:04")

	p.logger.LogInfo(fmt.Sprintf("Query started. Getting Pyxis events for %s from %s to %s", params.Device, params.Start.Format("2006-01-02 1504"), params.End.Format("2006-01-02 1504")))

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		events := []database.PyxisEventResponse{}
		err := json.Unmarshal(data, &events)
		if err != nil {
			return nil, err
		}
		return events, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	events, err := p.dbq.GetPyxisEventsForDeviceByDateRange(context.Background(), params)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return nil, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&events)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling Pyxis events: %s", err.Error()))
		return nil, err
	}
	p.cache.Add(key, data)

	return events, nil

}
