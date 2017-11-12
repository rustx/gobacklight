package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type GobacklightSuite struct {
	dir   string
	files [3]string
}

var _ = Suite(&GobacklightSuite{})
var secret = rand.NewSource(time.Now().UnixNano())

func (s *GobacklightSuite) SetUpSuite(c *C) error {
	s.dir = c.MkDir()
	s.files = [3]string{"brightness", "actual_brightness", "max_brightness"}

	syspath = filepath.Join(s.dir, "intel_backlight") + "/"
	if err := os.Mkdir(syspath, 0755); err != nil {
		c.Fatalf("error creating syspath directory : %s", err)
	}

	r := rand.New(secret)

	bright := r.Intn(500)
	max := 1000

	for _, file := range s.files {
		f, err := os.Create(filepath.Join(syspath, file))
		if err != nil {
			c.Fatalf("error creating file : ", file)
			return err
		} else {
			defer f.Close()
			if file == "actual_brightness" || file == "brightness" {
				if _, err := f.WriteString(strconv.Itoa(bright)); err != nil {
					c.Fatalf("error writing file :", file)
					return err
				}
				f.Sync()
			} else {
				if _, err := f.WriteString(strconv.Itoa(max)); err != nil {
					c.Fatalf("error writing file :", file)
					return err
				}
				f.Sync()
			}
		}
	}
	return nil
}

func (s *GobacklightSuite) TearDownTest(c *C) {
	os.Chmod(s.dir, 0755)
	os.Chmod(syspath, 0755)

	r := rand.New(secret)

	bright := r.Intn(500)
	max := 1000

	for _, file := range s.files {
		os.Chmod(syspath+file, 0644)
		f, err := os.OpenFile(filepath.Join(syspath, file), os.O_WRONLY, 0644)
		if err != nil {
			if f, err := os.Create(filepath.Join(syspath, file)); err != nil {
				c.Fatalf("error creating file : %s", file)
			} else {
				defer f.Close()
				if file == "actual_brightness" || file == "brightness" {
					if _, err := f.WriteString(strconv.Itoa(bright)); err != nil {
						c.Fatalf("error writing file :", file)
					}
					f.Sync()
				} else {
					if _, err := f.WriteString(strconv.Itoa(max)); err != nil {
						c.Fatalf("error writing file :", file)
					}
					f.Sync()
				}
			}
		} else {
			defer f.Close()
			if file == "actual_brightness" || file == "brightness" {
				if _, err := f.WriteString(strconv.Itoa(bright)); err != nil {
					c.Fatalf("error writing file :", file)
				}
				f.Sync()
			} else {
				if _, err := f.WriteString(strconv.Itoa(max)); err != nil {
					c.Fatalf("error writing file :", file)
				}
				f.Sync()
			}
		}
	}
}

func (s *GobacklightSuite) TearDownSuite(c *C) error {
	if err := os.RemoveAll(syspath); err != nil {
		return err
	}
	return nil
}

func (s *GobacklightSuite) TestConfig(c *C) {
	value := uint(5)
	conf := Config{
		Inc: value,
	}

	ic := BrightnessControl{Config: &conf, Path: syspath}

	c.Assert(ic.Config.Inc, Equals, value)
	c.Assert(ic.Config.Dec, Equals, uint(0))
	c.Assert(ic.Config.Set, Equals, uint(0))
	c.Assert(ic.Config.Get, Equals, false)
}

func (s *GobacklightSuite) TestGetConfig(c *C) {
	value := true
	conf := Config{
		Get: value,
	}

	gc := BrightnessControl{Config: &conf, Path: syspath}

	c.Assert(gc.Config.Get, Equals, value)
	c.Assert(gc.Config.Set, Equals, uint(0))
	c.Assert(gc.Config.Dec, Equals, uint(0))
	c.Assert(gc.Config.Inc, Equals, uint(0))
}

func (s *GobacklightSuite) TestSetConfig(c *C) {
	value := uint(75)
	conf := Config{
		Set: value,
	}

	sc := BrightnessControl{Config: &conf, Path: syspath}

	c.Assert(sc.Config.Set, Equals, value)
	c.Assert(sc.Config.Dec, Equals, uint(0))
	c.Assert(sc.Config.Inc, Equals, uint(0))
	c.Assert(sc.Config.Get, Equals, false)
}

