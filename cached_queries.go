package main

import (
	"context"
	"encoding/json"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
)

func getPyxisEvents(p *ProcessState, params database.GetPyxisEventsForDeviceByDateRangeParams) ([]database.PyxisEventResponse, error) {
	key := "GetPyxisEventsForDeviceByDateRange"
	key += params.Device
	key += params.Start.Format("2006-01-02 15:04")
	key += params.End.Format("2006-01-02 15:04")

	data, okay := p.cache.Get(key)
	if okay {
		events := []database.PyxisEventResponse{}
		err := json.Unmarshal(data, &events)
		if err != nil {
			return nil, err
		}
		return events, nil
	}

	events, err := p.dbq.GetPyxisEventsForDeviceByDateRange(context.Background(), params)
	if err != nil {
		return nil, err
	}

	data, err = json.Marshal(&events)
	if err != nil {
		return nil, err
	}
	p.cache.Add(key, data)

	return events, nil

}
