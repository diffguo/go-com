package gocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/diffguo/gocom/log"
	"github.com/diffguo/gocom/trace_id"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var AuthAes = NewAesCbcPKCS7("XGYUZj78QvlvyHQ1eKeSeNhCJcJRQOyQ")

func GinLogger(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		trace_id.SaveTraceId(c.GetHeader(trace_id.TraceIDName))

		if log.GLog.LogLevel == log.LogLevelDebug {
			if c.Request.Method == http.MethodGet {
				log.Debugf("[GIN DEBUG] %s %s URL: %s Header: %+v", c.Request.Method, c.Request.Proto,
					c.Request.URL.String(), c.Request.Header)
			} else {
				contentType := c.ContentType()
				if contentType == gin.MIMEJSON || contentType == gin.MIMEHTML || contentType == gin.MIMEXML ||
					contentType == gin.MIMEXML2 || contentType == gin.MIMEPlain || contentType == gin.MIMEPOSTForm ||
					contentType == gin.MIMEMultipartPOSTForm {
					body, _ := ioutil.ReadAll(c.Request.Body)
					c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

					if body != nil {
						log.Debugf("[GIN DEBUG] %s %s URL: %s Header: %+v Body: %s", c.Request.Method, c.Request.Proto,
							c.Request.URL.String(), c.Request.Header, string(body))
					} else {
						log.Debugf("[GIN DEBUG] %s %s URL: %s Header: %+v Body err", c.Request.Method, c.Request.Proto,
							c.Request.URL.String(), c.Request.Header)
					}
				} else {
					log.Debugf("[GIN DEBUG] %s %s URL: %s Header: %+v Body err", c.Request.Method, c.Request.Proto,
						c.Request.URL.String(), c.Request.Header)
				}
			}
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		statusColor := log.ColorForStatus(statusCode)
		methodColor := log.ColorForMethod(method)
		userId := c.Request.Header.Get("selfUserId")

		requestData := getRequestData(c)
		log.Infof("[GIN] %s%s%s %s%s %s%d%s %.03f [%s] [user_id:%s] %s",
			methodColor, method, log.Reset,
			c.Request.Host, requestData,
			statusColor, statusCode, log.Reset,
			latency.Seconds(),
			clientIP,
			userId,
			c.Errors.String())

		if latency > threshold {
			log.Warnf("[GIN SLOW] %s%s%s %s%s %s%d%s %.03f [%s] [user_id:%s] startAt: %s endAt: %s",
				methodColor, method, log.Reset,
				c.Request.Host, requestData,
				statusColor, statusCode, log.Reset,
				latency.Seconds(),
				clientIP,
				userId,
				start.Format("15:04:05.999999999"),
				end.Format("15:04:05.999999999"))
		}
	}
}

func getRequestData(c *gin.Context) string {
	var requestData string
	method := c.Request.Method
	if method == "GET" || method == "DELETE" {
		requestData = c.Request.RequestURI
	} else {
		c.Request.ParseForm()
		requestData = fmt.Sprintf("%s [%s]", c.Request.RequestURI, c.Request.Form.Encode())
	}

	if len(requestData) > 1024 {
		return requestData[:1024]
	} else {
		return requestData
	}
}

type ResponseHead struct {
	ErrCode int         `json:"err_code"` // 0为正常返回，其它为异常，异常时ErrDesc为异常描述。 固定的异常描述请存储于err_code_def.go文件中
	ErrDesc string      `json:"err_desc"`
	Content interface{} `json:"content"`
	TraceId string      `json:"trace_id"` // 用于定位某次调用的ID，客户端应该在错误时显示(ErrCode:TraceId)
}

func SendSimpleResponse(c *gin.Context, content interface{}) {
	SendResponseImp(c, content, 0, "")
}

//func SendResponse(c *gin.Context, content interface{}, errCode int) {
//	if errCode == 0 {
//		SendResponseImp(c, content, 0, "")
//		return
//	}
//
//	if errCode >= ErrCodeParamErr {
//		SendResponseImp(c, content, errCode, MapErrCode2Desc[errCode])
//		return
//	}
//
//	SendResponseImp(c, content, errCode, "")
//}

func SendResponseImp(c *gin.Context, content interface{}, errCode int, errDesc string) {
	c.Writer.Header().Set("Content-Type", "application/json")
	resp := ResponseHead{
		ErrCode: errCode,
		ErrDesc: errDesc,
		Content: content,
		TraceId: trace_id.GetTraceId(),
	}

	b, err := json.Marshal(&resp)
	if err != nil {
		log.Error(err.Error())
	} else {
		c.Writer.Write(b)
	}
	//c.JSON(http.StatusOK, resp)
}

