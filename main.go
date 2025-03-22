package main

/*
	TODO:
	finish tui
		refactor 
		implement middle adds
		implement replaces
		implement menu
		make pretty
	create json config
	add autostart
	autoload which files are primed to be blocked
	add cli
		start as background process
	h / left + l / right changes ruleset in menus and in switch ruleset and in view ruleset
	ctrl + w to clear buffer
	add gui + systray
	add children of directories to archive	
	autofix windows paths
	port to linux
		run in the background
	add file encryption
	add backup folder for file encryption
	add icon
	push to github

	BUGS: 
	Removing view from field list crashes program
	Deleting last ruleset crashes prog
	OOB switching rulesets, deleting fields
*/
import (
	"fmt"
	"time"
	"os"
//	"path/filepath"
	"github.com/shirou/gopsutil/process"
	"strconv"
	"strings"
//	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"reflect"
)

//TUI State
type model struct {
	ci int 				//Current index (of field)
	cf int 				//Current field
	crs int 			//Current ruleset
	temprs int			//Used for showing live feed of rulesets
	rs int				//Used to indicate ruleset menu
	invalidinput bool		//Warns user of invalid input
	page string			//Categorizes each page
	cursor int 			//Cursor position for each option
	inputbuffer strings.Builder	//Input buffer for manipulating fields
	pagestrings [][]string 		//Strings for specific pages
	active int			//Toggle for lock/unlock loop
	log []string			//Lock / unlock / kill output
}


//Reset to default between page switches
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

type BlockUpdate time.Time
type LogClear time.Time

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
		s += m.Page2()
	
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
		s += m.Page2()	

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
type Rules struct {
	Active []int // 0 = false, 1 = true
	Times []int // lowertime[i] uppertime[i+1](DONT USE LEADING ZEROES)
	Days []int // UMTWRFS, int[7] 0 = false, 1 = true
	Overrides []int // [overrides, days] (NOT IMPLEMENTED)
	Timelimit []int // [time allowed, time between next allotment] (NOT IMPLEMENTED)
	Processes []string //processname.exe (taskmgr->details or top)
	Files []string //full path (C:/bin/program.exe) (NEED TO ADD AUTOFIX FOR WINDOWS BACKSLASH)
}

var Ruleset = []Rules { //data for rulesets (will separate into json Eventually)
	//Ruleset 0
	{
		Active: []int{1},
		Times: []int{0,1700,2100,2400},
		Days: []int{1,1,1,1,1,1,1},
		Overrides: []int{3, 30},
		Timelimit: []int{1, 2, 3, 4, 5},
		Processes: []string{"osu!.exe", "cum.exe"},
		Files: []string{"C:/Users/jtir/AppData/Local/osulazer/osu!.exe"},
	},
}

//Creating a deep copy of a preinitialized Rules struct
func initEmptyRuleset() Rules {
	er := Rules {
		Active: []int{0},
		Times: []int{0,2400},
		Days: []int{1,1,1,1,1,1,1},
		Overrides: []int{0, 0},
		Timelimit: []int{0, 0},
		Processes: []string{},
		Files: []string{},
	}
	return er
}

func (m *model) killProcesses(processnames []string) error {
	var err error

	//Get list of all running processes
	processes, err := process.Processes()
	if err != nil {	return err }
	
	//Kill all matches
	for _, p := range processes {
		n, err := p.Name()
		if err == nil {
			for i := range processnames{
				if strings.ToLower(n) == processnames[i]{	
					m.log = append(m.log, fmt.Sprintf("match: %v" + processnames[i]))
					p.Kill()
				}
			}
		} else { //Skips when sometimes program can't find process from PID
			m.log = append(m.log, fmt.Sprintf("Skipping: %v", p))
		}
	}	
	return err
}

func (m *model) renameFiles(file []string, lock bool) error {
	var err error
	for i := range file{
		if lock == true {
			err := os.Rename(file[i], file[i] + ".lck")
			if err != nil{ m.log = append(m.log, fmt.Sprintf("[already locked] ")) } //error usually thrown if file already modified
			m.log = append(m.log, fmt.Sprintf("locking: %v", file[i]))
		} else {
			err := os.Rename(file[i] + ".lck", file[i])
			if err != nil{ 
				m.log = append(m.log, fmt.Sprintf("[already unlocked] ")) 
			}
			m.log = append(m.log, fmt.Sprintf("unlocking: %v", file[i]))
		}
	}
	return err
}


func blockCheck(times []int, days []int) (bool, error) {
	var err error
	
	//Check day before continuing
	currentTime := time.Now()
	dotw := currentTime.Weekday()
	switch dotws := dotw.String(); dotws {
	case "Sunday":
		if days[0] == 0{ return false, err }	
	case "Monday":
		if days[1] == 0{ return false, err } 
	case "Tuesday":
		if days[2] == 0{ return false, err }
	case "Wednesday":
		if days[3] == 0{ return false, err }
	case "Thursday":
		if days[4] == 0{ return false, err }
	case "Friday":
		if days[5] == 0{ return false, err }
	case "Saturday":
		if days[6] == 0{ return false, err }
	}

	//Check time
	formattedTime := currentTime.Format("1504")
	intTime, _ := strconv.Atoi(formattedTime)
	for i := 1; i < len(times); i+=2{
		if intTime < times[i] && intTime > times[i-1]{
			return true, err
		}
	}
	return false, err
}

func main() {
	//init model
	m := model{
	//Strings for specific pages
	pagestrings: [][]string{ 
		//Main menu
		[]string{"Run blocker", "Options", "Exit"},
		//Options
		[]string{"Rulesets", "Active", "Times", "Days", "Overrides", "Timelimit", "Processes", "Files", "Return to menu"},
		//Modify ruleset
		[]string{"Add", "Switch", "Delete", "View", "Return to options"},
		//Modify fields
		[]string{"Add", "Modify", "Delete", "Return to options"},
		//Delete ruleset formatting
		[]string{"Active   ", "Times    ", "Days     ", "Overrides", "Timelimit", "Processes", "Files    "}, 
	},	
	ci: 1, cf:4, crs: 0, cursor: 0, temprs: 1, page: "menu", invalidinput: false}

	//start tui
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil { panic(err) }

/*
	//init
	status, err := blockCheck(Ruleset[0].Times, Ruleset[0].Days)	
	if err != nil{ panic(err) }
	laststatus := !status

	//loop
	for true {
		for { //i := range Ruleset{ //uninitialized rule structs break program
			//check if block is active	
			status, err := blockCheck(Ruleset[0].Times, Ruleset[0].Days)	
			if err != nil{ panic(err) }
			fmt.Println("Block: ", status)
			if status == true && laststatus == false {
				err := killProcess(Ruleset[0].Procs)
				if err != nil{ panic(err) }	
				if laststatus == false{
					err = renameFile(Ruleset[0].Files, true)
					if err != nil { fmt.Println(err) }
					//encryptFile()
					laststatus = true
				}
			} else if status == false && laststatus == true {
				err := renameFile(Ruleset[0].Files, false)
				if err != nil { fmt.Println(err) }
				laststatus = false
			}
		time.Sleep(time.Minute)
		}
	}
*/
}
