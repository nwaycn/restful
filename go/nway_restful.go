package main

import (
	"database/sql"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/astaxie/beego/config"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	_ "time"
)

/*
接口模型
点击呼叫的模型

curl -i -u 1000:1234  -d '{"DialNumber":"18621575908"} http://192.168.1.140:8080/clickdial

呼叫事件接口
curl -i -u admin:admin http://127.0.0.1:8080/events
返回值：
说明：event:1=呼入;2=接通;3=挂机;4=呼出
[
  {
    "id": "1",
    "aleg_number": "18621575908",
	"bleg_number": "1002",
	"route_number": "50632345",
	"event_id": "1",
	"event_time": "2014-12-1 12:00:00:343",

  },
  {
    "id": "2",
    "aleg_number": "18621575908",
	"bleg_number": "1002",
	"route_number": "50632345",
	"event_id": "2",
	"event_time": "2014-12-1 12:00:05:343",

  },
  {
    "id": "3",
    "aleg_number": "18621575908",
	"bleg_number": "1002",
	"route_number": "50632345",
	"event_id": "3",
	"event_time": "2014-12-1 12:00:10:343",

  }
]
*/
var DB *sql.DB

func check_user(username string, password string) bool {
	var result int64
	//var strsql string
	strsql := "SELECT id FROM call_extension where extension_number='" + username + "' and extension_pswd='" + password + "'"

	fmt.Println("before check")
	rows, err := DB.Query(strsql)
	if err != nil {
		//log.Fatal("failed to scan", err)
		fmt.Println("failed to check_user")
		fmt.Println(err)
		return false
	}
	fmt.Println("after check")
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			//log.Fatal("failed to scan", err)
			fmt.Println("failed to check_user")
			fmt.Println(err)
			return false
		}
		if result > 0 {
			return true
		} else {
			return false
		}
		//fmt.Println(result)
	}
	return false
}
func main() {
	//加载配置文件
	//db = nil
	iniconf, errc := config.NewConfig("ini", "restconf.conf")
	var result int64
	if errc != nil {
		//log.Fatal(errc)
		fmt.Println("config file fatal")
	}
	dbstring := iniconf.String("database")
	fmt.Println(dbstring)

	db, errd := sql.Open("postgres", dbstring)
	if errd != nil {
		//log.Fatal(errd)
		fmt.Println("database open fatal")
		fmt.Println(errc)
	}
	rows, errd := db.Query("select id from test")
	if errd != nil {
		//log.Fatal(errd)
		fmt.Println(errd)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			log.Fatal("failed to scan", err)
		}
		fmt.Println(result)
	}
	DB = db
	DB.SetMaxIdleConns(5)
	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&rest.AuthBasicMiddleware{
				Realm: "www.nway.com.cn",
				Authenticator: func(userId string, password string) bool {
					fmt.Println("check user")
					fmt.Println(userId)
					fmt.Println(password)
					//if userId == "admin" && password == "admin" {
					check_res := check_user(userId, password)
					//if check_res {
					//	fmt.Println(r.Env["REMOTE_USER"])
					//}
					return check_res
				},
			},
		},
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
		&rest.Route{"POST", "/clickdial", ClickDial},
		&rest.Route{"GET", "/events", GetNwayEvents},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8078", &handler))
}

type Country struct {
	Code string
	Name string
}

/*"id": "3",
    "aleg_number": "18621575908",
	"bleg_number": "1002",
	"route_number": "50632345",
	"event_id": "3",
	"event_time": "2014-12-1 12:00:10:343",*/
type NwayEvent struct {
	AlegNumber  string
	BlegNumber  string
	RouteNumber string
	EventId     int
	EventTime   string
}

