# webmon
A POC for website monitoring application

## Setup Instructions
1. The first need is to have a system with [Go](https://golang.org/dl) installed.
2. Install [mysql](https://dev.mysql.com/doc/) database server.
3. Clone the repository.
4. Install Gin package
```sh
$ go get -u github.com/gin-gonic/gin
``` 
5. Install mysql driver
```sh
$ go get github.com/go-sql-driver/mysql
```
6. The most suitable ORM with golang is being used here. Install gorm.
```sh
$ go get -u github.com/jinzhu/gorm
```
7. Need to set up Cron to run the goroutines in the background. Setup Cron via following command
```sh
$ github.com/robfig/cron
```
8. To report Cron logs, following package is being used.
```sh
$ go get github.com/sirupsen/logrus
```
9. Need to create blank database.
10. Connect the database with code using proper mysql credentials and database name.
11. Open terminal and move to project `cd /path/to/project`
12. Execute the main function
```sh
$ go run webmon.go
```