package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Device string `short:"v" long:"device" default:"intel_backlight" required:"true" description:"brightness device"`
	Inc    uint   `short:"i" long:"inc" default:"nil" description:"increment brightness up to given percentage between [1 -10]"`
	Dec    uint   `short:"d" long:"dec" default:"nil" description:"decrement brightness down to percentage between [1 -10]"`
	Set    uint   `short:"s" long:"set" default:"nil" description:"set brightness to given percentage between [1-99]"`
	Get    bool   `short:"g" long:"get" description:"get actual brightness percentage"`
}

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
	nooptMsg    = "Error no options, try with -h"

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
	if value, err := ioutil.ReadFile(file); err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}

func writeStringToFile(file string, value string) error {
	if f, err := os.OpenFile(file, os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer f.Close()
		f.WriteString(value)
		f.Sync()
		return nil
	}
}

func checkDevice(path string) ([]os.FileInfo, error) {
	var result []os.FileInfo
	if files, err := ioutil.ReadDir(path); err != nil {
		return nil, err
	} else {
		for _, f := range files {
			for _, e := range driverFiles {
				if e == f.Name() {
					result = append(result, f)
				}
			}
		}
	}
	if len(result) == 3 {
		return result, nil
	} else {
		return nil, fmt.Errorf(driverMsg)
	}
}

func (bc *BrightnessControl) LoadParams(files []os.FileInfo) error {
	if len(files) == 3 {
		for _, file := range files {
			switch file.Name() {
			case "actual_brightness":
				if actual, err := readFile(bc.Path + file.Name()); err != nil {
					return err
				} else {
					if a, err := strconv.ParseInt(strings.Trim(actual, "\n"), 10, 0); err != nil {
						return err
					} else {
						bc.ActualBrightness = a
					}
				}
			case "brightness":
				if brightness, err := readFile(bc.Path + file.Name()); err != nil {
					return err
				} else {
					if c, err := strconv.ParseInt(strings.Trim(brightness, "\n"), 10, 0); err != nil {
						return err
					} else {
						bc.Brightness = c
					}
				}
			case "max_brightness":
				if max, err := readFile(bc.Path + file.Name()); err != nil {
					return err
				} else {
					if m, err := strconv.ParseInt(strings.Trim(max, "\n"), 10, 0); err != nil {
						return err
					} else {
						bc.MaxBrightness = m
					}
				}
			default:
				return fmt.Errorf("Error no proper files on driver folder")
			}
		}
		return nil
	} else {
		return fmt.Errorf(driverMsg)
	}

}

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

func (bc *BrightnessControl) Init() error {
	bc.Path = syspath + bc.Config.Device + "/"

	if files, err := checkDevice(bc.Path); err != nil {
		return err
	} else {
		if err := bc.LoadParams(files); err != nil {
			return err
		}
	}
	return nil
}

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
		if err := bc.Set(); err != nil {
			return "", err
		} else {
			return "", nil
		}

	}

	if bc.Config.Dec > 0 {
		if err := bc.ValidateOptions("dec"); err != nil {
			return "", err
		}
		if err := bc.Dec(); err != nil {
			return "", err
		} else {
			return "", nil
		}
	}

	if bc.Config.Inc > 0 {
		if err := bc.ValidateOptions("inc"); err != nil {
			return "", err
		}
		if err := bc.Inc(); err != nil {
			return "", err
		} else {
			return "", nil
		}
	}
	return "", fmt.Errorf(nooptMsg)

}

func (bc *BrightnessControl) Get() (string, error) {
	actualPct := strconv.Itoa(int(bc.ActualBrightness) * 100 / int(bc.MaxBrightness))
	return actualPct, nil
}

func (bc *BrightnessControl) Inc() error {
	value := int(bc.ActualBrightness) + int(int(bc.Config.Inc)*int(bc.MaxBrightness)/100)

	if value > 0 && value <= int(bc.MaxBrightness) {
		if err := writeStringToFile(bc.Path+"brightness", strconv.Itoa(value)); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BrightnessControl) Dec() error {
	value := int(bc.ActualBrightness) - int(int(bc.Config.Dec)*int(bc.MaxBrightness)/100)

	if value > 0 && value <= int(bc.MaxBrightness) {
		if err := writeStringToFile(bc.Path+"brightness", strconv.Itoa(value)); err != nil {
			return err
		}
	}
	return nil
}

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
		fmt.Println(example)
		os.Exit(1)
	} else {
		if out, err := bc.Run(); err != nil {
			fmt.Println("An error occured : ", err)
			os.Exit(1)
		} else {
			if out != "" {
				fmt.Println(out)
			}
			os.Exit(0)
		}
	}
}
