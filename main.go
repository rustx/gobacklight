package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

// Config struct parses and validates the options from the command line
type Config struct {
	Device string `short:"v" long:"device" default:"intel_backlight" required:"true" description:"brightness device"`
	Inc    uint   `short:"i" long:"inc" default:"nil" description:"increment brightness up to given percentage between [1 -10]"`
	Dec    uint   `short:"d" long:"dec" default:"nil" description:"decrement brightness down to percentage between [1 -10]"`
	Set    uint   `short:"s" long:"set" default:"nil" description:"set brightness to given percentage between [1-99]"`
	Get    bool   `short:"g" long:"get" description:"get actual brightness percentage"`
}

// BrightnessControl is the main object, loading the device values and executing actions.
type BrightnessControl struct {
	*Config
	Path             string
	Brightness       int64
	ActualBrightness int64
	MaxBrightness    int64
}

var (
	config  Config
	syspath = "/sys/class/backlight/"

	combinedMsg = "Error combined options"
	setMsg      = "Error value must be between 1 and 99"
	valueMsg    = "Error value must be between 1 and 10"
	nilMsg      = "Error action is nil"
	driverMsg   = "Error driver files not found in device path"
	nooptMsg    = "Error no options, try gobacklight -h"

	nofileMsg = "open .*: no such file or directory"

	driverFiles = [3]string{"brightness", "actual_brightness", "max_brightness"}

	example = `Examples :
	gobacklight -v intel_backlight -g
	gobacklight -v intel_backlight -i 5
	gobacklight -v intel_backlight -d 5
	gobacklight -v intel_backlight -s 25
`
)

