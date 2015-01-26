package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tarm/goserial"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Device struct {
	Name     string
	Status   bool
	Commands []*Command
	History  []*Action
	Link     io.ReadWriteCloser
}

type Command struct {
	Name  string
	Bytes []byte
}

type Action struct {
	Command string
	Date    time.Time
}

func (d *Device) Config(filepath string) {
	// Load json file into string
	jsonBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(-1)
	}
	// Anonym struct
	type InterStruct struct {
		Name     string `json:"name"`
		Commands []*struct {
			Name             string   `json:"name"`
			StringCodedBytes []string `json:"bytes"`
			Bytes            []byte   `json:"-"`
		} `json:"commands"`
	}

	// Load it into our intermediate struct containing string encoded commands either in base 10 or hexa
	var intermediateStruct = InterStruct{}
	err = json.Unmarshal(jsonBytes, &intermediateStruct)
	if err != nil {
		fmt.Println("err :", err)
		os.Exit(-1)
	}

	// Convert these string encoded commands into bytes
	for key, value := range intermediateStruct.Commands {
		command := value
		for _, cvalue := range command.StringCodedBytes {
			// TODO check whether string encoded commands actually begins with 0x, if not then it's base 10
			cmd_bytes, err := hex.DecodeString(cvalue[2:])
			if err != nil {
				fmt.Println("err :", err)
				os.Exit(-1)
			}
			// FIX this for commands containing more than one byte
			intermediateStruct.Commands[key].Bytes = append(intermediateStruct.Commands[key].Bytes, cmd_bytes[0])
		}
	}

	// Inject intermediate struct in official device structure
	d.Name = intermediateStruct.Name
	loadedCommands := 0
	for _, value := range intermediateStruct.Commands {
		command := Command{value.Name, value.Bytes}
		d.Commands = append(d.Commands, &command)
		loadedCommands++
	}

	log.Printf("Loaded %d commands for %s\n", loadedCommands, d.Name)
}

func (d *Device) Connect(addr string, baud int) error {
	var err error
	c := &serial.Config{Name: addr, Baud: baud}
	d.Link, err = serial.OpenPort(c)
	if err != nil {
		d.Status = false
	} else {
		d.Status = true
	}
	return err
}

func (d *Device) Close() error {
	err := d.Link.Close()
	if err == nil {
		d.Status = false
	}
	return err

}

func (d *Device) DoCommand(name string) error {
	if d.Status != true {
		return errors.New("DeviceNotConnected")
	}

	for _, value := range d.Commands {
		if name == value.Name {
			_, err := d.Link.Write(value.Bytes)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("CommandNotFound")
}
