package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
)

func getPyxisEvents(p *Process, params database.GetPyxisEventsForDeviceByDateRangeParams) ([]database.PyxisEventResponse, error) {
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

func getMarActions(p *Process, params database.GetMarAdminActionsByPatientDayMedIDsParams) ([]database.MarActionResponse, error) {
	key := "GetMarAdminActionsByPatientDayMedIDs"
	key += params.Date.Format("2006-01-02 1504")

	for _, deptID := range params.DeptIDs {
		key += deptID
	}

	key += params.Mrn

	for _, medID := range params.MedIDs {
		key += medID
	}

	itemID, logErr := p.erxItemIdLinks.GetItemId(params.MedIDs[0])
	if logErr != nil {
		p.logger.LogInfo(fmt.Sprintf("Query Started. Getting MAR actions for MRN %s in department %s on %s for unknown itemID",
			params.Mrn,
			params.DeptIDs,
			params.Date.Format("2006-01-02")))

	} else {
		p.logger.LogInfo(fmt.Sprintf("Query Started. Getting MAR actions for MRN %s in department %s on %s for itemID %s",
			params.Mrn,
			params.DeptIDs,
			params.Date.Format("2006-01-02"),
			itemID))

	}

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		actions := []database.MarActionResponse{}
		err := json.Unmarshal(data, &actions)
		if err != nil {
			return nil, err
		}
		return actions, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	actions, err := p.dbq.GetMarAdminActionsByPatientDayMedIDs(context.Background(), params)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return nil, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&actions)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling MAR actions: %s", err.Error()))
		return nil, err
	}
	p.cache.Add(key, data)

	return actions, nil

}
