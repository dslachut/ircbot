package main

import (
    "github.com/husio/go-irc/irc"
    "fmt"
    "flag"
    "runtime"
    "log"
    "bytes"
    "regexp"
    )

var server *string = flag.String("server", "irc.w3.org", "IRC Server Address")
var port *int = flag.Int("port", 6665, "IRC Server Port")
var nick *string = flag.String("nick", "hatchetBot", "IRC handle")
var channel *string = flag.String("chan", "#hummuschan", "IRC Channel")

var approved = []string{"dslachut","hummus","OgreMonk","acan","patty"}

func handle(msg *irc.Message, client *irc.Client) {
    fmt.Println(msg)
    nym := bytes.NewBuffer(msg.Nick()).String()
    app := false
    for _,v := range approved {
        if nym == v {
            app = true
            break
        }
    }
    if app {
        fmt.Println("Approved", nym)
        match,err := regexp.MatchString("opsplx", msg.String())
        fmt.Println(match)
        if err != nil {
            log.Print(err)
        } else if match {
            client.Send("MODE %s +o %s", *channel, nym)
            client.Send("PRIVMSG %s :\u0001ACTION tries to give %s ops, then slaps hummus \u0001", 
            *channel, nym)
        }
    }
}

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU())
    fmt.Println(runtime.GOMAXPROCS(-1))
    flag.Parse()
    
    client,err := irc.Connect(*server,*port)
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
                go handle(msg,client)
        }
    }
}
