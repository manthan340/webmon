package main

import (
	//"fmt"
	"time"
	"strings"
	"net/http"
	"io/ioutil"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"


)

var db *gorm.DB

type Website struct {
	gorm.Model
	//ID 					int 		`json:"id"`
	ReadableName 		string		`json:"readable_name"`
	Url 				string 		`gorm:"type:text"`
	MonitorDuration 	int			`json:"monitor_duration"`
	HttpStatusCheck 	int 		`json:"http_status_check"`
	HttpResponseContains string		`json:"http_response_contains"`
}

type MonitorLog struct {
	gorm.Model
	WebsiteID 			int			`json:"website_id"`
	Website 			Website
	ResponseTime 		float32 	`json:"response_time"`
	HttpStatusCode 		int 		`json:"http_status_code"`
	HttpResponse 		string		`gorm:"type:text"`
	AlertStatus			bool 		`json:"alert_status"`
	AlertReason			string 		`json:"alert_reason"`
}

func init() {
	//open a db connection
	var err error
	db, err = gorm.Open("mysql", "xxxx:xxxxxxx@tcp(localhost:3306)/webmon?parseTime=true")
	if err != nil {
		panic("failed to connect database")
	}

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	//defer db.Close()
	//Migrate the schema
	db.AutoMigrate(&Website{})
	db.AutoMigrate(&MonitorLog{})

	start_cron()
}

func main() {
	router := gin.Default()

	website := router.Group("/website")
	{
		website.POST("/", createWebsite)
		website.GET("/", fetchAllWebsites)
		//website.GET("/:id", fetchSingleWebsite)
		//website.PUT("/:id", updateWebsite)
		//website.DELETE("/:id", deleteWebsite)
	}
	router.GET("editWebsite/:id",fetchSingleWebsite)
	router.POST("updateWebsite",updateWebsite)
	router.GET("deleteWebsite/:id",deleteWebsite)
	router.GET("logs/:website_id",fetchAllLogs)
	router.GET("monitor/:website_id",monitorWebsite)
	router.LoadHTMLGlob("templates/*")
	router.Run()

}

func fetchAllWebsites(c *gin.Context) {
	//websites := []Website {}
	//db.Find(&websites)
	
	type WebsitesWithLastLog struct {
	 	ID int
		ReadableName 		string
		Url 				string	
		MonitorDuration 	int	
		HttpStatusCheck 	int	
		HttpResponseContains string
		LastAlertStatus 	bool
		LastLogTime 		*time.Time

	}

	var websitewithlastlog []WebsitesWithLastLog

	//Used Raw query here to optimize the SQL calls. The single SQL call will return the websites with the last log details. 
	db.Raw("SELECT websites.id, websites.readable_name, websites.url, websites.monitor_duration, websites.http_status_check, websites.http_response_contains, monitor_logs.alert_status as last_alert_status, monitor_logs.created_at as last_log_time FROM websites LEFT JOIN (SELECT website_id, MAX( id ) AS last_log_id FROM `monitor_logs` GROUP BY website_id)latest_log ON latest_log.website_id = websites.id LEFT JOIN monitor_logs ON latest_log.last_log_id = monitor_logs.id WHERE websites.deleted_at IS NULL").Scan(&websitewithlastlog)

	
	c.HTML(http.StatusOK, "index.html", gin.H{
			"websites" : websitewithlastlog,
			"website_details" : nil, 
	})
}

func createWebsite(c *gin.Context) {
	_MonitorDuration, _:= strconv.Atoi(c.PostForm("monitor_duration"))
	_HttpStatusCheck, _:= strconv.Atoi(c.PostForm("http_status_check"))
	website := Website {
					ReadableName: c.PostForm("readable_name"),
					Url: c.PostForm("url"),
					MonitorDuration: _MonitorDuration,
					HttpStatusCheck: _HttpStatusCheck,
					HttpResponseContains: c.PostForm("http_response_contains"),
				}
	db.Save(&website)
	//c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Website created successfully!", "resourceId": website.ID})
	c.Redirect(http.StatusMovedPermanently, "/website")


}

func fetchSingleWebsite(c *gin.Context){
	var website Website
	websiteID := c.Param("id")

	db.First(&website, websiteID)

	if website.ID == 0 {

	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"website_details" : website,
	})
}

