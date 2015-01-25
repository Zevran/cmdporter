//package nec
package main

// Note : user manual advises to lower baud rate to 9600 for long cables
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type nec_m271_m311 struct {
	Commands map[string][]byte //PowerOn, PowerOff, SoundMuteOn, SoundMuteOff, ...
}

var Nec_m271_m311 nec_m271_m311

type JSONCommands struct {
	Commands []JSONCommand
}

type JSONCommand struct {
	CommandName      string
	StringCodedBytes []string `json:"bytes"`
	Bytes            []byte
}

//func init() {
func main() {

	//MAKE MAP
	Nec_m271_m311.Commands = make(map[string][]byte)

	/*********************************************************************************************************
		FROM JSON FILE, WHICH INCLUDED ALL SCENARIOS (PowerOn, PowerOff, SoundMuteOn, SoundMuteOff, ...) ON THE DEVICE,
	LET S BUILD AN EASY MANNER TO GET FROM A COMMAND, ITS REPRESENTATION IN THE SEQUENCE OF BYTES
	FOR INSTANCE Commands["POWER ON"] = [0x02, 0x12, 0x00, 0x00, 0x00, 0x14]
	**********************************************************************************************************/

	var err error

	//IMPORT FROM A JSON FILE
	file, err := ioutil.ReadFile("../../config/commands.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(-1)
	}
	json_string := string(file)

	//TRANSFORM THE JSON TO STUCT DATA
	var res = JSONCommands{}
	err = json.Unmarshal([]byte(json_string), &res)
	if err != nil {
		fmt.Println("err :", err)
		os.Exit(-1)
	}

	//CONVERT THE STRING DATA TO BYTE DATA, USING FOR SEND INSTRUCTIONS TO HARDWARE DEVICES
	for key, value := range res.Commands {
		command := value
		for _, cvalue := range command.StringCodedBytes {
			chex, err := hex.DecodeString(cvalue[2:])
			if err != nil {
				fmt.Println("err :", err)
				os.Exit(-1)
			}
			res.Commands[key].Bytes = append(res.Commands[key].Bytes, chex[0])
		}
	}

	//CREATE A MAPPING FOR THE nec_m271_m311 COMMANDS
	Nec_m271_m311.Commands = make(map[string][]byte)

	for _, value2 := range res.Commands {
		command := value2
		Nec_m271_m311.Commands[command.CommandName] = command.Bytes
	}
	fmt.Println(Nec_m271_m311.Commands["PowerOn"])

	//FROM JSON FILE, WHICH INCLUDED ALL SCENARIOS TO
	// type BytesContainer struct {
	// StringCodedBytes []string `json:"bytes"`
	// Bytes []byte
	// }

	// var err error
	// res := &BytesContainer{}
	// err = json.Unmarshal([]byte(json_string), &res)
	// if err != nil {
	// 	fmt.Println("err :", err)
	// 	os.Exit(-1)
	// }

	// for _, StringCodedByte := range res.StringCodedBytes {
	// 	var v []byte
	// 	v, err = hex.DecodeString(StringCodedByte[2:])
	// 	fmt.Printf("Byte %d\n", v)

	// 	if err != nil {
	// 		fmt.Println("err :", err)
	// 		os.Exit(-1)
	// 	}

	// 	res.Bytes = append(res.Bytes, v[0])
	// }
	// fmt.Println(res.Bytes) //v[len(v)-2:])
	// fmt.Println(res)

}
