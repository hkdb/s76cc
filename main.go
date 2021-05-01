package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	// "path/filepath"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	scribble "github.com/nanobox-io/golang-scribble"
)

var mainwin *ui.Window

// User Env
var uname string
var homedir string
var dbdir string

// Counter Struct
type counter struct {
	Count int
}

type selected struct {
	Profile int
}

// Profile Struct
type profiler struct {
	Name  string
	Start string
	End   string
}

// Profile List
var count = 0
var profiles []profiler
var cindex int
var current string
var start string
var end string

// Get/Set User Configurations
func getUserEnv() {
	usr, err := user.Current() // Get Current User
	if err != nil {
		fmt.Print("Get User Error: ")
		fmt.Println(err)
	}
	uname = usr.Username
	homedir = usr.HomeDir
	dbdir = homedir + "/.s76cc"
	// Check if db Directory exists. If not, create the directory
	_, serr := os.Stat(dbdir)
	if serr != nil {
		initConfig()
	}

	loadConfig()
}

// Init config dir
func initConfig() {

	fmt.Println("No config directory... creating...")
	// Create directories
	os.MkdirAll(dbdir, os.ModePerm)
	os.MkdirAll(dbdir+"/count", os.ModePerm)
	os.MkdirAll(dbdir+"/profiles", os.ModePerm)
	os.MkdirAll(dbdir+"/current", os.ModePerm)

	// Add Default Full Charge Profile
	addProfile("Full Charge", "96", "100")
	setCurrent(0)

	// Set current settings
	current = "Full Charge"
	start = "96"
	end = "100"

	current, start, end = checkCurrent()

	if strings.Contains(current, "Custom") {
		// Add Custom Profile
		addProfile(current, start, end)
		setCurrent(1)
	}

	fmt.Println(current)
	fmt.Println(start)
	fmt.Println(end)

}

// Check current charge threshold settings
func checkCurrent() (string, string, string) {

	// check if it's already on custom
	cmd := exec.Command("sh", "-c", "system76-power charge-thresholds")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd_err := cmd.Run()
	if cmd_err != nil {
		fmt.Println(cmd_err)
	}
	fmt.Println(out.String())

	// read output
	cp := strings.Split(out.String(), "\n")
	if cp == nil {
		fmt.Println("No charge configs?")
	}
	n := cp[0][9:]
	s := cp[1][7:]
	e := cp[2][5:]

	fmt.Println("Name: ", n)
	fmt.Println("Start: ", s)
	fmt.Println("End: ", e)

	return n, s, e

}

// Change Profile
func changeProfile(ci int, s string, e string) bool {

	// check if it's already on custom
	cmd := exec.Command("sh", "-c", "system76-power charge-thresholds "+s+" "+e)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd_err := cmd.Run()
	if cmd_err != nil {
		fmt.Println(cmd_err)
		return false
	}
	fmt.Println(out.String())

	setCurrent(ci)

	return true
}

// Add custom profile
func addProfile(name string, s string, e string) {

	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	// initialize default profile
	dp := profiler{}
	dp.Name = name
	dp.Start = s
	dp.End = e

	if err := db.Write("profiles", strconv.Itoa(count), &dp); err != nil {
		fmt.Println("Error", err)
	}

	count++
	incrementCount()

}

// Refresh profiles
func refreshAddedProfile(name string, s string, e string) {

	// initialize default profile
	dp := profiler{}
	dp.Name = name
	dp.Start = s
	dp.End = e

	profiles = append(profiles, dp)
}

// Remove custom profile
func delProfile(ci string) {

	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	if err := db.Delete("profiles", ci); err != nil {
		fmt.Println("Error", err)
	}

	count--
	decreaseCount()

}

// Refresh deleted custom profile
func refreshDeletedProfile(ci string) {

	// Convert ci to int so to use it for adjusting slice
	i, cerr := strconv.Atoi(ci)
	if cerr != nil {
		fmt.Println(cerr)
	}

	// Instantiating eraser
	zero := profiler{}

	// Adjust slice
	copy(profiles[i:], profiles[i+1:])    // Shift profiles[ci+1:] left one index
	profiles[len(profiles)-1] = zero      // Erase last element (write zero value)
	profiles = profiles[:len(profiles)-1] // Truncate slice

}

