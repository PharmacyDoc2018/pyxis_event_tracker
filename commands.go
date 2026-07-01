package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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

	//---------------- ERx - ItemID Link Commands ------------------------//
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
		p.logger.LogInfo("remove ERxItemId Link command executed")

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

		itemID, logErr := p.erxItemIdLinks.GetItemId(erx)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		logErr = p.erxItemIdLinks.Remove(erx)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("erx %s link removed from itemID %s", erx, itemID))
		fmt.Printf("erx %s link removed from itemID %s\n", erx, itemID)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	})

	p.cliConfig.AddCommand("list all ERxItemId links", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list all ERxItemId links command executed")

		itemIDs := p.erxItemIdLinks.GetAllItemIds()
		for _, itemID := range itemIDs {
			medIDs := p.erxItemIdLinks.GetMedIds(itemID)
			fmt.Printf("ERXs linked to itemID %s %s:\n", itemID, quickDisplayName(p.itemIDs.DisplayName, itemID))
			for _, medID := range medIDs {
				fmt.Printf("   %s [%s]\n", quickDisplayName(p.erxs.DisplayName, medID), medID)
			}
			fmt.Println()
		}

		return nil

	})

	//-------------------- Department Coverage Commands -------------------------//
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

	//------------------ ERX Commands ----------------------//
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

	//----------------------ItemID Commands -------------------------//
	p.cliConfig.AddCommand("add itemID", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add itemID command executed")

		itemID := ""
		name := ""

		for _, arg := range args {
			switch arg.Name {
			case "itemID":
				itemID = arg.Val

			case "name":
				name = arg.Val
			}
		}

		if itemID == "" {
			p.logger.LogError("Command failed: itemID cannot be blank")
			return fmt.Errorf("error. itemID cannot be blank")
		}

		if name == "" {
			p.logger.LogError("Command failed: name cannot be blank")
			return fmt.Errorf("error. name cannot be blank")
		}

		logErr := p.itemIDs.Add(itemID, name)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("ItemID %s %s added", itemID, name))
		fmt.Printf("itemID %s %s added\n", itemID, name)
		return nil

	}, cli.CommandArg{
		Name:     "itemID",
		Required: true,
	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	})

	p.cliConfig.AddCommand("remove itemID", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove itemID command executed")

		itemID := ""

		for _, arg := range args {
			switch arg.Name {
			case "itemID":
				itemID = arg.Val
			}
		}

		if itemID == "" {
			p.logger.LogError("Command failed: itemID cannot be blank")
			return fmt.Errorf("error. itemID cannot be blank")
		}

		logErr := p.itemIDs.Remove(itemID)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("ItemID %s removed", itemID))
		fmt.Printf("ItemID %s removed\n", itemID)
		return nil

	}, cli.CommandArg{
		Name:     "itemID",
		Required: true,
	})

	p.cliConfig.AddCommand("list itemIDs", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list itemIDs command executed")

		itemIDs := p.itemIDs.GetAll()
		for _, itemID := range itemIDs {
			fmt.Printf("%s %s\n", itemID.ID, itemID.DisplayName)
		}

		return nil
	})

	//-------------------- Control Event Trail Commands ------------------------//
	p.cliConfig.AddCommand("generate control trail", func(args []cli.CommandArg) error {
		p.logger.LogInfo("generate control trail command executed")

		type ReportRowValue struct {
			Key         string
			Type        string
			DateTime    string
			UserID      string
			UserName    string
			DisplayName string
			Amount      string
			Units       string
			MRN         string
			WitPtName   string
		}

		type RowNames struct {
			KeyName         string
			TypeName        string
			DateTimeName    string
			UserIDName      string
			UserNameName    string
			DisplayNameName string
			AmountName      string
			UnitsName       string
			MrnName         string
			WitPtName       string
		}

		eventRowNames := RowNames{
			KeyName:         "Item Transaction Key",
			TypeName:        "Transaction Type",
			DateTimeName:    "Tx Date Time",
			UserIDName:      "User ID",
			UserNameName:    "User Name",
			DisplayNameName: "Display Name",
			AmountName:      "Amount",
			UnitsName:       "Units",
			MrnName:         "MRN",
			WitPtName:       "Witness",
		}

		actionRowNames := RowNames{
			KeyName:         "Order Number",
			TypeName:        "MAR Action",
			DateTimeName:    "Saved Time",
			UserIDName:      "User ID",
			UserNameName:    "User Name",
			DisplayNameName: "Display Name",
			AmountName:      "Dose",
			UnitsName:       "Units",
			MrnName:         "MRN",
			WitPtName:       "Patient Name",
		}

		pyxis := ""
		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. Pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		for _, pyxisEventLog := range p.PyxisEventLogs {
			if pyxisEventLog.PyxisName == pyxis {
				controlTrailSlices := pyxisEventLog.ControlEventLog.GenerateTrailSlices()

				report := [][]string{}

				for _, controlTrailSlice := range controlTrailSlices {
					for _, controlTrail := range controlTrailSlice {
						batch := make([][]string, 10)
						for i := range batch {
							batch[i] = make([]string, len(controlTrail.Trail)*2)
						}
						for x, event := range controlTrail.Trail {
							switch event.Type {
							case pyxisEvent:
								batch[0][x*2] = eventRowNames.KeyName
								batch[0][(x*2)+1] = event.PyxisEvent.ItemTransactionKey.String()
								batch[1][x*2] = eventRowNames.TypeName
								batch[1][(x*2)+1] = event.PyxisEvent.TransactionType
								batch[2][x*2] = eventRowNames.DateTimeName
								batch[2][(x*2)+1] = event.PyxisEvent.TxDateTime.Format("2006-01-02")
								batch[3][x*2] = eventRowNames.UserIDName
								batch[3][(x*2)+1] = event.PyxisEvent.UserID
								batch[4][x*2] = eventRowNames.UserNameName
								batch[4][(x*2)+1] = event.PyxisEvent.UserName
								batch[5][x*2] = eventRowNames.DisplayNameName
								batch[5][(x*2)+1] = event.PyxisEvent.MedDisplayName
								batch[6][x*2] = eventRowNames.AmountName
								batch[6][(x*2)+1] = strconv.FormatFloat(event.PyxisEvent.AmountReferenced, 'f', -1, 64)
								batch[7][x*2] = eventRowNames.UnitsName
								batch[7][(x*2)+1] = event.PyxisEvent.AmountReferencedUnits
								batch[8][x*2] = eventRowNames.MrnName
								batch[8][(x*2)+1] = event.PyxisEvent.MRN
								batch[9][x*2] = eventRowNames.WitPtName
								batch[9][(x*2)+1] = event.PyxisEvent.WitnessName

							case marAction:
								batch[0][x*2] = actionRowNames.KeyName
								batch[0][(x*2)+1] = event.MarAction.OrderNumber
								batch[1][x*2] = actionRowNames.TypeName
								batch[1][(x*2)+1] = event.MarAction.MarAction
								batch[2][x*2] = actionRowNames.DateTimeName
								batch[2][(x*2)+1] = event.MarAction.SavedTime.Format("2006-01-02")
								batch[3][x*2] = actionRowNames.UserIDName
								batch[3][(x*2)+1] = event.MarAction.UserID
								batch[4][x*2] = actionRowNames.UserNameName
								batch[4][(x*2)+1] = event.MarAction.UserName
								batch[5][x*2] = actionRowNames.DisplayNameName
								batch[5][(x*2)+1] = event.MarAction.DisplayName
								batch[6][x*2] = actionRowNames.AmountName
								batch[6][(x*2)+1] = strconv.FormatFloat(event.MarAction.CalcMinDose, 'f', -1, 64)
								batch[7][x*2] = actionRowNames.UnitsName
								batch[7][(x*2)+1] = event.MarAction.CalcDoseUnitDescription
								batch[8][x*2] = actionRowNames.MrnName
								batch[8][(x*2)+1] = event.MarAction.MRN
								batch[9][x*2] = actionRowNames.WitPtName
								batch[9][(x*2)+1] = event.MarAction.PtName
							}
						}
						report = append(report, batch...)
					}
				}

				file, err := os.OpenFile(filepath.Join(p.pathToOut, pyxis+"_ControlEventTrails"+".csv"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					p.logger.LogError(fmt.Sprintf("Command failed: Error opening %s_ControlEventTrails.csv : %s", pyxis, err.Error()))
					return err
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				err = writer.WriteAll(report)
				if err != nil {
					p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
					fmt.Println(err.Error())
				}

				return nil
			}
		}

		p.logger.LogError(fmt.Sprintf("Command failed. %s pyxis not found", pyxis))
		return fmt.Errorf("error. %s pyxis not found", pyxis)

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	})
}
