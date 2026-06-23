package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const departmentCoverageFileName = "DepartmentCoverage.json"

type Department struct {
	ID   string
	Name string
}

type DepartmentCoverage struct {
	Map map[string]map[string]Department
	//-- map[PyxisName]map[DeptID]Department
}

func (d *DepartmentCoverage) GetCoveredDepartments(pyxisName string) ([]Department, *logError) {
	results := []Department{}
	if _, okay := d.Map[pyxisName]; !okay {
		return nil, &logError{
			errMessage: fmt.Sprintf("error. no covered departments found for %s", pyxisName),
			logMessage: fmt.Sprintf("Error. No covered departments found for %s", pyxisName),
		}

	}

	for _, department := range d.Map[pyxisName] {
		results = append(results, department)
	}

	return results, nil
}

func (d *DepartmentCoverage) GetAllCoveredDepartments() ([]Department, *logError) {
	results := []Department{}
	tempMap := map[Department]struct{}{}

	for _, departmentMap := range d.Map {
		for _, department := range departmentMap {
			if _, okay := tempMap[department]; !okay {
				tempMap[department] = struct{}{}
			}
		}
	}

	for department := range tempMap {
		results = append(results, department)
	}

	if len(results) == 0 {
		return nil, &logError{
			errMessage: "error. no departments found",
			logMessage: "Error. No departments found",
		}
	}

	return results, nil
}

func (d *DepartmentCoverage) Add(pyxisName string, department Department) *logError {
	if _, okay := d.Map[pyxisName]; !okay {
		d.Map[pyxisName] = make(map[string]Department)
	}

	deptID := department.ID
	if _, okay := d.Map[pyxisName][deptID]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. %s already listed as covered by %s", department.Name, pyxisName),
			logMessage: fmt.Sprintf("Error. %s already listed as covered by %s", department.Name, pyxisName),
		}
	}

	d.Map[pyxisName][deptID] = department
	return nil
}

func (d *DepartmentCoverage) Remove(pyxisName string, department Department) *logError {
	if _, okay := d.Map[pyxisName]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("error. %s pyxis not found", pyxisName),
			logMessage: fmt.Sprintf("Error. %s pyxis not found", pyxisName),
		}
	}

	deptID := department.ID
	if _, okay := d.Map[pyxisName][deptID]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("error. %s not listed as covered by %s", department.Name, pyxisName),
			logMessage: fmt.Sprintf("Error. %s not listed as covered by %s", department.Name, pyxisName),
		}
	}

	delete(d.Map[pyxisName], deptID)
	return nil
}

func initDepartmentCoverage() *DepartmentCoverage {
	d := DepartmentCoverage{}
	d.Map = make(map[string]map[string]Department)
	return &d
}

func (p *Process) saveDepartmentCoverage() error {
	p.logger.LogInfo("Saving DepartmentCoverage")

	file, err := os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to open %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}
	defer file.Close()

	data, err := json.Marshal(p.departmentCoverage)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling DepartmentCoverage: %s", err.Error()))
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error writting marshalled DepartmentCoverage to file %s: %s", departmentCoverageFileName, err.Error()))
		return err
	}

	return nil

}

func (p *Process) loadDepartmentCoverage() error {
	p.logger.LogInfo("Loading DepartmentCoverage")

	//-- Check if save file exists. Create file if not found.
	if _, err := os.Stat(filepath.Join(p.pathToData, departmentCoverageFileName)); err != nil {
		if os.IsNotExist(err) {
			p.logger.LogInfo(fmt.Sprintf("%s not found. Attempting to create file", departmentCoverageFileName))
			file, err := os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to create DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create DepartmentCoverage save file: %s", err.Error())
			}
			_, err = file.WriteString("{}")
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to write DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write DepartmentCoverage save file: %s", err.Error())
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(p.pathToData, departmentCoverageFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, p.departmentCoverage)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err

	}

	p.logger.LogInfo("DepartmentCoverage loaded successfully")
	p.state.DepartmentCoverageSuccessful()
	return nil

}