func (s *GobacklightSuite) TestDecConfig(c *C) {
	value := uint(5)
	conf := Config{
		Dec: value,
	}

	dc := BrightnessControl{Config: &conf, Path: syspath}

	c.Assert(dc.Config.Dec, Equals, value)
	c.Assert(dc.Config.Inc, Equals, uint(0))
	c.Assert(dc.Config.Set, Equals, uint(0))
	c.Assert(dc.Config.Get, Equals, false)
}

func (s *GobacklightSuite) TestGetAndDecConfig(c *C) {
	value := uint(5)

	conf := Config{
		Dec: value,
		Get: true,
	}

	gd := BrightnessControl{Config: &conf, Path: syspath}

	c.Assert(gd.Config.Dec, Equals, value)
	c.Assert(gd.Config.Inc, Equals, uint(0))
	c.Assert(gd.Config.Set, Equals, uint(0))
	c.Assert(gd.Config.Get, Equals, true)
}

func (s *GobacklightSuite) TestCheckDeviceFilePathOk(c *C) {
	conf := Config{
		Get: true,
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, err := checkDevice(syspath)

	c.Assert(err, IsNil)
	c.Assert(files, HasLen, 3)
	c.Assert(bc.MaxBrightness, Not(Equals), 0)
	c.Assert(bc.ActualBrightness, Not(Equals), 0)
	c.Assert(bc.Brightness, Not(Equals), 0)
}

func (s *GobacklightSuite) TestCheckDeviceFilePathPermissionKo(c *C) {
	if err := os.Chmod(syspath, 0300); err != nil {
		c.Fatal("Error chmod test_data directory")
	}

	files, err := checkDevice(syspath)

	c.Assert(err, Not(IsNil))
	c.Assert(files, HasLen, 0)
	c.Assert(err, ErrorMatches, "open .*: permission denied")
}

func (s *GobacklightSuite) TestReadFileOk(c *C) {
	for _, f := range s.files {
		value, err := readFile(syspath + f)
		c.Assert(err, IsNil)
		c.Assert(value, Not(Equals), "")
		c.Assert(value, Not(HasLen), 0)
	}
}

func (s *GobacklightSuite) TestReadFilePermissionKo(c *C) {
	for _, f := range s.files {
		if err := os.Chmod(syspath+f, 0300); err != nil {
			c.Fatalf("error chmod file ", f)
		}

		value, err := readFile(syspath + f)
		c.Assert(err, Not(IsNil))
		c.Assert(value, Equals, "")
		c.Assert(err, ErrorMatches, "open .*: permission denied")
	}
}

func (s *GobacklightSuite) TestWriteFileToStringOk(c *C) {
	err := writeStringToFile(syspath+"brightness", strconv.Itoa(500))
	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestWriteFileToStringFileKo(c *C) {
	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal("error chmod file ")
	}
	err := writeStringToFile(syspath+"brightness", strconv.Itoa(500))
	c.Assert(err, Not(IsNil))
}

func (s *GobacklightSuite) TestLoadParamsFilePathOk(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)
	err := bc.LoadParams(files)

	c.Assert(err, IsNil)
	c.Assert(bc.MaxBrightness, Equals, int64(1000))
	c.Assert(bc.Brightness, Equals, bc.ActualBrightness)
	c.Assert(bc.Brightness, Not(Equals), 0)
	c.Assert(bc.ActualBrightness, Not(Equals), 0)
}

func (s *GobacklightSuite) TestLoadParamsFilePathKo(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, errdev := checkDevice("/tmp")
	errlp := bc.LoadParams(files)

	c.Assert(errdev, Not(IsNil))
	c.Assert(errdev, ErrorMatches, driverMsg)

	c.Assert(errlp, Not(IsNil))
	c.Assert(errlp, ErrorMatches, driverMsg)

	c.Assert(bc.Brightness, Equals, int64(0))
	c.Assert(bc.MaxBrightness, Equals, int64(0))
	c.Assert(bc.ActualBrightness, Equals, int64(0))
}

