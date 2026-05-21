package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ERxItemIdLinksFileName = "ERxItemIdLinks.json"

type ERxItemIdLink struct {
	ERx    string
	ItemID string
}

type ERxItemIdLinks struct {
	Map map[string]ERxItemIdLink
}

func (e *ERxItemIdLinks) Add(erx, itemId string) error {
	if _, okay := e.Map[erx]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. link for ERx %s already exists", erx),
			logMessage: fmt.Sprintf("Error. Link for ERx %s already exists.", erx),
		}
	}

	e.Map[erx] = ERxItemIdLink{
		ERx:    erx,
		ItemID: itemId,
	}

	return nil
}

func (p *ProcessState) saveERxItemIdLinks() error {
	p.logger.LogInfo("Saving ERxItemIdLinks")

	file, err := os.OpenFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to open %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}
	defer file.Close()

	data, err := json.Marshal(p.erxItemIdLinks)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling ERxItemIdLinks: %s", err.Error()))
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error writting marshalled ERxItemIdLinks to file %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}

	p.logger.LogInfo("ERxItemIdLinks saved")
	return nil

}

func (p *ProcessState) loadERxItemIdLinks() error {
	p.logger.LogInfo("Loading ERxItemIdLinks")

	data, err := os.ReadFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to read %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}

	err = json.Unmarshal(data, p.erxItemIdLinks)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error unmarshalling data from %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}

	p.logger.LogInfo("ERxItemIdLinks loaded successfully")
	return nil
}
