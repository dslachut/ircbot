package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/husio/go-irc/irc"
	"github.com/mitchellh/osext"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"time"
)

var settings struct {
	Server  string
	Port    int
	Channel string
	Nick    string
}

var server *string
var port *int
var nick *string
var channel *string

var approved = []string{"dslachut", "hummus", "OgreMonk", "acan", "patty"}
var acts = []string{"slaps", "brofists", "chases", "examines", "flicks",
	"hugs", "mimics", "knifes", "surveils", "pokes"}

func initialize() {
	wdir, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}

	settingsFile, err := os.Open(wdir + "\\settings.cfg")
	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(settingsFile)
	if err = jsonParser.Decode(&settings); err != nil {
		log.Fatal(err)
	}

	server = flag.String("server", settings.Server, "IRC Server Address")
	port = flag.Int("port", settings.Port, "IRC Server Port")
	nick = flag.String("nick", settings.Nick, "IRC handle")
	channel = flag.String("chan", settings.Channel, "IRC Channel")
}

func handle(msg *irc.Message, client *irc.Client) {
	fmt.Println(msg)
	nym := bytes.NewBuffer(msg.Nick()).String()
	app := false
	for _, v := range approved {
		if nym == v {
			app = true
			break
		}
	}
	if app {
		fmt.Println("Approved", nym)
		match, err := regexp.MatchString("opsplx", msg.String())
		fmt.Println(match)
		if err != nil {
			log.Print(err)
		} else if match {
			act := acts[rand.Intn(len(acts))]
			vic := approved[rand.Intn(len(approved))]
			client.Send("MODE %s +o %s", *channel, nym)
			client.Send("PRIVMSG %s :\u0001ACTION tries to give %s ops, then %s %s \u0001",
				*channel, nym, act, vic)
		}
	} else {
		match, err := regexp.MatchString("opsplx", msg.String())
		if err != nil {
			log.Print(err)
		} else if match {
			client.Send("PRIVMSG %s :No ops for you, %s!", *channel, nym)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.GOMAXPROCS(-1))
	flag.Parse()
	rand.Seed(time.Now().Unix())

	initialize()

	client, err := irc.Connect(*server, *port)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Send("NICK %s", *nick); err != nil {
		log.Fatal(err)
	}

	if err := client.Send("USER bot * * : ..."); err != nil {
		log.Fatal(err)
	}

	if err := client.Send("JOIN %s", *channel); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-client.Error:
			log.Fatal(err)
		case msg := <-client.Received:
			go handle(msg, client)
		}
	}
}