// Increment count of profiles
func incrementCount() {
	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	initCount := counter{}
	initCount.Count = count

	if err := db.Write("count", "profiles", &initCount); err != nil {
		fmt.Println("Error", err)
	}
}

// Increment count of profiles
func decreaseCount() {
	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	initCount := counter{}
	initCount.Count = count

	if err := db.Write("count", "profiles", &initCount); err != nil {
		fmt.Println("Error", err)
	}
}

// Load profiles and current configs
func loadConfig() {

	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	records, err := db.ReadAll("profiles")
	if err != nil {
		fmt.Println("Error", err)
	}

	for _, f := range records {
		profileFound := profiler{}
		if err := json.Unmarshal([]byte(f), &profileFound); err != nil {
			fmt.Println("Unmarshaling Error", err)
		}
		profiles = append(profiles, profileFound)
	}

	fmt.Println(profiles)

	// Get count of profiles
	counter := counter{}
	if err := db.Read("count", "profiles", &counter); err != nil {
		fmt.Println("Error", err)
	}
	count = counter.Count

	fmt.Println("Count:", count)
	cindex = getCurrentIndex()
	fmt.Println("Count Index:", cindex)
	current, start, end = checkCurrent()

}

// Set currently selected profile for tracking purposes
func setCurrent(ci int) {

	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	// initialize default profile
	sp := selected{}
	sp.Profile = ci

	if err := db.Write("current", "profile", &sp); err != nil {
		fmt.Println("Error", err)
	}

	cindex = ci

}

func getCurrentIndex() int {

	// initialize default count
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	sp := selected{}

	if err := db.Read("current", "profile", &sp); err != nil {
		fmt.Println("Error", err)
	}

	return sp.Profile
}

func refreshStartEnd(ci int) bool {
	start = profiles[ci].Start
	end = profiles[ci].End
	if start == "" || end == "" {
		return false
	}
	return true
}

