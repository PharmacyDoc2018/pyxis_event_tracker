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

func (e *ERxItemIdLinks) Add(erx, itemId string) *logError {
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

func (e *ERxItemIdLinks) Remove(erx string) *logError {
	if _, okay := e.Map[erx]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("error. ERx %s not currently linked to an itemId", erx),
			logMessage: fmt.Sprintf("Error. ERx %s not currently linked to an itemId", erx),
		}
	}

	delete(e.Map, erx)
	return nil
}

func (e *ERxItemIdLinks) GetItemId(erx string) (string, *logError) {
	itemId, okay := e.Map[erx]
	if !okay {
		return "", &logError{
			errMessage: fmt.Sprintf("error. link for erx %s not found", erx),
			logMessage: fmt.Sprintf("Error. Link for ERx %s not found", erx),
		}
	}

	return itemId.ItemID, nil
}

func initERxItemIdLink() *ERxItemIdLinks {
	e := ERxItemIdLinks{}
	e.Map = map[string]ERxItemIdLink{}
	return &e
}

func (p *Process) saveERxItemIdLinks() error {
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

	return nil

}

func (p *Process) loadERxItemIdLinks() error {
	p.logger.LogInfo("Loading ERxItemIdLinks")

	data, err := os.ReadFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName))
	if err != nil {
		if err.Error() == "open data/ERxItemIdLinks.json: no such file or directory" {
			p.logger.LogInfo("ERxItemIdLinks.json not found. Attempting to create file")
			f, e := os.OpenFile("./data/ERxItemIdLinks.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if e != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to read %s: %s. Unable to create file: %s",
					ERxItemIdLinksFileName,
					err.Error(),
					e.Error()))
				return fmt.Errorf("%s - %s", err.Error(), e.Error())
			}

			_, e = f.WriteString("{}")
			if e != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to read %s: %s. File created. Unable to write to file: %s",
					ERxItemIdLinksFileName,
					err.Error(),
					e.Error()))
				return fmt.Errorf("%s - %s", err.Error(), e.Error())
			}
		} else {
			p.logger.LogError(fmt.Sprintf("Error. Unable to read %s: %s", ERxItemIdLinksFileName, err.Error()))
			return err

		}
		p.logger.LogInfo("ERxItemIdLinks.json file created")
		data = []byte("{}")
	}

	err = json.Unmarshal(data, p.erxItemIdLinks)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error unmarshalling data from %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}

	p.logger.LogInfo("ERxItemIdLinks loaded successfully")
	return nil
}
