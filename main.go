package main

/*
	TODO:
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
	"github.com/shirou/gopsutil/process"
	"strconv"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
)

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
}

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

//Ruleset layout
type Rules struct {
	Active []int // 0 = false, 1 = true
	Times []int // lowertime[i] uppertime[i+1](DONT USE LEADING ZEROES)
	Days []int // UMTWRFS, int[7] 0 = false, 1 = true
	Overrides []int // [overrides, days] (NOT IMPLEMENTED)
	Timelimit []int // [time allowed, time between next allotment] (NOT IMPLEMENTED)
	Processes []string //processname.exe (taskmgr->details or top)
	Files []string //full path (C:/bin/program.exe) (NEED TO ADD AUTOFIX FOR WINDOWS BACKSLASH)
}

//Base ruleset object
var Ruleset = []Rules {
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


