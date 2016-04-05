package main

import "github.com/ArchieT/3manchess/client"
import "github.com/ArchieT/3manchess/multi"
import "github.com/ArchieT/3manchess/client/remote"
import "github.com/ArchieT/3manchess/ai/constsitval"
import "github.com/ArchieT/3manchess/player"
import "flag"
import "os"
import "log"

func main() {
	bu := flag.String("baseurl", os.Getenv("CHESSBASEURL"), "3manchess/multi base URL")
	login := flag.String("login", "remote", "login")
	passwd := flag.String("passwd", "remote", "passwd")
	//	name := flag.String("name", "", "if you want to sign up")
	botid := flag.Int64("botid", -1, "botid")
	flag.Parse()
	c := client.NewClient(nil, *bu)
	//	var u, p int64
	var u int64
	var a []byte
	/*
		var err error
		if len(*name) > 0 {
			l, _, err := c.SignUp(multi.SignUpPost{*login, *passwd, *name})
			if err == nil {
				u, p, a = l.User, l.Player, l.Auth
			}
		}
		if len(*name) == 0 || err != nil {
			log.Println(err)
	*/
	ll, _, err := c.LogIn(multi.LoggingIn{*login, *passwd})
	if err != nil {
		log.Fatalln(err)
	}
	u, a = ll.ID, ll.AuthKey
	//ppp := c.UserInfo(u)
	//p = ppp.Player
	//}
	//log.Println(u, p, a)
	log.Println(u, a)
	log.Println("botid", *botid)
	bbb, _, err := c.BotKey(multi.BotKeyGetting{*botid, multi.Authorization{u, a}})
	if err != nil {
		log.Fatal(err)
	}
	binf, _, err := c.BotInfo(*botid)
	log.Println(binf)
	if err != nil {
		log.Fatal(err)
	}
	bbp, bba := bbb.ID, bbb.AuthKey
	echn := make(chan error)
	endchn := make(chan bool)
	go func() {
		for u := range echn {
			log.Println(u)
		}
	}()
	var aii player.Player
	if binf.WhoAmI[:11] == []byte("constsitval") {
		aii = new(constsitval.AIPlayer)
	} else {
		log.Fatal(binf)
	}
	//load aiconf
}
