/*
* Http (curl) request in golang
* @author Shashank Tiwari
*/
 
package main
 
import (
	"fmt"
	"time"
	"net/http"
	"io/ioutil"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
)

type MonitorLogs struct {
	gorm.Model
	//ID 					int 		`json:"id"`
	ReadableName 		string		`json:"readable_name"`
	Url 				string 		`json:"url"`
	MonitorDuration 	int			`json:"monitor_duration"`
	HttpStatusCheck 	int 		`json:"http_status_check"`
	HttpResponseContains string		`json:"http_response_contains"`
}

 
func main() {
 
	url := "https://admin.powertoexhaletravelbookit.org/"
	reqStart := time.Now()
	req, _ := http.NewRequest("GET", url, nil)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	reqEnd := time.Now()
	//fmt.Println("%+v\n", reqStart)
	fmt.Println(reqEnd.Sub(reqStart))
	fmt.Println(res.StatusCode)


	fmt.Println(string(body))
 
}

//&{Status:200 OK StatusCode:200 Proto:HTTP/1.1 ProtoMajor:1 ProtoMinor:1 Header:map[Content-Type:[text/html; charset=UTF-8] Date:[Mon, 02 Mar 2020 19:02:33 GMT] Server:[Apache/2.4.18 (Ubuntu)] Set-Cookie:[ci_session=a%3A5%3A%7Bs%3A10%3A%22session_id%22%3Bs%3A32%3A%2220cd501eddfda5cd5aecdcd547a26b32%22%3Bs%3A10%3A%22ip_address%22%3Bs%3A12%3A%2249.36.71.224%22%3Bs%3A10%3A%22user_agent%22%3Bs%3A18%3A%22Go-http-client%2F1.1%22%3Bs%3A13%3A%22last_activity%22%3Bi%3A1583175753%3Bs%3A9%3A%22user_data%22%3Bs%3A0%3A%22%22%3B%7Dd098dd27571152c04c5012857a0acda7; expires=Mon, 02-Mar-2020 21:02:33 GMT; Max-Age=7200; path=/] Vary:[Accept-Encoding]] Body:0xc00032c100 ContentLength:-1 TransferEncoding:[] Close:false Uncompressed:true Trailer:map[] Request:0xc0000f2000 TLS:0xc000108000}