func updateWebsite(c *gin.Context) {
	websiteID := c.PostForm("id")
	var website Website
	db.First(&website, websiteID)

	_MonitorDuration, _ := strconv.Atoi(c.PostForm("monitor_duration"))
	website.MonitorDuration = _MonitorDuration
	_HttpStatusCheck, _ := strconv.Atoi(c.PostForm("http_status_check"))
	website.HttpStatusCheck = _HttpStatusCheck
	website.ReadableName = c.PostForm("readable_name")
	website.Url = c.PostForm("url")
	website.HttpResponseContains = c.PostForm("http_response_contains")
				
	db.Save(&website)

	c.Redirect(http.StatusMovedPermanently, "/website")
}

func deleteWebsite(c *gin.Context) {
	websiteID := c.Param("id")	

	var website Website
	db.First(&website, websiteID)

	db.Delete(website)
	c.Redirect(http.StatusMovedPermanently, "/website")

}

/******************************************/


func fetchAllLogs(c *gin.Context) {
	logs := []MonitorLog {}
	websiteID := c.Param("website_id")


	db.Where("website_id = ?", websiteID).Order("id desc").Find(&logs)

	c.HTML(http.StatusOK, "logs.html", gin.H{
			"logs" : logs,
	})

}

func monitorWebsite(c *gin.Context) {
	websiteID, _ := strconv.Atoi(c.Param("website_id"))

	var website Website
	db.First(&website, websiteID)

	log := scan_website(website)

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "log saved successfully!", "resourceId": log.ID})
	//c.Redirect(http.StatusMovedPermanently, "/website")


}

func scan_website(website Website)(monitor MonitorLog) {

	_ResponseTime, _HttpStatusCode, _HttpResponse := http_call(website.Url)

	_AlertStatus := false
	var _AlertReason string
	if _HttpResponse == "" && _HttpStatusCode == 404 && _ResponseTime == 0 {
		_AlertStatus = true
		_AlertReason = "URL not reachable"
	} else if website.HttpStatusCheck != _HttpStatusCode {
		_AlertStatus = true
		_AlertReason = "HTTP Status code not matching with "+ strconv.Itoa(website.HttpStatusCheck)
	}else if website.HttpResponseContains != "" && !strings.Contains(_HttpResponse, website.HttpResponseContains)  {
		_AlertStatus = true
		_AlertReason = "Response does not contains the keyword"
	}


	monitor = MonitorLog {
					WebsiteID: int(website.ID),
					ResponseTime: float32(_ResponseTime),
					HttpStatusCode: _HttpStatusCode,
					HttpResponse: _HttpResponse,
					AlertStatus: _AlertStatus,
					AlertReason: _AlertReason,
				}
	db.Save(&monitor)
	return
	
}


func http_call(url string)(response_time float64, http_status_code int, http_response string) {

	reqStart := time.Now()
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		//panic("failed to monitor URL")
		response_time = 0
		http_status_code = 404
		http_response = ""
		return	
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	reqEnd := time.Now()
	//fmt.Println("%+v\n", reqStart)
	response_time = (reqEnd.Sub(reqStart)).Seconds()
	http_status_code = res.StatusCode
	http_response = string(body)
	return	
}

func printCronEntries(cronEntries []cron.Entry) {
	log.Infof("Cron Info: %+v\n", cronEntries)
}

func start_cron() {

	c := cron.New()
	c.AddFunc("*/1 * * * *", start_all_website_monitoring)

	// Start cron with one scheduled job
	log.Info("Start cron")
	c.Start()
	printCronEntries(c.Entries())
	
}

func start_all_website_monitoring(){
	minutes := time.Now().Minute()
	min:=strconv.Itoa(minutes)
	log.Info("[Job 1]Current minute: "+min+"\n")
	
	websites := []Website {}
	db.Where(min+" MOD monitor_duration = 0").Find(&websites)

	for i:=0; i<len(websites);i++ {
		log.Info("Matched: "+websites[i].ReadableName)
		monitorlog:=scan_website(websites[i])
		logid := strconv.Itoa(int(monitorlog.ID))
		log.Info(" -> Monitor Log Added: "+logid+"\n")
	}
}