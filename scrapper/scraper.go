package scrapper

import (
	"encoding/csv"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct {
	name        string
	title       string
	location    string
	career      string
	education   string
	homepageURL string
}

func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=" + term
	c := make(chan []extractedJob)
	var jobs []extractedJob
	totalPages := getPages(baseURL)
	for i := 0; i < totalPages; i++ {
		go getPage(i, c, baseURL)
	}
	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
}

func getPage(pageNo int, mainC chan<- []extractedJob, baseURL string) {
	c := make(chan extractedJob)
	pageUrl := baseURL + "&recruitPageCount=" + strconv.Itoa((pageNo+1)*20)

	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	contents := doc.Find(".area_job")
	crops := doc.Find(".corp_name")
	var companyNames []string
	var HompageURLs []string
	var jobs []extractedJob

	crops.Each(func(_ int, homepage *goquery.Selection) {
		uri, _ := homepage.Find("A").Attr("href")
		url := "https://www.saramin.co.kr"
		url += strings.TrimSpace(uri)
		HompageURLs = append(HompageURLs, url)

		companyNames = append(companyNames, strings.TrimSpace(homepage.Find("a").Text()))
	})

	contents.Each(func(i int, card *goquery.Selection) {
		go doExtractJob(card, c, companyNames[i], HompageURLs[i])
	})

	for i := 0; i < contents.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func doExtractJob(card *goquery.Selection, c chan<- extractedJob, name string, url string) {
	title := card.Find("a").AttrOr("title", "") // Extract the title attribute from the <a> tag within the card
	var locationContents []string
	location := card.Find(".job_condition")
	location.Contents().Map(func(_ int, s *goquery.Selection) string {
		if strings.TrimSpace(s.Text()) != "" {
			locationContents = append(locationContents, strings.TrimSpace(s.Text()))
			return ""
		} else {
			return ""
		}
	})
	place := locationContents[0]
	career := locationContents[1]
	education := locationContents[2]
	companyname := name
	companyurl := url

	c <- extractedJob{
		name:        companyname,
		title:       title,
		location:    place,
		career:      career,
		education:   education,
		homepageURL: companyurl,
	}
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Name", "Title", "Location", "Career", "Education", "HomepageURL"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.name, job.title, job.location, job.career, job.education, job.homepageURL}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func getPages(baseURL string) int {
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	return doc.Find(".page").Length()
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with ", res.StatusCode)
	}
}
