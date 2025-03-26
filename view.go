package main

import (
	"fmt"
	"reflect"
	"strings"
)


//Frontend for TUI
func (m model) View() string {
	var s string 
	//Header
	s += "\nProgBlock\n\n"	

	switch m.page { 
	case "run": //Run blocker
		s += m.viewRun()	

	case "menu": //Main Menu
		s += m.viewMenu()		

	case "options": //Options
		s += m.viewOptions()

	case "optfi", "optrs": //Field options	
		s += m.viewDataOptions()

	case "addfi": //Add to field
		s += m.listFieldValues()
		s += fmt.Sprintf("Value to add: %v\n", m.inputbuffer.String())

	case "modfi": //Select field to modify
		s += m.listFieldValues()
		s += fmt.Sprintf("Choose index: %v\n", m.inputbuffer.String())

	case "changefi": //Change field value
		s += m.listFieldValues()
		s += fmt.Sprintf("Change value: %v\n", m.inputbuffer.String())

	case "delfi": //Delete field
		s += m.listFieldValues()
		s += fmt.Sprintf("Index to delete: %v\n", m.inputbuffer.String())

	case "switchrs": //Switch to ruleset
		s += m.viewSwitch()

	case "delrs": //Delete ruleset
		s += m.listRulesetValues()	
		s += fmt.Sprintf("\nAre you sure you want to delete this ruleset?\n")
		s += fmt.Sprintf("(Y)es or (N)o")

	case "viewrs": //View ruleset
		s += m.listRulesetValues()
	}

	//Warn user of invalid input
	if m.invalidinput == true {
		s += fmt.Sprintf("Invalid input\n")
	}
	
	return s
}

// START VIEW PAGES

func (m model) viewMenu() string {
	var s string
	menulist := []string{"Run blocker", "Options", "Exit"}
		for i := range menulist { 
			if i==m.cursor { 
				s += fmt.Sprintf("[x] %s\n", menulist[i])  
			} else { 
				s += fmt.Sprintf("[ ] %s\n", menulist[i]) 
			}
		}
	return s
}

func (m model) viewOptions() string {
	var s string
	optionslist := []string{"Rulesets", "Active", "Times", "Days", "Overrides", "Timelimit", "Processes", "Files", "Return to menu"}	
		s += fmt.Sprintf("Current ruleset: %v\n\n", m.crs+1)
		for i := range optionslist{ 
			if i==m.cursor { 
				s += fmt.Sprintf("[x] %s\n", optionslist[i]) 
			} else { 
				s += fmt.Sprintf("[ ] %s\n", optionslist[i]) 
			}
		}
	return s
}

func (m model) viewSwitch() string {
	var s string
	switchstring := []string{"Active   ", "Times    ", "Days     ", "Overrides", "Timelimit", "Processes", "Files    "}
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()
	temprsv := reflect.ValueOf(&Ruleset[m.temprs-1]).Elem()

		s += fmt.Sprintf("Current ruleset: %v\n", m.crs+1)
		s += fmt.Sprintf("Total rulesets: %v\n\n", len(Ruleset)) 
	
		//Show all fields of ruleset
		for i := 0; i < rsv.NumField(); i++{
			s += fmt.Sprintf("%v | %v\n", switchstring[i], temprsv.Field(i))
		}

		s += fmt.Sprintf("\nSwitch to: %v\n", m.inputbuffer.String())
	return s
}

func (m model) viewRun() string {
	switchstring := []string{"Active   ", "Times    ", "Days     ", "Overrides", "Timelimit", "Processes", "Files    "}
	var s string
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()
	//Show active rulesets
	s += fmt.Sprintf("Rulesets active | ")
	for i:=0;i < reflect.ValueOf(Ruleset).Len();i++{
		if Ruleset[i].Active[0] == 1 { 
			s += fmt.Sprintf("%v | ", i+1)
		}
	}
	s += fmt.Sprintf("\n\n")
	//Show value of ruleset from input
	for i := 0; i < rsv.NumField(); i++{
		s += fmt.Sprintf("%v | %v\n", switchstring[i], rsv.Field(i))
	}

	//Prevent crash with null access
	if len(m.log) > 0 { 
	s += fmt.Sprintf("\n%v", m.log[0])
	}
	return s
}

func (m *model) viewDataOptions() string {
	var s string 
	modifystrings := []string{"Add", "Switch", "Delete", "View", "Return to options"}

	rst := reflect.TypeOf(Ruleset[m.crs])
	olen := len(modifystrings) - 1
		
		//Modifying ruleset
		if m.rs == 1 {
			s += fmt.Sprintf("Current ruleset: %d\n", m.crs+1) 
			s += fmt.Sprintf("Total rulesets: %d\n\n", len(Ruleset))
			
			for i:=0; i < olen; i++ {
				if i==m.cursor { 
					s += fmt.Sprintf("[x] %v rulesets\n", modifystrings[i])
				} else { 
					s += fmt.Sprintf("[ ] %v rulesets\n", modifystrings[i]) 
				}
			}
				if m.cursor == olen { 
					s += fmt.Sprintf("[x] %v\n", modifystrings[olen])
				} else { 
					s += fmt.Sprintf("[ ] %v\n", modifystrings[olen]) 
				}

		} else { //Modifying fields
			rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()
			field := rsv.Field(m.cf)

			s += fmt.Sprintf("Current values: %v\n\n", field)

			fieldname := strings.ToLower(rst.Field(m.cf).Name)

			for i:=0; i < olen-1; i++ {
				if i==m.cursor { 
					s += fmt.Sprintf("[x] %v %v\n", modifystrings[i], fieldname)
				} else { 
					s += fmt.Sprintf("[ ] %v %v\n", modifystrings[i], fieldname)
				}
			}
				if m.cursor == olen-1 { 
					s += fmt.Sprintf("[x] %v\n", modifystrings[olen-1])
				} else { 
					s += fmt.Sprintf("[ ] %v\n", modifystrings[olen-1]) 
				}
		}
		return s
}

// END VIEW PAGES

//Show all values of each field in current Ruleset
func (m *model) listRulesetValues() string {
	rulesetstrings := []string{"Add", "Switch", "Delete", "View", "Return to options"}

	var s string
	
	//Get current Ruleset Value
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()

	s += fmt.Sprintf("Current ruleset: %v\n", m.crs+1)
	s += fmt.Sprintf("Total rulesets: %v\n\n", len(Ruleset))
		
//	Show all fields
	for i := 0; i < rsv.NumField(); i++{
		s += fmt.Sprintf("%v | %v\n", rulesetstrings[i], rsv.Field(i))
	}	

	return s
}

//Show all values of field
func (m *model) listFieldValues() string {
	var s string
	
	//Get current Ruleset Value
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()

	s += fmt.Sprintf("Current values\n---\n")

	//Iterate through all field values
	switch v := rsv.Field(m.cf).Interface().(type) {
	case []int:
		for i := range v{
			s += fmt.Sprintf("%v | %v\n", i+1, v[i])
		}
	case []string:
		for i := range v{
			s += fmt.Sprintf("%v | %v\n", i+1, v[i])
		}
	}

	s += fmt.Sprintf("---\n")
	
	return s
}

