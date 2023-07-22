package main

import (
	"WebScrapper/scrapper"
	"github.com/gin-gonic/gin"
	"net/http"
)

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

func handleScrape(c *gin.Context) {
	term := c.PostForm("term")
	scrapper.Scrape(term)

	// Set a cookie with the success message
	c.SetCookie("flash", "Scraping completed successfully!", 3600, "/", "", false, false)

	// Redirect to the home page
	c.Redirect(http.StatusMovedPermanently, "/")
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", indexHandler)
	r.POST("/scrape", handleScrape)
	r.Run()
}