func readFile(file string) (string, error) {
	value, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func writeStringToFile(file string, value string) error {
	f, err := os.OpenFile(file, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(value)
	f.Sync()
	return nil
}

func checkDevice(path string) ([]os.FileInfo, error) {
	var result []os.FileInfo
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		for _, e := range driverFiles {
			if e == f.Name() {
				result = append(result, f)
			}
		}
	}

	if len(result) == 3 {
		return result, nil
	}
	return nil, fmt.Errorf(driverMsg)
}

// LoadParams reads all files in driver folder, and fills BrightnessControl fields.
// It expects that your driver folder contains at least 3 files : brightness, actual_brightness, max_brightness.
// It returns an error when the driver files could not be read, or converted to string.
func (bc *BrightnessControl) LoadParams(files []os.FileInfo) error {
	if len(files) == 3 {
		for _, file := range files {
			switch file.Name() {
			case "actual_brightness":
				actual, err := readFile(bc.Path + file.Name())
				if err != nil {
					return err
				}
				a, err := strconv.ParseInt(strings.Trim(actual, "\n"), 10, 0)
				if err != nil {
					return err
				}
				bc.ActualBrightness = a
			case "brightness":
				brightness, err := readFile(bc.Path + file.Name())
				if err != nil {
					return err
				}
				b, err := strconv.ParseInt(strings.Trim(brightness, "\n"), 10, 0)
				if err != nil {
					return err
				}
				bc.Brightness = b
			case "max_brightness":
				max, err := readFile(bc.Path + file.Name())
				if err != nil {
					return err
				}
				m, err := strconv.ParseInt(strings.Trim(max, "\n"), 10, 0)
				if err != nil {
					return err
				}
				bc.MaxBrightness = m
			default:
				return fmt.Errorf("Error no proper files on driver folder")
			}
		}
		return nil
	}
	return fmt.Errorf(driverMsg)
}

// ValidateOptions validate the settings provided to BrightnessControl with the command line.
// It checks that a unique action was called with command line, and returns an error when combined actions are called.
// It also checks the command line values prior to execute actions, and returns an error when not OK.
func (bc *BrightnessControl) ValidateOptions(action string) error {

	switch action {
	case "get":
		if bc.Config.Set > 0 || bc.Config.Inc > 0 || bc.Config.Dec > 0 {
			return fmt.Errorf(combinedMsg)
		}
	case "set":
		if bc.Config.Inc > 0 || bc.Config.Dec > 0 || bc.Config.Get == true {
			return fmt.Errorf(combinedMsg)
		}
		if bc.Config.Set > 100 || bc.Config.Set <= 0 {
			return fmt.Errorf(setMsg)
		}
	case "dec":
		if bc.Config.Inc > 0 || bc.Config.Set > 0 || bc.Config.Get == true {
			return fmt.Errorf(combinedMsg)
		}
		if bc.Config.Dec > 10 || bc.Config.Dec <= 0 {
			return fmt.Errorf(valueMsg)
		}
	case "inc":
		if bc.Config.Dec > 0 || bc.Config.Set > 0 || bc.Config.Get == true {
			return fmt.Errorf(combinedMsg)
		}
		if bc.Config.Inc > 10 || bc.Config.Inc <= 0 {
			return fmt.Errorf(valueMsg)
		}
	default:
		return fmt.Errorf(nilMsg)
	}
	return nil
}

// Init checks if the BrightnessControl can load the values from the device files.
// It uses the checkDevice helper to ensure all files are present in the device folder given by the command line.
// It returns an error if the device path doesn't contain files needed.
func (bc *BrightnessControl) Init() error {
	bc.Path = syspath + bc.Config.Device + "/"
	if _, err := os.Stat(bc.Path); os.IsNotExist(err) {
		return err
	}
	files, err := checkDevice(bc.Path)
	if err != nil {
		return err
	}
	return bc.LoadParams(files)
}

// Run validate BrightnessControl options, and run actions from the command line arguments.
// When calling the get action it returns the current brightness in stdout, else stdout is empty.
// It returns an error if the action called encountered an error.
func (bc *BrightnessControl) Run() (string, error) {
	if bc.Config.Get == true {
		if err := bc.ValidateOptions("get"); err != nil {
			return "", err
		}
		v, _ := bc.Get()
		return v, nil
	}
	if bc.Config.Set > 0 {
		if err := bc.ValidateOptions("set"); err != nil {
			return "", err
		}
		err := bc.Set()
		if err != nil {
			return "", err
		}
		return "", nil
	}
	if bc.Config.Dec > 0 {
		if err := bc.ValidateOptions("dec"); err != nil {
			return "", err
		}
		err := bc.Dec()
		if err != nil {
			return "", err
		}
		return "", nil
	}
	if bc.Config.Inc > 0 {
		if err := bc.ValidateOptions("inc"); err != nil {
			return "", err
		}
		err := bc.Inc()
		if err != nil {
			return "", err
		}
		return "", nil
	}
	return "", fmt.Errorf(nooptMsg)

}

// Get returns the current brightness expressed as percentage
// It uses the MaxBrightness and ActualBrightness fields to return the current Brightness as a percentage.
func (bc *BrightnessControl) Get() (string, error) {
	actualPct := strconv.Itoa(int(bc.ActualBrightness) * 100 / int(bc.MaxBrightness))
	return actualPct, nil
}

// Inc will increment the current brightness with a percentage between 1 and 10
// It uses the MaxBrightness and ActualBrightness fields to apply the given percentage to current Brightness.
// It returns an error if it could not write the new value to the brightness file.
func (bc *BrightnessControl) Inc() error {
	value := int(bc.ActualBrightness) + int(int(bc.Config.Inc)*int(bc.MaxBrightness)/100)

	if value > 0 && value <= int(bc.MaxBrightness) {
		if err := writeStringToFile(bc.Path+"brightness", strconv.Itoa(value)); err != nil {
			return err
		}
	}
	return nil
}

// Dec will decrement the current brightness with a percentage between 1 and 10
// It uses the MaxBrightness and ActualBrightness fields to apply the given percentage to current Brightness.
// It returns an error if it could not write the new value to the brightness file.
func (bc *BrightnessControl) Dec() error {
	value := int(bc.ActualBrightness) - int(int(bc.Config.Dec)*int(bc.MaxBrightness)/100)

	if value > 0 && value <= int(bc.MaxBrightness) {
		if err := writeStringToFile(bc.Path+"brightness", strconv.Itoa(value)); err != nil {
			return err
		}
	}
	return nil
}

// Set will set the current brightness with a percentage between 1 and 100
// It uses the MaxBrightness and ActualBrightness fields to apply the given percentage to current Brightness.
// It returns an error if it could not write the new value to the brightness file.
func (bc *BrightnessControl) Set() error {
	value := int(int(bc.Config.Set) * int(bc.MaxBrightness) / 100)

	if value > 0 && value <= int(bc.MaxBrightness) {
		if err := writeStringToFile(bc.Path+"brightness", strconv.Itoa(value)); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	bc := BrightnessControl{Config: &config}
	if _, err := flags.Parse(&config); err != nil {
		fmt.Println(example)
		os.Exit(1)
	}
	if err := bc.Init(); err != nil {
		fmt.Println("An error occurred : ", err)
		fmt.Println(example)
		os.Exit(1)
	} else {
		if out, err := bc.Run(); err != nil {
			fmt.Println("An error occurred : ", err)
			fmt.Println(example)
			os.Exit(1)
		} else {
			if out != "" {
				fmt.Println(out)
			}
			os.Exit(0)
		}
	}
}
