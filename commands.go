package main

import (
	"fmt"
	"strconv"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
)

func (p *Process) setupCommands() {
	p.cliConfig.AddCommand("hello", func([]cli.CommandArg) error {
		fmt.Println("Hello, World!")
		p.logger.LogInfo("Command hello executed")
		return nil
	})

	p.cliConfig.AddCommand("exit", func([]cli.CommandArg) error {
		p.logger.LogInfo("Command exit executed")
		fmt.Println("closing... goodbye!")
		p.exit()
		return nil
	})

	p.cliConfig.AddCommand("echo", func(args []cli.CommandArg) error {
		if len(args) == 0 {
			return fmt.Errorf("error. need phrase argument")
		}

		if len(args) > 1 {
			return fmt.Errorf("error. too many arguments")
		}

		fmt.Println(args[0].Val)
		return nil
	}, cli.CommandArg{
		Name: "phrase",
	})

	p.cliConfig.AddCommand("add pyxis", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Add pyxis command executed")
		pyxisName := ""
		startDateString := ""

		for _, arg := range args {
			switch arg.Name {
			case "name":
				pyxisName = arg.Val

			case "start_date":
				startDateString = arg.Val
			}
		}

		startDate, err := parseDate(startDateString)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error. Unable to parse start date for new Pyxis event log: %s", err.Error()))
			return err
		}

		err = p.createNewPyxisEventLog(pyxisName, startDate)
		if err != nil {
			return err
		}

		if p.dbConnection {
			p.findMissingPyxisEvents()
		}

		return nil

	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	}, cli.CommandArg{
		Name:     "start_date",
		Required: true,
	})

	p.cliConfig.AddCommand("status", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Status command executed")
		mode := p.state.Mode()
		p.logger.LogInfo(fmt.Sprintf("Current mode: %d", mode))

		fmt.Println(p.state.GetState())
		return nil
	})

	//-- ERx - ItemID Link Commands:
	p.cliConfig.AddCommand("add ERxItemId link", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add ERxItemId link command executed")

		erx := ""
		itemId := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "itemid":
				itemId = arg.Val
			}
		}

		//-- Check for blank args
		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}
		if itemId == "" {
			p.logger.LogError("Command failed: itemid cannot be blank")
			return fmt.Errorf("error. itemid cannot be blank")
		}

		//-- Check for letters in args
		_, err := strconv.Atoi(erx)
		if err != nil {
			p.logger.LogError("Command failed: erx can only contain numbers")
			return fmt.Errorf("error. erx can only contain numbers")
		}
		_, err = strconv.Atoi(itemId)
		if err != nil {
			p.logger.LogError("Command failed: itemid can only contain numbers")
			return fmt.Errorf("error. itemid can only contain numbers")
		}

		//-- Call method to add link
		logErr := p.erxItemIdLinks.Add(erx, itemId)
		if logErr != nil {
			p.logger.LogError(logErr.LogError())
			return logErr
		}
		p.logger.LogInfo(fmt.Sprintf("Link created. ItemId %s now links to ERx %s", itemId, erx))
		fmt.Println("link created")
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "itemid",
		Required: true,
	})

	p.cliConfig.AddCommand("remove ERxItemId link", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Remove ERxItemId link command executed")

		erx := ""
		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val
			}
		}

		if erx == "" {
			err := fmt.Errorf("error. erx cannot be blank")
			p.logger.LogError("Command failed: " + err.Error())
			return err
		}

		itemId, logErr := p.erxItemIdLinks.GetItemId(erx)
		if logErr != nil {
			p.logger.LogError(logErr.logMessage)
			return logErr
		}

		logErr = p.erxItemIdLinks.Remove(erx)
		if logErr != nil {
			p.logger.LogError(logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("ERx %s unlinked from ItemId %s", erx, itemId))
		fmt.Printf("erx %s unlinked from itemid %s\n", erx, itemId)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	})

	//-- Department Coverage Commands:
	p.cliConfig.AddCommand("add department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add department coverage command executed")

		pyxisName := ""
		deptID := ""
		deptName := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxisName = arg.Val

			case "deptID":
				deptID = arg.Val

			case "dept_name":
				deptName = arg.Val
			}
		}

		if pyxisName == "" {
			err := fmt.Errorf("error. pyxis cannot be blank")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		if deptID == "" {
			err := fmt.Errorf("error. deptID cannot be blank")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		if !isNumeric(deptID) {
			err := fmt.Errorf("error. deptID must contain only numbers")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		if deptName == "" {
			err := fmt.Errorf("error. dept_name cannot be blank")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		department := Department{
			ID:   deptID,
			Name: deptName,
		}

		logErr := p.departmentCoverage.Add(pyxisName, department)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("Department %s added to %s's covered departments", deptName, pyxisName))
		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "deptID",
		Required: true,
	}, cli.CommandArg{
		Name: "dept_name",
	})

	p.cliConfig.AddCommand("remove department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove department coverage command executed")

		pyxis := ""
		deptID := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val

			case "deptID":
				deptID = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed: pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		if deptID == "" {
			p.logger.LogError("Command failed: deptID cannot be blank")
			return fmt.Errorf("error. deptID cannot be blank")
		}

		depts, logErr := p.departmentCoverage.GetCoveredDepartments(pyxis)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		deptToRemove := Department{}
		for _, dept := range depts {
			if dept.ID == deptID {
				deptToRemove = dept
				break
			}
		}

		logErr = p.departmentCoverage.Remove(pyxis, deptToRemove)
		if logErr != nil {
			p.logger.LogError("Command Failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("%s [%s] removed from %s as a covered department",
			deptToRemove.Name,
			deptToRemove.ID,
			pyxis))

		fmt.Printf("%s [%s] removed from %s as a covered department\n",
			deptToRemove.Name,
			deptToRemove.ID,
			pyxis)

		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "deptID",
		Required: true,
	})

	p.cliConfig.AddCommand("list department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list department coverage command executed")

		pyxis := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed: pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		depts, logError := p.departmentCoverage.GetCoveredDepartments(pyxis)
		if logError != nil {
			p.logger.LogError("Command failed: " + logError.logMessage)
			return logError
		}

		fmt.Printf("Departments covered by %s:\n", pyxis)
		for _, dept := range depts {
			fmt.Printf("%s [%s]\n", dept.Name, dept.ID)
		}

		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	})

	//-- ERX Commands:
	p.cliConfig.AddCommand("list erxs", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list ERXs command executed")

		erxs := p.erxs.GetAll()
		for _, erx := range erxs {
			fmt.Printf("%s [%s]\n", erx.DisplayName, erx.MedID)
		}

		return nil
	})

	p.cliConfig.AddCommand("add erx", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add erx command executed")

		erx := ""
		name := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "name":
				name = arg.Val
			}
		}

		if erx == "" {
			p.logger.LogError("Command Failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}
		if name == "" {
			p.logger.LogError("Command failed: name cannot be blank")
			return fmt.Errorf("error name cannot be blank")
		}

		logErr := p.erxs.Add(erx, name)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("%s [%s] added to ERXs", name, erx))
		fmt.Printf("%s [%s] added to ERXs\n", name, erx)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	})

	p.cliConfig.AddCommand("remove erx", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove erx command executed")

		erx := ""
		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val
			}
		}

		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}

		logErr := p.erxs.Remove(erx)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("erx %s removed", erx))
		fmt.Printf("erx %s removed\n", erx)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	})
}