func (s *GobacklightSuite) TestLoadParamsFileContentMaxKo(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)

	for _, f := range s.files {
		if f == "max_brightness" {
			if f, err := os.OpenFile(bc.Path+f, os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
				c.Fatal(err)
			} else {
				defer f.Close()
				//if err := os.Truncate(f.Name(), 0); err != nil {
				//	c.Fatal(err)
				//}
				if _, err := f.WriteString("a"); err != nil {
					c.Fatal(err)
				} else {
					f.Sync()
					err := bc.LoadParams(files)
					c.Assert(err, Not(IsNil)) // strconv value type error
					c.Assert(err, ErrorMatches, "strconv.ParseInt:.*")
				}
			}
		}
	}
}

func (s *GobacklightSuite) TestLoadParamsFileContentActualKo(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)

	for _, f := range s.files {
		if f == "actual_brightness" {
			if f, err := os.OpenFile(bc.Path+f, os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
				c.Fatal(err)
			} else {
				defer f.Close()
				if err := os.Truncate(f.Name(), 0); err != nil {
					c.Fatal(err)
				}
				if _, err := f.WriteString("a"); err != nil {
					c.Fatal(err)
				} else {
					f.Sync()
					err := bc.LoadParams(files)
					c.Assert(err, Not(IsNil)) // strconv value type error
					c.Assert(err, ErrorMatches, "strconv.ParseInt:.*")
				}
			}
		}
	}
}

func (s *GobacklightSuite) TestLoadParamsFileContentBrightKo(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)

	for _, f := range s.files {
		if f == "brightness" {
			if f, err := os.OpenFile(bc.Path+f, os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
				c.Fatal(err)
			} else {
				defer f.Close()
				if err := os.Truncate(f.Name(), 0); err != nil {
					c.Fatal(err)
				}
				if _, err := f.WriteString("a"); err != nil {
					c.Fatal(err)
				} else {
					f.Sync()
					err := bc.LoadParams(files)
					c.Assert(err, Not(IsNil)) // strconv value type error
					c.Assert(err, ErrorMatches, "strconv.ParseInt:.*")
				}
			}
		}
	}
}

func (s *GobacklightSuite) TestLoadParamsFileContentEmpty(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)

	for _, f := range s.files {
		os.Remove(bc.Path + f)
		if _, err := os.Create(bc.Path + f); err != nil {
			c.Fatal(err)
		}
		err := bc.LoadParams(files)

		c.Assert(err, Not(IsNil)) // ioutil read type error
		c.Assert(err, ErrorMatches, "strconv.ParseInt:.*")
	}
}

func (s *GobacklightSuite) TestLoadParamsFileRenamedKo(c *C) {
	bc := BrightnessControl{Path: syspath}
	files, _ := checkDevice(syspath)

	c.Assert(files, HasLen, 3)

	for _, f := range s.files {
		if err := os.Rename(syspath+f, syspath+f+"_test"); err != nil {
			c.Fatal("Error chmod test_data files : ", f)
		} else {
			err := bc.LoadParams(files)

			c.Assert(err, Not(IsNil))
			c.Assert(err, ErrorMatches, nofileMsg)
		}
		if err := os.Rename(syspath+f+"_test", syspath+f); err != nil {
			c.Fatal("Error chmod test_data files : ", f)
		}
	}
}

func (s *GobacklightSuite) TestLoadParamsFileAbsentKo(c *C) {
	bc := BrightnessControl{Path: syspath}

	for _, f := range s.files {
		if err := os.Rename(syspath+f, syspath+f+"_test"); err != nil {
			c.Fatal("Error chmod test_data files : ", f)
		}
	}

	files, _ := ioutil.ReadDir(syspath)

	err := bc.LoadParams(files)
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, "Error no proper files on driver folder")

	for _, f := range s.files {
		if err := os.Rename(syspath+f+"_test", syspath+f); err != nil {
			c.Fatal("Error chmod test_data files : ", f)
		}
	}
}

