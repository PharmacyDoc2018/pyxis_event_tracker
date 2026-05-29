package main

type Department struct {
	ID   string
	Name string
}

type DepartmentCoverage struct {
	Map map[string]map[string]Department
	//-- map[PyxisName]map[DeptID]Department
}

func (d *DepartmentCoverage) GetCoveredDepartments(pyxisName string) ([]Department, *logError)
