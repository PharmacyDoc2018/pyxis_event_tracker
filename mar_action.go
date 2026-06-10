package main

import "time"

type MarAction struct {
	SavedTime               time.Time //-- SAVED_TIME
	OrderNumber             string    //-- ORDER_MED_ID - convert to string
	MarAction               string    //-- FilteredMARAction
	DisplayName             string    //--DISPLAY_NAME
	MedicationID            string    //-- MEDICATION_ID convert to string
	UserID                  string    //--SYSTEM_LOGIN
	UserName                string    //-- NAME
	CalcDoseUnitDescription string    //-- CalcDoseUnitDescription
	CalcMinDose             float64   //-- CALC_MIN_DOSE
	MRN                     string    //-- PAT_MRN_ID
	PtName                  string    //-- PAT_NAME
}