type UserAgent struct {
	AppVersion        string `json:"app_version"`
	MobilePlatform    string `json:"mobile_platform"`
	MobileSystem      string `json:"mobile_system"`
	MobileDeviceBrand string `json:"mobile_device_brand"`
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPathInWhiteList(c.Request.URL.Path) {
			// Process request
			c.Next()
			return
		}

		ua := c.GetHeader("UserAgent")
		if ua == "" {
			log.Warnf("No UserAgent in the req: %+v", c.Request.Header)
			SendResponseImp(c, "", http.StatusUnauthorized, "No UserAgent in the req")
			c.Abort()
			return
		}

		var userAgent UserAgent
		err := json.Unmarshal([]byte(ua), &userAgent)
		if err != nil {
			log.Errorf("Unmarshal UserAgent Fail: %s %s", ua, err.Error())
			SendResponseImp(c, "", http.StatusUnauthorized, "Error UserAgent")
			c.Abort()
			return
		}

		authToken := c.GetHeader("Authorization")
		tokenPlaintext, err := AuthAes.Decrypt(authToken)
		if err != nil {
			log.Errorf("Decrypt Authorization error : %s %s", authToken, err.Error())
			SendResponseImp(c, "", http.StatusUnauthorized, "Wrong Authorization")
			c.Abort()
			return
		}

		items := strings.Split(tokenPlaintext, "|")
		if len(items) != 5 {
			log.Errorf("Wrong1 Authorization: %s", tokenPlaintext)
			SendResponseImp(c, "", http.StatusUnauthorized, "Wrong1 Authorization")
			c.Abort()
			return
		}

		if items[0] != userAgent.AppVersion || items[1] != userAgent.MobileSystem || items[2] != userAgent.MobileDeviceBrand {
			log.Errorf("Wrong2 Authorization: %+v", userAgent)
			SendResponseImp(c, "", http.StatusUnauthorized, "Wrong2 Authorization")
			c.Abort()
			return
		}

		// 判断过期时间
		timeStamp, err := strconv.ParseInt(items[4], 10, 64)
		if err != nil {
			log.Error("wrong time")
			c.Abort()
			return
		}

		tokenTime := time.Unix(timeStamp, 0)
		if tokenTime.Add(time.Duration(time.Hour * 6)).Before(time.Now()) {
			SendResponseImp(c, "", http.StatusUnauthorized, "Wrong Authorization, timeout")
			log.Infof("token timeout. token time: %d", timeStamp)
			c.Abort()
			return
		}

		userId, err := strconv.ParseInt(items[4], 10, 64)
		if err != nil {
			SendResponseImp(c, "", http.StatusUnauthorized, "Wrong Authorization, Parse userId err")
			log.Error("token wrong userId")
			c.Abort()
			return
		}

		err = GenAuth(c, userId)
		if err != nil {
			log.Errorf("gen new auth err: %s", err.Error())
		}

		c.Request.Header.Set("selfUserId", items[3])
	}
}

var mWhitePathMap = map[string]Empty{
	"/favicon.ico":        empty,
	"/wx/wx_login":        empty,
	"/admin/login":        empty,
	"/wechat/wx_callback": empty,
	"/test/*":             empty,
	"/debug/pprof/*":      empty,
	"/swagger/*":          empty,
}

var mWhitePathMapLen = len(mWhitePathMap)

func isPathInWhiteList(path string) bool {
	for k, _ := range mWhitePathMap {
		if k[len(k)-1] == '*' {
			// 进行前缀匹配
			if strings.HasPrefix(path, k[0:len(k)-1]) {
				return true
			}
		} else if k == path {
			return true
		}
	}
	_, ok := mWhitePathMap[path]
	return ok
}

func GenAuth(c *gin.Context, userId int64) error {
	ua := c.GetHeader("Useragent")
	if ua == "" {
		return fmt.Errorf("No UserAgent")
	}

	var userAgent UserAgent
	err := json.Unmarshal([]byte(ua), &userAgent)
	if err != nil {
		return fmt.Errorf("Unmarshal UserAgent Fail: %s %s", ua, err.Error())
	}

	plaintext := fmt.Sprintf("%s|%s|%s|%d|%d", userAgent.AppVersion, userAgent.MobileSystem, userAgent.MobileDeviceBrand, userId, time.Now().Unix())
	out, err := AuthAes.Encrypt([]byte(plaintext))
	if err != nil {
		return err
	}

	c.Writer.Header().Set("Authorization", out)
	return nil
}

func ShowAllRawBody(c *gin.Context) {
	data, _ := ioutil.ReadAll(c.Request.Body)
	log.Infof("Raw Http Body: %s", string(data))
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Useragent,selfuserid")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin,"+
			" Access-Control-Allow-Headers, Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func Bind(c *gin.Context, obj interface{}) bool {
	err := c.Bind(obj)
	if err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		// validator 参考： https://github.com/go-playground/validator
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Errorf("bind err. InvalidValidationError: %s", err.Error())
			return false
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.Errorf("bind err. ValidationError. StructField: %s, Tag: %s %s, Type: %+v, Value: %+v", err.StructNamespace(), err.ActualTag(), err.Param(), err.Type(), err.Value())
		}

		return false
	} else {
		return true
	}
}
