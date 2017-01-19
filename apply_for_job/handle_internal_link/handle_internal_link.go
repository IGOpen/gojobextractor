package handle_internal_link

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/remotejob/gojobextractor/apply_for_job/handle_internal_link/coverletter"
	"github.com/remotejob/gojobextractor/apply_for_job/handle_internal_link/mytags"
	"github.com/remotejob/gojobextractor/dbhandler"
	"github.com/remotejob/gojobextractor/domains"
	"github.com/tebeka/selenium"
	"gopkg.in/mgo.v2"
)

type InternalJobOffer struct {
	Id           string
	Company      string
	Title        string
	Location     string
	Tags         []string
	Externallink string
	Email        string
	Hits         int
	Created_at   time.Time
	Applied      bool
	Description  string
}

func NewInternalJobOffers(job domains.JobOffer) *InternalJobOffer {

	return &InternalJobOffer{

		Id:           job.Id,
		Company:      job.Company,
		Title:        job.Title,
		Location:     job.Location,
		Tags:         job.Tags,
		Externallink: job.Externallink,
		Email:        job.Email,
		Hits:         job.Hits,
		Created_at:   job.Created_at,
		Applied:      job.Applied,
		Description:  job.Description,
	}

}

func (jo *InternalJobOffer) Apply_headless(dbsession mgo.Session, page selenium.WebDriver, link string, cvpdf string) bool {

	reCaph := false

	page.Get(link)
	time.Sleep(time.Millisecond * 2500)
	jobdetails, err := page.FindElement(selenium.ByClassName, "jobdetail")
	if err != nil {

		fmt.Println(err.Error())
		if strings.HasPrefix(err.Error(), "no such element") {
			// if err.Error() == "no such element" {

			fmt.Println("Check for Page not found 1")

			_, err := page.FindElement(selenium.ByID, "jobs-not-found")
			if err != nil {
				fmt.Println(err.Error())
			} else {

				fmt.Println("It's Page not found!! make Applie True an contunue")
				jo.Applied = true
				jo.UpdateApplyedEmployer(dbsession)

			}

		}

	} else {
		time.Sleep(time.Millisecond * 2500)

		alllinks, err := jobdetails.FindElements(selenium.ByTagName, "a")
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("Check for Page not found 2")

		}
		count_links := len(alllinks)

		fmt.Println("count_links", count_links)
		var idtoapply int
		idtoapply = 0

		var applybtm []selenium.WebElement

		for i := 0; i < count_links; i++ {

			if data_jobid, err := alllinks[i].GetAttribute("data-jobid"); err == nil {

				text, _ := alllinks[i].Text()
				id, _ := alllinks[i].GetAttribute("id")

				if text == "apply now" && id == "apply" {
					idtoapply = i
					fmt.Println("apply id", idtoapply, data_jobid)

					applybtm = append(applybtm, alllinks[i])

				}

			}

			if href, err := alllinks[i].GetAttribute("href"); err == nil {
				if strings.HasPrefix(href, "mailto:") {

					emailtxt, _ := alllinks[i].Text()
					jo.Email = emailtxt
				}

			}

		}
		if idtoapply > 0 {
			time.Sleep(time.Millisecond * 1000)
			reCaph = jo.ElaborateFrame_headless(dbsession, page, applybtm[0], cvpdf)

		} else {

			fmt.Println("Can't find apply link id->", link)

			file, err := os.OpenFile("cant_apply.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			if _, err = file.WriteString(link + "\n"); err != nil {
				panic(err)
			}

		}

	}

	return reCaph
}