// Controls Page
func makeControlsPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	group := ui.NewGroup("Current Setting")
	group.SetMargined(true)
	vbox.Append(group, false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group.SetChild(ui.NewNonWrappingMultilineEntry())

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)
	group.SetChild(entryForm)

	cbox := ui.NewCombobox()
	for i := 0; i < len(profiles); i++ {
		cbox.Append(profiles[i].Name)
	}
	cbox.SetSelected(cindex)
	entryForm.Append("Profile :", cbox, false)

	startLabel := ui.NewLabel(start)
	entryForm.Append("Start (%) :", startLabel, true)

	endLabel := ui.NewLabel(end)
	entryForm.Append("End (%) :", endLabel, true)

	cbox.OnSelected(func(*ui.Combobox) {
		// Update the UI directly as it is called from the main thread
		startLabel.SetText(profiles[cbox.Selected()].Start)
		endLabel.SetText(profiles[cbox.Selected()].End)
	})

	cbutton := ui.NewButton("Change")
	cbutton.OnClicked(func(*ui.Button) {
		changeProfile(cbox.Selected(), profiles[cbox.Selected()].Start, profiles[cbox.Selected()].End)
	})
	entryForm.Append("", cbutton, false)

	rGroup := ui.NewGroup("Remove")
	rGroup.SetMargined(true)
	vbox.Append(rGroup, false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	rGroup.SetChild(ui.NewNonWrappingMultilineEntry())

	rForm := ui.NewForm()
	rForm.SetPadded(true)
	rGroup.SetChild(rForm)

	rbox := ui.NewCombobox()
	for i := 0; i < len(profiles); i++ {
		rbox.Append(profiles[i].Name)
	}
	rbox.SetSelected(cindex)
	rForm.Append("Profile :", rbox, false)

	startLabel2 := ui.NewLabel(start)
	rForm.Append("Start (%) :", startLabel2, true)

	endLabel2 := ui.NewLabel(end)
	rForm.Append("End (%) :", endLabel2, true)

	rbox.OnSelected(func(*ui.Combobox) {
		// Update the UI directly as it is called from the main thread
		startLabel.SetText(profiles[rbox.Selected()].Start)
		endLabel.SetText(profiles[rbox.Selected()].End)
	})

	rbutton := ui.NewButton("Remove")
	rbutton.OnClicked(func(*ui.Button) {
		delProfile(strconv.Itoa(rbox.Selected()))
		refreshDeletedProfile(strconv.Itoa(rbox.Selected()))
		ui.MsgBox(mainwin,
			"Profile Removed",
			"Please restart the application")

	})
	rForm.Append("", rbutton, false)

	aGroup := ui.NewGroup("Add")
	aGroup.SetMargined(true)
	vbox.Append(aGroup, false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	aGroup.SetChild(ui.NewNonWrappingMultilineEntry())

	aForm := ui.NewForm()
	aForm.SetPadded(true)
	aGroup.SetChild(aForm)

	n := ui.NewEntry()
	aForm.Append("Name :", n, false)
	sSlider := ui.NewSlider(0, 100)
	eSlider := ui.NewSlider(0, 100)
	aForm.Append("Start (%) :", sSlider, false)
	aForm.Append("End (%) :", eSlider, false)

	abutton := ui.NewButton("Add")
	abutton.OnClicked(func(*ui.Button) {
		newItem := &profiler{}
		newItem.Name = n.Text()
		newItem.Start = strconv.Itoa(sSlider.Value())
		newItem.End = strconv.Itoa(eSlider.Value())
		addProfile(newItem.Name, newItem.Start, newItem.End)
		refreshAddedProfile(newItem.Name, newItem.Start, newItem.End)
		cbox.Append(newItem.Name)
		rbox.Append(newItem.Name)
		ui.MsgBox(mainwin,
			"Profile Added",
			"You can now change to this profile in \"Current Settings\"")

	})
	aForm.Append("", abutton, false)

	return vbox
}

// About Page
func makeAboutPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("")
	group.SetMargined(true)
	vbox.Append(group, false)

	group.SetChild(ui.NewNonWrappingMultilineEntry())

	aForm := ui.NewForm()
	aForm.SetPadded(true)
	group.SetChild(aForm)

	about := ui.NewGroup("ABOUT:")
	about.SetMargined(true)
	aForm.Append("", about, false)

	aForm.Append("", ui.NewLabel(""), false)

	aGroup := ui.NewLabel("Charge Control for System76 Laptops by 3DF OSI")
	aForm.Append("", aGroup, false)
	aForm.Append("", ui.NewLabel("v0.1.0"), false)

	aForm.Append("", ui.NewLabel(""), false)

	disclaimer := ui.NewGroup("DISCLAIMER:")
	disclaimer.SetMargined(true)
	aForm.Append("", disclaimer, false)
	aForm.Append("", ui.NewLabel(""), false)
	aForm.Append("", ui.NewLabel("This is NOT software produced by System76. In no way do the maintainers make any guarantees."), false)
	aForm.Append("", ui.NewLabel(""), false)
	aForm.Append("", ui.NewLabel("USE AT YOUR OWN RISK!"), false)
	aForm.Append("", ui.NewLabel(""), false)
	aForm.Append("", ui.NewLabel("For more information, please visit:"), false)
	aForm.Append("", ui.NewLabel(""), false)
	aForm.Append("", ui.NewLabel("GitHub Page - https://github.com/hkdb/s76cc"), false)
	aForm.Append("", ui.NewLabel("3DF Open Source Initiative - https://osi.3df.io"), false)
	aForm.Append("", ui.NewLabel("3DF Limited - https://3df.io"), false)

	return vbox
}

func setupUI() {
	mainwin = ui.NewWindow("Charge Control for System76 Laptops", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	tab := ui.NewTab()
	mainwin.SetChild(tab)
	mainwin.SetMargined(true)

	tab.Append("Controls", makeControlsPage())
	tab.SetMargined(0, true)

	tab.Append("About", makeAboutPage())
	tab.SetMargined(0, true)

	mainwin.Show()
}

func main() {
	getUserEnv()
	ui.Main(setupUI)
}