/*
可用
 curl -i -u 1000:1234  http://192.168.1.140:8080/events
*/
func GetNwayEvents(w rest.ResponseWriter, r *rest.Request) {
	//authHeader := r.Header.Get("Authorization")
	userId := r.Env["REMOTE_USER"]
	fmt.Println(userId)
	//var strsql string
	//strsql := "'" + userId + "'"

	fmt.Println("before check")
	rows, err := DB.Query("SELECT  a.id, a.aleg_number, a.bleg_number, a.router_number::text, a.event_id, a.event_time::text FROM call_in_out_event a ,call_extension b where a.is_read=False and a.extension_id =b.id and b.extension_number=$1", userId)
	if err != nil {
		//log.Fatal("failed to scan", err)
		fmt.Println("failed to NwayEvents")
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("after check")
	NwayEvents := make([]NwayEvent, 0)
	defer rows.Close()
	var id int64
	//var my sql.NullString
	var alegNumber string
	var blegNumber string
	var routeNumber sql.NullString
	var eventId int
	var eventTime string
	for rows.Next() {

		fmt.Println("Sacn to NwayEvents")
		err := rows.Scan(&id, &alegNumber, &blegNumber, &routeNumber, &eventId, &eventTime)
		if err != nil {
			fmt.Println(err.Error())
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stmt, err := DB.Prepare("update call_in_out_event set is_read=True where id=$1")
		if err != nil {
			fmt.Println(err.Error())
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			fmt.Println(err.Error())
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var ne NwayEvent
		ne.AlegNumber = alegNumber
		ne.BlegNumber = blegNumber
		ne.EventTime = eventTime
		ne.EventId = eventId
		ne.RouteNumber = routeNumber.String
		NwayEvents = append(NwayEvents, ne)
		//fmt.Println(result)
	}
	w.WriteJson(
		NwayEvents,
	)
}

type ClickD struct {
	DialNumber string
}

/*
 可用的
curl -i -u 1000:1234 -H 'content-type: application/json' -d '{"DialNumber":"18621575908"}' http://192.168.1.140:8080/clickdial
*/
func ClickDial(w rest.ResponseWriter, r *rest.Request) {
	//authHeader := r.Header.Get("Authorization")
	fmt.Println("ClickDial")
	clickD := ClickD{}
	err := r.DecodeJsonPayload(&clickD)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if clickD.DialNumber != "" {
		userId := r.Env["REMOTE_USER"]
		fmt.Println(clickD.DialNumber)
		fmt.Println(userId)
		//here to save the number
		stmt, err := DB.Prepare("INSERT INTO call_click_dial(caller_number, is_agent, is_immediately, trans_number, account_number)values($1,True,True,$2,'')")
		//strsql := "INSERT INTO'" + clickD.DialNumber + "',True,True,'" + userId + "','')"
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(clickD.DialNumber, userId)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteJson(
			[]ClickD{
				ClickD{
					DialNumber: "success",
				},
			},
		)
	} else {
		w.WriteJson(
			[]ClickD{
				ClickD{
					DialNumber: "no_dial_number",
				},
			},
		)
	}

}

//curl -i -u 1000:1234 http://192.168.1.140:8080/countries

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
	w.WriteJson(
		[]NwayEvent{
			NwayEvent{
				//id:           1,
				AlegNumber:  "1001",
				BlegNumber:  "18621575408",
				RouteNumber: "7362372",
				//event_id:     1,
				EventTime: "2014-11-10 12:01:09",
			},
			NwayEvent{
				//id:           1,
				AlegNumber:  "1001",
				BlegNumber:  "18621575408",
				RouteNumber: "7362372",
				//event_id:     1,
				EventTime: "2014-11-10 12:01:33",
			},
		},
	)
}

////////////////////////////////////////////////////////////
//语音代码，或者说语音播报，由以下接口完成
////////////////////////////////////////////////////////////
type NwayVoiceCode struct {
	CallNumber    string
	PlayTimes     int
	VoiceCode     string
	SourceWebsite string
}

/*
func voice_code(w rest.ResponseWriter, r *rest.Request) {

	voiceCode := NwayVoiceCode{}
	err := r.DecodeJsonPayload(&voiceCode)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if voiceCode.CallNumber != "" {
		userId := r.Env["REMOTE_USER"]
		fmt.Println(voiceCode.CallNumber)
		fmt.Println(userId)

		stmt, err := DB.Prepare("INSERT INTO call_voice_verify_code(call_number, play_times, call_out_code,source_website )values($1,$2,$3,$4)")
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(clickD.CallNumber, userId)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteJson(
			[]ClickD{
				ClickD{
					DialNumber: "success",
				},
			},
		)
	} else {
		w.WriteJson(
			[]ClickD{
				ClickD{
					DialNumber: "no_dial_number",
				},
			},
		)
	}

}*/
