package main

import (
	"fmt"
	"reflect"
	"strings"
)



//Frontend for TUI
func (m model) View() string {
	var s string 

	//Get current Ruleset Value
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()
//	rst := reflect.TypeOf(Ruleset[m.crs])
	//Get field from current field
//	field := rsv.Field(m.cf)
	
	//Header
	s += "\nProgBlock\n\n"	

	switch m.page { 
	case "run":
		s += m.runPage()	

	//Main Menu
	case "menu":
		for i := range m.pagestrings[0]{ //CHANGE
			if i==m.cursor { 
				s += fmt.Sprintf("[x] %s\n", m.pagestrings[0][i])  
			} else { 
				s += fmt.Sprintf("[ ] %s\n", m.pagestrings[0][i]) 
			}
		}

	//Options
	case "options":
		s += fmt.Sprintf("Current ruleset: %v\n\n", m.crs+1)

		for i := range m.pagestrings[1]{ //CHANGE
			if i==m.cursor { 
				s += fmt.Sprintf("[x] %s\n", m.pagestrings[1][i]) 
			} else { 
				s += fmt.Sprintf("[ ] %s\n", m.pagestrings[1][i]) 
			}
		}

	//Field options	
	case "optfi":
		s += m.optionsPage()
	
	//Add to field
	case "addfi":
		s += m.listFieldValues()
		s += fmt.Sprintf("Value to add: %v\n", m.inputbuffer.String())
	
	//Select field to modify
	case "modfi":
		s += m.listFieldValues()
		s += fmt.Sprintf("Choose index: %v\n", m.inputbuffer.String())

	//Change field value
	case "changefi":
		s += m.listFieldValues()
		s += fmt.Sprintf("Change value: %v\n", m.inputbuffer.String())

	//Delete field
	case "delfi":
		s += m.listFieldValues()
		s += fmt.Sprintf("Index to delete: %v\n", m.inputbuffer.String())

	//Ruleset options
	case "optrs":
		s += m.optionsPage()

	//Switch to ruleset
	case "switchrs":
		temprsv := reflect.ValueOf(&Ruleset[m.temprs-1]).Elem()

		s += fmt.Sprintf("Current ruleset: %v\n", m.crs+1)
		s += fmt.Sprintf("Total rulesets: %v\n\n", len(Ruleset)) 
	
		//Show all fields of ruleset
		for i := 0; i < rsv.NumField(); i++{
			s += fmt.Sprintf("%v | %v\n", m.pagestrings[4][i], temprsv.Field(i))
		}

		s += fmt.Sprintf("\nSwitch to: %v\n", m.inputbuffer.String())

	//Delete ruleset
	case "delrs":
		s += m.listRulesetValues()	
		s += fmt.Sprintf("\nAre you sure you want to delete this ruleset?\n")
		s += fmt.Sprintf("(Y)es or (N)o")
	
	//View ruleset
	case "viewrs":
		s += m.listRulesetValues()
	
	}

	//Warn user of invalid input
	if m.invalidinput == true {
		s += fmt.Sprintf("Invalid input\n")
	}
	
	return s
}

func (m model) runPage() string {
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
		s += fmt.Sprintf("%v | %v\n", m.pagestrings[4][i], rsv.Field(i))
	}

	//Prevent crash with null access
	if len(m.log) > 0 { 
	s += fmt.Sprintf("\n%v", m.log[0])
	}
	return s
}

func (m *model) optionsPage() string {
	var s string 
	rst := reflect.TypeOf(Ruleset[m.crs])
	olen := len(m.pagestrings[2]) - 1
		
		//Modifying ruleset
		if m.rs == 1 {
		s += fmt.Sprintf("Current ruleset: %d\n", m.crs+1) 
		s += fmt.Sprintf("Total rulesets: %d\n\n", len(Ruleset))
			
			for i:=0; i < olen; i++ {
				if i==m.cursor { 
					s += fmt.Sprintf("[x] %v rulesets\n", m.pagestrings[2][i])
				} else { 
					s += fmt.Sprintf("[ ] %v rulesets\n", m.pagestrings[2][i]) 
				}
			}
				if m.cursor == olen { 
					s += fmt.Sprintf("[x] %v\n", m.pagestrings[2][olen])
				} else { 
					s += fmt.Sprintf("[ ] %v\n", m.pagestrings[2][olen]) 
				}

		} else { //Modifying fields
			rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()
			field := rsv.Field(m.cf)

			s += fmt.Sprintf("Current values: %v\n\n", field)

			fieldname := strings.ToLower(rst.Field(m.cf).Name)

			for i:=0; i < olen-1; i++ {
				if i==m.cursor { 
					s += fmt.Sprintf("[x] %v %v\n", m.pagestrings[3][i], fieldname)
				} else { 
					s += fmt.Sprintf("[ ] %v %v\n", m.pagestrings[3][i], fieldname)
				}
			}
				if m.cursor == olen-1 { 
					s += fmt.Sprintf("[x] %v\n", m.pagestrings[3][olen-1])
				} else { 
					s += fmt.Sprintf("[ ] %v\n", m.pagestrings[3][olen-1]) 
				}
		}
		return s
}

//Show all values of each field in current Ruleset
func (m *model) listRulesetValues() string {
	var s string
	
	//Get current Ruleset Value
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()

	s += fmt.Sprintf("Current ruleset: %v\n", m.crs+1)
	s += fmt.Sprintf("Total rulesets: %v\n\n", len(Ruleset))
		
//	Show all fields
	for i := 0; i < rsv.NumField(); i++{
		s += fmt.Sprintf("%v | %v\n", m.pagestrings[4][i], rsv.Field(i))
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
