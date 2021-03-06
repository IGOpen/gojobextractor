package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/remotejob/gojobextractor/apply_for_job/handle_internal_link"
	"github.com/remotejob/gojobextractor/signup/accounts"

	"github.com/remotejob/gojobextractor/domains"
	"github.com/remotejob/gojobextractor/elasticLoader/loader/dbhandler"
	"github.com/tebeka/selenium"
	"gopkg.in/gcfg.v1"
	"gopkg.in/mgo.v2"
)

var login string
var pass string
var addrs []string
var database string
var username string
var password string
var mechanism string
var cvpdf string
var accountstodo [][]string

func init() {

	var cfg domains.ServerConfig
	if err := gcfg.ReadFileInto(&cfg, "config.gcfg"); err != nil {
		log.Fatalln(err.Error())

	} else {

		// login = cfg.Login.Slogin
		// pass = cfg.Pass.Spass
		addrs = cfg.Dbmgo.Addrs
		database = cfg.Dbmgo.Database
		username = cfg.Dbmgo.Username
		password = cfg.Dbmgo.Password
		mechanism = cfg.Dbmgo.Mechanism
		// cvpdf = cfg.Cvpdf.File
		cvpdf = "/tmp/mazurov_cv.pdf"
	}

	accountstodo = accounts.GetCsv("accounts.csv")

}

func main() {

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:     addrs,
		Timeout:   60 * time.Second,
		Database:  database,
		Username:  username,
		Password:  password,
		Mechanism: mechanism,
	}

	dbsession, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		panic(err)
	}
	defer dbsession.Close()

	results := dbhandler.FindNotApplyedEmployers(*dbsession)

	fmt.Println("implouers to apply", len(results))

	if len(results) > 0 {

		caps := selenium.Capabilities{"browserName": "chrome"}
		//				caps := selenium.Capabilities{"browserName": "phantomjs"}
		wd, err := selenium.NewRemote(caps, "")
		if err != nil {
			fmt.Println(err.Error())
		}
		defer wd.Quit()

		for _, account := range accountstodo {

			if !strings.HasPrefix(account[0], "#") {

				login = account[0]
				pass = account[1]

				wd.Get("https://stackoverflow.com/users/login?ssrc=head&returnurl=http%3a%2f%2fstackoverflow.com%2fjobs")

				time.Sleep(time.Millisecond * 1500)

				elem, err := wd.FindElement(selenium.ByID, "email")
				if err != nil {
					fmt.Println(err.Error())
				}
				spass, err := wd.FindElement(selenium.ByID, "password")
				if err != nil {
					fmt.Println(err.Error())
				}
				time.Sleep(time.Millisecond * 1000)

				err = elem.SendKeys(login)
				if err != nil {
					fmt.Println(err.Error())
				}
				err = spass.SendKeys(pass)
				if err != nil {
					fmt.Println(err.Error())
				}
				btm, err := wd.FindElement(selenium.ByID, "submit-button")
				if err != nil {
					fmt.Println(err.Error())
				}
				btm.Click()
				time.Sleep(time.Millisecond * 4000)

				// alllinks, err := wd.FindElements(selenium.ByTagName, "a")
				// if err != nil {
				// 	fmt.Println(err.Error())
				// 	fmt.Println("Check for Another Way to login")

				// }
				// count_links := len(alllinks)

				// log.Println(count_links)

				for i := 0; i < len(results); i++ {
					// for i := 0; i < 40; i++ {

					fmt.Println("id", results[i].Id)

					employer := handle_internal_link.NewInternalJobOffers(results[i], login)
					reCaph := (*employer).Apply_headless(*dbsession, wd, results[i].Id, cvpdf)

					if reCaph {

						log.Println("ReCaph Present Stop loop")
						break

					} else {
						log.Println("ReCaph NOT present Continue loop")
					}

				}

				wd.Quit()
			}

		}

	}
}
