package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/mitsuse/pushbullet-go"
	"github.com/mitsuse/pushbullet-go/requests"
	"github.com/pascalw/go-alfred"
)

type pushArgs struct {
	Iden  string
	Title string
	Body  string
}

type pushSettings struct {
	Token string
}

var action string
var query string
var pbToken string
var settings *pushSettings
var pb *pushbullet.Pushbullet
var workflowDataPath string
var settingsPath string

func init() {
	flag.StringVar(&pbToken, "token", "", "Set your Pushbullet API token")
	flag.StringVar(&action, "action", "list", "Which action to take (list/push)")
	flag.Parse()

	query = strings.Join(flag.Args(), " ")

	workflowDataPath = os.Getenv("alfred_workflow_data")

	if workflowDataPath == "" {
		panic("alfred_workflow_data not set")
	}

	settingsPath = workflowDataPath + "/settings"

	if pbToken == "" {
		loadSettings()
	}

	pb = pushbullet.New(pbToken)
}

func main() {
	switch action {
	case "list":
		list()
	case "push":
		push()
	case "settoken":
		setToken()
	default:
		panic("Unknown action")
	}
}

func list() {
	devices, err := pb.GetDevices()
	if err != nil {
		panic(err)
	}

	response := alfred.NewResponse()

	ipadRegexp, _ := regexp.Compile("(?i:iPad)")
	macbookRegexp, _ := regexp.Compile("(?i:MacBook)")

	for _, device := range devices {
		var deviceIcon string

		switch device.Type {
		case "ios":
			if ipadRegexp.Match([]byte(device.Model)) {
				deviceIcon = "icons/ionicons/ipad.png"
			} else {
				deviceIcon = "icons/ionicons/iphone.png"
			}
		case "mac":
			if macbookRegexp.Match([]byte(device.Model)) {
				deviceIcon = "icons/ionicons/laptop.png"
			} else {
				deviceIcon = "icons/ionicons/monitor.png"
			}
		case "android":
			deviceIcon = "icons/ionicons/social-android-outline.png"
		case "pc", "windows":
			deviceIcon = "icons/ionicons/social-windows-outline"
		case "opera":
			deviceIcon = "icons/devicons/opera.png"
		case "chrome":
			deviceIcon = "icons/devicons/chrome.png"
		case "firefox":
			deviceIcon = "icons/devicons/firefox.png"
		case "safari":
			deviceIcon = "icons/devicons/safari.png"
		default:
			deviceIcon = "icons/ionicons/earth.png"
		}

		args := &pushArgs{
			Iden:  device.Iden,
			Title: "Pushed from Alfred",
			Body:  query,
		}

		json, err := json.Marshal(args)
		if err != nil {
			panic(err)
		}

		hexArgs := hex.EncodeToString(json)

		response.AddItem(&alfred.AlfredResponseItem{
			Valid:    true,
			Uid:      device.Iden,
			Title:    device.Nickname,
			Subtitle: device.Manufacturer + " " + device.Model,
			Arg:      hexArgs,
			Icon:     deviceIcon,
		})
	}

	// Print the resulting Alfred Workflow XML
	response.Print()
}

func push() {
	fargs := flag.Args()

	jsonArgs, err := hex.DecodeString(fargs[0])
	if err != nil {
		panic(err)
	}

	args := &pushArgs{}
	err = json.Unmarshal([]byte(jsonArgs), &args)
	if err != nil {
		panic(err)
	}

	urlRegexp, _ := regexp.Compile("(?i:https?://)")

	if urlRegexp.Match([]byte(args.Body)) {
		l := requests.NewLink()
		l.DeviceIden = args.Iden
		l.Title = args.Title
		l.Body = ""
		l.Url = args.Body

		_, err = pb.PostPushesLink(l)
	} else {
		n := requests.NewNote()
		n.DeviceIden = args.Iden
		n.Title = args.Title
		n.Body = args.Body

		_, err = pb.PostPushesNote(n)
	}

	if err != nil {
		panic(err)
	}
}

func setToken() {
	settings = &pushSettings{
		Token: pbToken,
	}

	json, err := json.Marshal(settings)
	if err != nil {
		panic(err)
	}

	hexSettings := hex.EncodeToString(json)

	err = os.MkdirAll(workflowDataPath, 0744)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(settingsPath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.Write([]byte(hexSettings))

	f.Sync()
}

func loadSettings() {
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		response := alfred.NewResponse()
		response.AddItem(&alfred.AlfredResponseItem{
			Valid:    false,
			Uid:      "SettingsNotFound",
			Title:    "API token not set",
			Subtitle: "Use 'set-push-token' to set the API token",
			Icon:     "icons/ionicons/alert-circled.png",
		})
		response.Print()
		os.Exit(1)
	}

	hexData, err := ioutil.ReadFile(settingsPath)

	jsonData, err := hex.DecodeString(string(hexData))
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(jsonData), &settings)
	if err != nil {
		panic(err)
	}

	pbToken = settings.Token
}
