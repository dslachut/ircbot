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
	"runtime"
	"strings"
	"time"
)

var settings struct {
	Server   string
	Port     int
	Channel  string
	Nick     string
	Password string
	Greeting string
	Goodbye  string
	Bots     map[string][]string
}

var server *string
var port *int
var nick *string
var channel *string

var currentNicks []string
var currentBots []string

var changedCmds = []string{"KICK", "PART", "JOIN", "QUIT"}
var commands = []string{"*op", "*stop"}
var acts = []string{"slaps", "brofists", "chases", "examines", "flicks",
	"hugs", "mimics", "knifes", "surveils", "pokes"}

func initialize() {
	wdir, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}

	settingsFile, err := os.Open(wdir + string(os.PathSeparator) + "settings.cfg")
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

	fmt.Println("Approved", nym)

	act := acts[rand.Intn(len(acts))]
	vic := currentNicks[rand.Intn(len(currentNicks))]
	client.Send("MODE %s +o %s", *channel, nym)
	client.Send("PRIVMSG %s :\u0001ACTION tries to give %s ops, then %s %s \u0001",
		*channel, nym, act, vic)
}

//Returns list of nicks in channel excluding any nicks with "bot"
//Could tweak to check against slice of bots in case of non-bot nick containing "bot"
func updateNicks(msg *irc.Message) ([]string, []string) {
	nickArray := strings.Fields(string(msg.Trailing))
	outNicks := make([]string, 0)
	outBots := make([]string, 0)
	for i := 0; i < len(nickArray); i++ {
		nickArray[i] = strings.TrimPrefix(nickArray[i], "@")
		if strings.Contains(strings.ToLower(nickArray[i]), "bot") {
			outBots = append(outBots, nickArray[i])
		} else {
			outNicks = append(outNicks, nickArray[i])
		}
	}
	return outNicks, outBots
}

func isPrivMsg(msg *irc.Message) bool {
	if string(msg.Command) == "PRIVMSG" && string(msg.Params) == settings.Nick {
		return true
	} else {
		return false
	}
}

func nicksChanged(msg *irc.Message) bool {
	changed := false
	for _, v := range changedCmds {
		if v == string(msg.Command) {
			changed = true
			break
		}
	}
	return changed
}

func goodCommand(msg *irc.Message) (bool, []string) {
	cmd := false
	trailing := strings.Fields(string(msg.Trailing))
	for _, v := range commands {
		if v == trailing[0] {
			cmd = true
			break
		}
	}
	return cmd, trailing
}

func joinServer(client *irc.Client) {
	if err := client.Send("NICK %s", *nick); err != nil {
		log.Fatal(err)
	}

	if err := client.Send("USER bot * * : ..."); err != nil {
		log.Fatal(err)
	}

	if err := client.Send("JOIN %s", *channel); err != nil {
		log.Fatal(err)
	}

	//Gets names of users in channel upon first joining
	if err := client.Send("NAMES %s", *channel); err != nil {
		log.Fatal(err)
	}

	if len(currentBots) > 0 {
		target := "none"
		for _, v := range currentBots {
			if _, ok := settings.Bots[v]; ok {
				target = v
				break
			}
		}
		if target != "none" {
			if err := client.Send("PRIVMSG %s : *op %s", target, settings.Bots[target][0]); err != nil {
				log.Print(err)
			}
		} else {
			if err := client.Send("PRIVMSG %s : opsplx", *channel); err != nil {
				log.Print(err)
			}
		}
	} else {
		if err := client.Send("PRIVMSG %s : opsplx", *channel); err != nil {
			log.Print(err)
		}
	}

	//Sends greeting message
	if err := client.Send("PRIVMSG %s :%s", *channel, settings.Greeting); err != nil {
		log.Fatal(err)
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

	joinServer(client)

	for {
		select {
		case err := <-client.Error:
			log.Fatal(err)
		case msg := <-client.Received:
			switch {
			case nicksChanged(msg):
				{
					client.Send("NAMES %s", *channel)
				}
			case string(msg.Command) == "353":
				{
					currentNicks, currentBots = updateNicks(msg)
				}
			case isPrivMsg(msg):
				{
					isCmd, msgText := goodCommand(msg)
					if msgText[0] == "*op" {
						if isCmd && msgText[1] == settings.Password {
							handle(msg, client)
						} else {
							client.Send("PRIVMSG %s :No ops for you, %s!", *channel, string(msg.Nick()))
						}
					} else if msgText[0] == "*stop" {
						if isCmd && msgText[1] == settings.Password {
						    client.Send("PRIVMSG %s :%s has ordered me to shutdown. Goodbye.", *channel, string(msg.Nick()))
						    fmt.Println("%s ordered a shutdown: %s", string(msg.Nick()), time.Now())
							os.Exit(0)
						} else {
						    fmt.Println("%s attempted shutdown was denied.", string(msg.Nick()))
							client.Send("PRIVMSG %s :%s, you can't stop me!", *channel, string(msg.Nick()))
						}
					}
				}
			}
		}
	}
}
