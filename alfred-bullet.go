package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
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

var action string
var query string
var pbToken string
var pb *pushbullet.Pushbullet

func init() {
	flag.StringVar(&action, "action", "list", "Which action to take (list/push)")
	flag.Parse()

	query = strings.Join(flag.Args(), " ")

	pbToken = "idezkme6sQ73laTbv9ENSxleLUxnbJYO"
	pb = pushbullet.New(pbToken)
}

func main() {
	switch action {
	case "list":
		list()
	case "push":
		push()
	default:
		panic("Unknown action")
	}
}

func list() {
	devices, err := pb.GetDevices()
	if err != nil {
		panic(err)
	}

	// optimize query terms for fuzzy matching
	// alfred.InitTerms(queryTerms)

	// create a new alfred workflow response
	response := alfred.NewResponse()
	// repos := getRepos()

	ipadRegexp, _ := regexp.Compile("(?i:iPad)")
	macbookRegexp, _ := regexp.Compile("(?i:MacBook)")

	for _, device := range devices {
		// check if the repo name fuzzy matches the query terms
		// if !alfred.MatchesTerms(queryTerms, repo.Name) {
		// 	continue
		// }

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
	json.Unmarshal([]byte(jsonArgs), &args)

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
