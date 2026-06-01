package main

import "fmt"

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
	if _, okay := d.Map[pyxisName]; okay {
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

func (d *DepartmentCoverage) RemovePyxisDepartmentLink(pyxisName string, department Department) *logError {
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
