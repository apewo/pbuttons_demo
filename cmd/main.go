package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-filemutex"
	"github.com/holoplot/go-evdev"
)

var kb_names = []string{
	//"Apewo Keyboard",             //zmk bluetooth name
	"ZMK Project Apewo Keyboard", //zmk usb name
}

func listDevices() (string, error) {
	//count := 0
	basePath := "/dev/input"

	files, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf(fmt.Sprintf("Cannot read /dev/input: %v\n", err))
	}

	for _, fileName := range files {
		if fileName.IsDir() {
			continue
		}

		full := fmt.Sprintf("%s/%s", basePath, fileName.Name())
		d, err := evdev.Open(full)
		if err == nil {
			name, _ := d.Name()
			fmt.Println(full + " : " + name)
			for _, n := range kb_names {
				if strings.HasPrefix(name, n) {
					fmt.Printf("found : %s:\t%s\n", d.Path(), name)
					return d.Path(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("kb not found")
}

func execcmd(cmd string, args ...string) {

	//out, err := exec.Command(cmd, args...).Output()
	ecmd := exec.Command(cmd, args...)
	go ecmd.Run()
}

func definesignals() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		for {
			sig := <-sigs
			switch sig {
			case os.Interrupt:
				fmt.Println("received INT signal")
				os.Exit(0)
				//return
			case syscall.SIGTERM:
				fmt.Println("received TERM signal")
				os.Exit(0)
			default:
				fmt.Println("received unkown signal signal")
				msg := fmt.Sprint("Ignoring Signal : ", sig)
				fmt.Println(msg)
			}
		}
	}()
}

func main() {

	m, err := filemutex.New("/tmp/evzmk.lock")
	if err != nil {
		fmt.Println("Directory did not exist or file could not created")
		os.Exit(1)
	}

	err = m.TryLock()

	if err != nil {
		fmt.Println("another process is already running")
		os.Exit(1)
	}

	definesignals()

	for {
		dpath := ""
		var err error
		for {
			dpath, err = listDevices()
			if err != nil {
				fmt.Println("error : " + err.Error())
				time.Sleep(time.Second * 5)
			}
			break
		}

		fmt.Println("found path : " + dpath)
		d, err := evdev.Open(dpath)
		if err != nil {
			fmt.Printf("Cannot read: %v\n", err)
			continue
		}
		vMajor, vMinor, vMicro := d.DriverVersion()
		fmt.Printf("Input driver version is %d.%d.%d\n", vMajor, vMinor, vMicro)
		inputID, err := d.InputID()
		if err == nil {
			fmt.Printf("Input device ID: bus 0x%x vendor 0x%x product 0x%x version 0x%x\n",
				inputID.BusType, inputID.Vendor, inputID.Product, inputID.Version)
		}
		phys, err := d.PhysicalLocation()
		if err == nil {
			fmt.Printf("Input device physical location: %s\n", phys)
		}
		err = d.NonBlock()
		if err != nil {
			panic(err)
		}

		for {
			e, err := d.ReadOne()
			if err != nil {
				fmt.Printf("Error reading from device: %v\n", err)
				break
			}
			fmt.Println("read...")
			switch e.Type {
			case evdev.EV_KEY:
				if e.Value == 0 { //pressed
					switch e.Code {
					case evdev.KEY_MACRO1:
						fmt.Println("receive KEY_MACRO1")
						execcmd("rofi" /*, rofi_args_1...*/)
					case evdev.KEY_MACRO2:
						fmt.Println("receive KEY_MACRO2")
						execcmd("rofi" /*, rofi_args_2...*/)
					case evdev.KEY_MACRO3:
						fmt.Println("receive KEY_MACRO3")
						execcmd("/usr/bin/onboard")
						//........ up to 32
					}
				}
			}
		}
	}
}
