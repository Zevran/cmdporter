package main

/* ====================================================================================================

cmdporter : a wifi intercom to talk to various devices

By Fred Ménez, Gaël Reyrol, Thierry Vo

==================================================================================================== */

/* TODO Serial

x looks for serial device depending on OS (Macos, Linux)
x discover serial device or read configuration
x load commands params from file

*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"text/template"
)

type CmdRequest struct {
	Command string `json:"command"`
}

type CmdResponse struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}

func Render(w http.ResponseWriter, view string, content interface{}) {
	layout, err := ioutil.ReadFile(path.Join("views", "layout.html"))
	if err != nil {
		log.Fatal(err)
	}
	page, err := ioutil.ReadFile(path.Join("views", view))
	if err != nil {
		log.Fatal(err)
	}

	layoutTemplate := template.New("layout")
	pageTemplate := template.New("page")

	template.Must(layoutTemplate.Parse(string(layout)))
	template.Must(pageTemplate.Parse(string(page)))

	pageBuffer := new(bytes.Buffer)
	pageTemplate.Execute(pageBuffer, content)

	layoutContent := map[string]interface{}{"View": string(pageBuffer.Bytes())}
	layoutTemplate.Execute(w, layoutContent)
}

func ParseBody(r *http.Request) []byte {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
	}

	return body
}

func main() {
	mainDevice := new(Device)
	mainDevice.Connect("/dev/tty.usb", 9600)
	mainDevice.Config("devices/vp_nec_m271_m311.json")

	// Start Http Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		content := map[string]interface{}{
			"SerialPortStatus": mainDevice.Status,
			"Device":           mainDevice,
		}

		Render(w, "index.html", content)
	})

	http.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			req := CmdRequest{}
			res := CmdResponse{nil, nil}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if err := json.Unmarshal(ParseBody(r), &req); err != nil {
				fmt.Println(err)
			}

			// Search for submited command on device
			for _, value := range mainDevice.Commands {
				if req.Command == value.Name {
					fmt.Printf("Found command : %s\n", req.Command)
					err := mainDevice.DoCommand(req.Command)
					if err != nil {
						res.Error = err.Error()
					} else {
						res.Data = "Success"
					}
					jsonRes, _ := json.Marshal(res)
					fmt.Fprintf(w, "%s", string(jsonRes))
					return
				}
			}

			// Command not found in device commands list
			res.Error = "CommandNotFound"
			jsonRes, _ := json.Marshal(res)
			fmt.Fprintf(w, "%s", string(jsonRes))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			res := CmdResponse{nil, nil}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if mainDevice.Status {
				res.Error = "DeviceAlreadyConnected"
			} else {
				err := mainDevice.Connect("/dev/tty.usb", 9600)
				if mainDevice.Status == false && err != nil {
					res.Error = "FailedConnectDevice"
				} else {
					res.Data = "DeviceConnected"
				}
			}

			jsonRes, _ := json.Marshal(res)
			fmt.Fprintf(w, "%s", string(jsonRes))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	http.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			res := CmdResponse{nil, nil}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if mainDevice.Status == false {
				res.Error = "DeviceAlreadyDisconnected"
			} else {
				err := mainDevice.Close()
				if mainDevice.Status && err != nil {
					res.Error = "FailedDisconnectDevice"
				} else {
					res.Data = "DeviceDisconnected"
				}
			}

			jsonRes, _ := json.Marshal(res)
			fmt.Fprintf(w, "%s", string(jsonRes))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	//log.Println("Running for device", g_Device.GetName())
	log.Println("Waiting for http connections on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