func (jo *InternalJobOffer) ElaborateFrame_headless(dbsession mgo.Session, page selenium.WebDriver, link selenium.WebElement, cvpdf string) bool {

	reCaph := false
	time.Sleep(1000 * time.Millisecond)

	link.Click()
	time.Sleep(6000 * time.Millisecond)

	if form, err := page.FindElement(selenium.ByID, "apply-dialog"); err == nil {

		if frms, err := form.FindElements(selenium.ByTagName, "iframe"); err == nil {

			log.Println("Check recapha if 1", len(frms))

			if len(frms) == 0 {

				if applydialog_style, err := form.GetAttribute("style"); err == nil {

					if strings.HasPrefix(applydialog_style, "display: none") {

						fmt.Println(" need move Up and try again")

						fmt.Println(link.Location())
						linkloc, _ := link.Location()
						x := linkloc.X
						y := linkloc.Y
						rawscript := "scroll(" + strconv.Itoa(y) + "," + strconv.Itoa(x) + ")"
						fmt.Println(rawscript)
						args := []interface{}{}
						page.ExecuteScriptRaw(rawscript, args)

						time.Sleep(2000 * time.Millisecond)
						if err := link.Click(); err != nil {

							fmt.Println("error clicking ", err.Error())
							fmt.Println(link.Location())
							if err := link.Click(); err != nil {
								fmt.Println("SECOND error clicking!! ", err.Error())

							}

						} else {
							fmt.Println("Click on link OK")

						}

						time.Sleep(1000 * time.Millisecond)

					}

					if allinputs, err := form.FindElements(selenium.ByTagName, "input"); err == nil {

						fmt.Println("allinputs", len(allinputs))

						if len(allinputs) == 12 {
							for _, input := range allinputs {

								if type_atr, err := input.GetAttribute("type"); err == nil {
									if type_atr == "file" {
										input.SendKeys(cvpdf)
										time.Sleep(2000 * time.Millisecond)

									}

								}

							}

							mytagstoinsert := mytags.GetMyTags("mytags.csv", jo.Tags)
							coverlettertxt := coverletter.Create(mytagstoinsert, "coverletter.csv")

							if coverletter, err := form.FindElement(selenium.ByID, "CoverLetter"); err == nil {

								log.Println("ready submit!!")
								coverletter.SendKeys(coverlettertxt)

								time.Sleep(2000 * time.Millisecond)

								if submitbtm, err := form.FindElement(selenium.ByID, "apply-submit"); err == nil {

									// fmt.Printf("Please enter an integer: ")

									// // Read in an integer
									// var i int
									// _, err := fmt.Scanln(&i)
									// if err != nil {
									// 	fmt.Printf("Error: %s", err.Error())

									// 	// If int read fails, read as string and forget
									// 	var discard string
									// 	fmt.Scanln(&discard)
									// 	// return
									// }

									log.Println("Try submit!!")
									time.Sleep(3000 * time.Millisecond)
									submitbtm.Submit()

									// fmt.Printf("Please enter an integer: ")

									// // Read in an integer
									// var i int
									// _, err := fmt.Scanln(&i)
									// if err != nil {
									// 	fmt.Printf("Error: %s", err.Error())

									// 	// If int read fails, read as string and forget
									// 	var discard string
									// 	fmt.Scanln(&discard)
									// 	// return
									// }

									// jo.Applied = true
									// jo.UpdateApplyedEmployer(dbsession)

								}
								time.Sleep(1000 * time.Millisecond)

							}

						} else {

							fmt.Println("!!!Input num not ==12")

						}

					} else {

						fmt.Println(err.Error())
					}
				}
			} else {

				log.Println("recapha PRESENT", len(frms))
				reCaph = true

			}

		}

	} else {

		fmt.Println(err.Error())
	}

	return reCaph

}

func (jo *InternalJobOffer) UpdateApplyedEmployer(dbsession mgo.Session) {

	applyedemployer := domains.JobOffer{
		Id:           jo.Id,
		Company:      jo.Company,
		Title:        jo.Title,
		Location:     jo.Location,
		Tags:         jo.Tags,
		Externallink: jo.Externallink,
		Email:        jo.Email,
		Hits:         jo.Hits,
		Created_at:   jo.Created_at,
		Applied:      jo.Applied,
		Description:  jo.Description,
	}

	dbhandler.UpdateEmployer(dbsession, applyedemployer)

}
