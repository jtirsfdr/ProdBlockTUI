package main 

import (
	"time"
	"strconv"
	tea "github.com/charmbracelet/bubbletea"
	"reflect"
)

type BlockUpdate time.Time
type LogClear time.Time

//Reset to default settings between page switches
func (m *model) switchPage(page string) tea.Model {
	m.inputbuffer.Reset()
	m.invalidinput = false
	m.cursor = 0
	m.page = page
	return m
}

//Check inputs common amongst page types
func (m *model) checkMsgString(msg tea.KeyMsg) tea.KeyMsg {
	switch m.page{
	case "menu", "options", "optrs", "optfi":
		switch msg.String(){ 

		//Go down
		case "j":
			//Clamp cursor range
			switch m.page{
			case "menu":
				if m.cursor < 2 { m.cursor++ } 
			case "options":
				if m.cursor < 8 { m.cursor++ } 
			case "optrs":
				if m.cursor < 4 { m.cursor++ }
			case "optfi":
				if m.cursor < 3 { m.cursor++ }
			}

		//Go up	
		case "k":
			//Clamp cursor range
			if m.cursor > 0 { m.cursor-- }
		}

	case "addfi", "switchrs", "modfi", "changefi", "delfi":
		switch msg.String(){

		// Erase character in input buffer
		case "backspace": 
			s := m.inputbuffer.String()
			if len(s) > 0 { s = s[:len(s)-1] }
			m.inputbuffer.Reset()
			m.inputbuffer.WriteString(s)
		
		case "enter":
			//Pass

		case "esc":
			//Pass
		
		default: 
			//Add all other characters to input buffer
			m.inputbuffer.WriteString(msg.String())
		}
	}
	return msg
}

//Update block every minute
func refreshBlock() tea.Cmd {
	return tea.Tick(time.Second * 5, func(t time.Time) tea.Msg{		
		return BlockUpdate(t)
	})
}

//Erase 1 entry from log each cycle
func refreshLog() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg{		
		return LogClear(t)
	})
}


//Necessary method for TUI framework
func (m model) Init() tea.Cmd {
	return refreshLog()
}

