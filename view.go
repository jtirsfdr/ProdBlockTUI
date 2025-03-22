package main

import (
	"fmt"
	"reflect"
	"strings"
)

func (m *model) Page2() string {
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