func (s *GobacklightSuite) TestValidateOptionsGetOk(c *C) {
	value := true
	conf := Config{
		Get: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("get")

	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestValidateOptionsIncOk(c *C) {
	value := uint(5)
	conf := Config{
		Inc: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("inc")

	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestValidateOptionsDecOk(c *C) {
	value := uint(5)
	conf := Config{
		Dec: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("dec")

	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestValidateOptionsSetOk(c *C) {
	value := uint(25)
	conf := Config{
		Set: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("set")

	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestValidateOptionsGetKo(c *C) {
	value := true
	conf := Config{
		Get: value,
		Set: uint(25),
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("get")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestValidateOptionsIncKo(c *C) {
	value := uint(5)
	conf := Config{
		Inc: value,
		Get: true,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("inc")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestValidateOptionsSetKo(c *C) {
	value := uint(25)
	conf := Config{
		Set: value,
		Get: true,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("set")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestValidateOptionsDecKo(c *C) {
	value := uint(5)
	conf := Config{
		Dec: value,
		Get: true,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("dec")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestValidateOptionsSetValueKo(c *C) {
	value := uint(105)
	conf := Config{
		Set: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("set")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, setMsg)
}

func (s *GobacklightSuite) TestValidateOptionsIncValueKo(c *C) {
	value := uint(105)
	conf := Config{
		Inc: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("inc")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, valueMsg)
}

func (s *GobacklightSuite) TestValidateOptionsDecValueKo(c *C) {
	value := uint(105)
	conf := Config{
		Dec: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("dec")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, valueMsg)
}

func (s *GobacklightSuite) TestValidateOptionsDefaultValueKo(c *C) {
	value := uint(105)
	conf := Config{
		Dec: value,
	}

	bc := BrightnessControl{Config: &conf, Path: syspath}
	err := bc.ValidateOptions("test")

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, nilMsg)
}

func (s *GobacklightSuite) TestGetActionOk(c *C) {
	conf := Config{
		Get: true,
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	v, err := bc.Get()
	c.Assert(err, IsNil)
	c.Assert(v, Not(IsNil))
	c.Assert(v, Not(Equals), 0)
}

func (s *GobacklightSuite) TestIncActionOk(c *C) {
	conf := Config{
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	err := bc.Inc()
	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestDecActionOk(c *C) {
	conf := Config{
		Dec: uint(5),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	err := bc.Dec()
	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestSetActionOk(c *C) {
	conf := Config{
		Set: uint(25),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	err := bc.Set()
	c.Assert(err, IsNil)
}

func (s *GobacklightSuite) TestIncActionFileKo(c *C) {
	conf := Config{
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal("error chmod file ")
	}

	err := bc.Inc()
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestDecActionFileKo(c *C) {
	conf := Config{
		Dec: uint(5),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal("error chmod file ")
	}

	err := bc.Dec()
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestSetActionFileKo(c *C) {
	conf := Config{
		Set: uint(25),
	}
	bc := BrightnessControl{Config: &conf, Path: syspath}

	files, _ := checkDevice(bc.Path)

	_ = bc.LoadParams(files)

	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal("error chmod file ")
	}

	err := bc.Set()
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestInitOk(c *C) {
	conf := Config{}
	bc := BrightnessControl{Config: &conf}
	bc.Init()

	c.Assert(bc, Not(IsNil))
	c.Assert(bc.Path, Not(IsNil))
	c.Assert(bc.Device, Not(IsNil))
	c.Assert(bc.Path, Equals, syspath+bc.Device+"/")
	c.Assert(bc.Brightness, Not(Equals), uint(0))
	c.Assert(bc.MaxBrightness, Not(Equals), uint(0))
	c.Assert(bc.ActualBrightness, Not(Equals), uint(0))
}

func (s *GobacklightSuite) TestInitDevicePermissionKo(c *C) {
	conf := Config{}
	bc := BrightnessControl{Config: &conf}

	if err := os.Chmod(syspath+bc.Device, 0300); err != nil {
		c.Fatalf("%s", err)
	}
	err := bc.Init()

	c.Assert(err, Not(IsNil))
	c.Assert(bc, Not(IsNil))
	c.Assert(bc.Path, Not(IsNil))
	c.Assert(bc.Device, Not(IsNil))
	c.Assert(bc.Path, Equals, syspath+bc.Device+"/")
	c.Assert(bc.Brightness, Equals, int64(0))
	c.Assert(bc.MaxBrightness, Equals, int64(0))
	c.Assert(bc.ActualBrightness, Equals, int64(0))
}

func (s *GobacklightSuite) TestInitFilePermissionKo(c *C) {
	conf := Config{}
	bc := BrightnessControl{Config: &conf}

	for _, f := range s.files {
		if err := os.Chmod(syspath+f, 0300); err != nil {
			c.Fatalf("%s", err)
		}
	}
	err := bc.Init()

	c.Assert(err, Not(IsNil))
	c.Assert(bc, Not(IsNil))
	c.Assert(bc.Path, Not(IsNil))
	c.Assert(bc.Device, Not(IsNil))
	c.Assert(bc.Path, Equals, syspath+bc.Device+"/")
	c.Assert(bc.Brightness, Equals, int64(0))
	c.Assert(bc.MaxBrightness, Equals, int64(0))
	c.Assert(bc.ActualBrightness, Equals, int64(0))
}

func (s *GobacklightSuite) TestRunGetOk(c *C) {
	conf := Config{
		Get: true,
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if v, err := bc.Run(); err != nil {
		c.Fatal(err)
	} else {
		c.Assert(err, IsNil)
		c.Assert(v, Not(IsNil))
		c.Assert(v, Not(Equals), 0)
	}
}

func (s *GobacklightSuite) TestRunOptKo(c *C) {
	conf := Config{}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, nooptMsg)
	c.Assert(v, Equals, "")
}

func (s *GobacklightSuite) TestRunGetCombinedKo(c *C) {
	conf := Config{
		Get: true,
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}

	v, err := bc.Run()
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestRunIncOk(c *C) {
	conf := Config{
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if v, err := bc.Run(); err != nil {
		c.Fatal(err)
	} else {
		c.Assert(err, IsNil)
		c.Assert(v, Not(IsNil))
		c.Assert(v, Not(Equals), 0)
	}
}

func (s *GobacklightSuite) TestRunIncValueKo(c *C) {
	conf := Config{
		Inc: uint(15),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, valueMsg)
}

func (s *GobacklightSuite) TestRunIncFileKo(c *C) {
	conf := Config{
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestRunIncCombinedKo(c *C) {
	conf := Config{
		Get: true,
		Inc: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}

	v, err := bc.Run()
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestRunDecOk(c *C) {
	conf := Config{
		Dec: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if v, err := bc.Run(); err != nil {
		c.Fatal(err)
	} else {
		c.Assert(err, IsNil)
		c.Assert(v, Not(IsNil))
		c.Assert(v, Not(Equals), 0)
	}
}

func (s *GobacklightSuite) TestRunDecValueKo(c *C) {
	conf := Config{
		Dec: uint(15),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, valueMsg)
}

func (s *GobacklightSuite) TestRunDecFileKo(c *C) {
	conf := Config{
		Dec: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestRunDecCombinedKo(c *C) {
	conf := Config{
		Get: true,
		Dec: uint(5),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}

	v, err := bc.Run()
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, combinedMsg)
}

func (s *GobacklightSuite) TestRunSetOk(c *C) {
	conf := Config{
		Set: uint(25),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if v, err := bc.Run(); err != nil {
		c.Fatal(err)
	} else {
		c.Assert(err, IsNil)
		c.Assert(v, Not(IsNil))
		c.Assert(v, Not(Equals), 0)
	}
}

func (s *GobacklightSuite) TestRunSetValueKo(c *C) {
	conf := Config{
		Set: uint(150),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, setMsg)
}

func (s *GobacklightSuite) TestRunSetFileKo(c *C) {
	conf := Config{
		Set: uint(25),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}
	if err := os.Remove(syspath + "brightness"); err != nil {
		c.Fatal(err)
	}
	v, err := bc.Run()

	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, nofileMsg)
}

func (s *GobacklightSuite) TestRunSetCombinedKo(c *C) {
	conf := Config{
		Get: true,
		Set: uint(25),
	}
	bc := BrightnessControl{Config: &conf}
	if err := bc.Init(); err != nil {
		c.Fatal(err)
	}

	v, err := bc.Run()
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
	c.Assert(err, ErrorMatches, combinedMsg)
}