//Backend TUI logic
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	//Get current Ruleset Value
	rsv := reflect.ValueOf(&Ruleset[m.crs]).Elem()

	//Get field Value
	field := rsv.Field(m.cf)

	switch msg := msg.(type){

	//Block update
	case BlockUpdate:
		for i := range Ruleset{
			if Ruleset[i].Active[0] == 1 {
				m.killProcesses(Ruleset[i].Processes)
				m.renameFiles(Ruleset[i].Files, true)
				//Add file to blocklist
			} else {
				m.renameFiles(Ruleset[i].Files, false)
				//Compare file with blocked file list
				//If match
				//Leave alone
				//Else
				//Unblock
			}	
		}
		//Restart lock/unlock timer
		if m.active == 1{ return m, refreshBlock() }

	case LogClear:
		//Prevents crashing on empty log slice
		switch{
		case len(m.log) > 1:
			m.log = m.log[1:len(m.log)]
		case len(m.log) == 1:
			m.log = []string{}
		}
		return m, refreshLog()

	case tea.KeyMsg:

		//Quit
		if msg.String() == "ctrl+c"{ return m, tea.Quit }

		//Check page
		switch m.page{

		//Main Menu
		case "menu":
			m.checkMsgString(msg)
			if msg.String() == "enter"{
				switch m.cursor{
					case 0:
						m.active = 1
						m.switchPage("run")
						//Start lock/unlock loop
						return m, refreshBlock() 
					case 1:
						m.switchPage("options")
					case 2:
						return m, tea.Quit
				}
			} else if msg.String() == "esc" {
				return m, tea.Quit
			}

		//Options
		case "options":
			m.checkMsgString(msg)
			if msg.String() == "enter"{
				if m.cursor == 0 {
					m.rs = 1
					m.switchPage("optrs")
				} else if m.cursor != 8 {
					m.cf = m.cursor - 1
					m.switchPage("optfi")
				} else {
					m.switchPage("menu")
				}
			} else if msg.String() == "esc"{
				m.switchPage("menu")
			}

		//Field options
		case "optfi":
			m.checkMsgString(msg)
			if msg.String() == "enter"{
				switch m.cursor{
				case 0:
					m.switchPage("addfi")	
				case 1:
					m.switchPage("modfi")
				case 2:
					m.switchPage("delfi")
				case 3: 
					m.switchPage("options")
				}
			} else if msg.String() == "esc" {
				m.rs = 0
				m.switchPage("options")
			}
	
		//Add to field
		case "addfi":
			//Process all but enter + escape
			m.checkMsgString(msg)

			if msg.String() == "enter"{
				//Convert field / check type
				switch val := field.Interface().(type) {
				case []string:
					//Add input buffer to field
					newval := append(val, m.inputbuffer.String())

					//Assign to field
					field.Set(reflect.ValueOf(newval))

					//Reset input
					m.inputbuffer.Reset()

				case []int:
					//Convert to int
					input, err := strconv.Atoi(m.inputbuffer.String())
					
					//Retry if not a number
					if err != nil {
						m.inputbuffer.Reset()
						m.invalidinput = true
					} else {
						//Add int to field
						newval := append(val, input)	
						
						//Assign to field
						field.Set(reflect.ValueOf(newval))

						//Reset input
						m.inputbuffer.Reset()
					}
				}
			} 

			if msg.String() == "esc" { m.switchPage("optfi") }

		//Selecting field to modify
		case "modfi":
			//Process all but enter + escape
			m.checkMsgString(msg) 

			//
			if msg.String() == "enter"{
				input, err := strconv.Atoi(m.inputbuffer.String())
				if err != nil || input > field.Len() { 
					m.inputbuffer.Reset()
					m.invalidinput = true
				} else {
					m.ci = input - 1
					m.switchPage("changefi")
				}
			} else if msg.String() == "esc" {
				m.switchPage("optfi")
			}
		
		//Changing field
		case "changefi":
			//Process all but enter + escape
			m.checkMsgString(msg) 
			
			if msg.String() == "enter"{
				//Convert / check type
				switch val := field.Interface().(type) {
				case []string:
					//Set field[index] value to input buffer
					val[m.ci] = m.inputbuffer.String()

					//Assign to field
					field.Set(reflect.ValueOf(val))

				case []int:
					//Convert input buffer to int
					input, err := strconv.Atoi(m.inputbuffer.String())

					//Retry if not a number
					if err != nil {
						m.inputbuffer.Reset()
						m.invalidinput = true
					} else { 
						//Set field index value to input buffer
						val[m.ci] = input

						//Assign to field
						field.Set(reflect.ValueOf(val))
						m.switchPage("modfi")
					}

				} 			
			} 
			if msg.String() == "esc" { m.switchPage("optfi") }

		case "delfi":
			//Process all but enter + escape
			m.checkMsgString(msg)
			if msg.String() == "enter"{
				//Convert input buffer to int
				input, err := strconv.Atoi(m.inputbuffer.String())

				//Retry if not a number or out of range
				if err != nil || input > field.Len() || input < 0 {
					m.inputbuffer.Reset()
					m.invalidinput = true
				} else {
					//Convert / check type
					switch val := field.Interface().(type){
					case []string:
						// Account for 0 index
						input = input - 1
						
						//Create slice skipping inputted index
						newval := append(val[:m.ci], val[m.ci+1:]...)
						
						//Assign value to field
						field.Set(reflect.ValueOf(newval))
	
						//Reset input buffer
						m.inputbuffer.Reset()
						m.invalidinput = false
					case []int:
						// Account for 0 index
						input = input - 1
						
						//Create slice skipping inputted index
						newval := append(val[:m.ci], val[m.ci+1:]...)
						
						//Assign value to field
						field.Set(reflect.ValueOf(newval))
	
						//Reset input buffer
						m.inputbuffer.Reset()
						m.invalidinput = false
					} 
				}
			}
			if msg.String() == "esc" {
				if m.rs == 1 {
					m.switchPage("optrs")
				} else {
					m.switchPage("optfi")
				}
			}

		case "optrs": //Ruleset options
			m.checkMsgString(msg)
			if msg.String() == "enter"{
				switch m.cursor{ 
				//Add ruleset (no page)
				case 0:
					//Deep copy (slices are references)
					er := initEmptyRuleset()
					//Add copy
					Ruleset = append(Ruleset, er)
				case 1: 
					m.switchPage("switchrs")
				case 2: 
					m.switchPage("delrs")
				case 3:
					m.switchPage("viewrs")
				case 4:
					m.rs = 0
					m.switchPage("options")
				}
			} else if msg.String() == "esc" {
				m.rs = 0
				m.switchPage("options")
			}

		case "switchrs": //Switch ruleset
			m.checkMsgString(msg)	

			input, err := strconv.Atoi(m.inputbuffer.String())
			if err == nil && input > 0 && input <= reflect.ValueOf(Ruleset).Len() { 
				m.temprs = input //live view of ruleset
			}

			if msg.String() == "enter"{
				//check if input is valid
				input, err := strconv.Atoi(m.inputbuffer.String())
				if err != nil { 
					m.inputbuffer.Reset()
					m.invalidinput = true
				} else if input <= len(Ruleset) && input > 0 {
					m.crs = input - 1 //0 index
					m.switchPage("optrs")
				} else {
					m.invalidinput = true
					m.inputbuffer.Reset()
				}
			} else if msg.String() == "esc" {
				m.switchPage("optrs")
			}

		case "delrs":
			switch msg.String(){
				case "Y", "y": 
					RulesetCopy := append([]Rules{}, Ruleset...)
					Ruleset = append(Ruleset[:m.crs], RulesetCopy[m.crs+1:]...)	
					m.switchPage("optrs")
				case "N", "n", "esc", "backspace":
					m.switchPage("optrs")
			} 

		case "viewrs": //View ruleset
			switch msg.String(){
			case "esc", "backspace":
				m.switchPage("optrs")
					//add left and right to scroll
			}

		case "run": //Run blocker
			switch msg.String(){
			case "esc", "backspace":
				m.active = 0
				m.switchPage("menu")
				m.log = []string{}
			
			//Show ruleset 
			case "1","2","3","4","5","6","7","8","9":
				input, _ := strconv.Atoi(msg.String())
				if input < reflect.ValueOf(Ruleset).Len() {
					m.crs = input - 1 //0 index
				}
			}
		}
	}
	return m, nil
}
