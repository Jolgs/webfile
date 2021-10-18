package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)
func init() {
	Setuplog()
}
var Txt =[]string{"txt","log"}
var Zip =[]string{"zip","tar"}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")  // 可将将 * 替换为指定的域名
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func Ginlog() gin.HandlerFunc {
		return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

			// 你的自定义格式
			return fmt.Sprintf("[%s] - %s %s %s %d  %s \n",
				param.TimeStamp.Format(time.RFC3339),
				param.ClientIP,
				param.Method,
				param.Path,
				//param.Request.Proto,
				param.StatusCode,
				//param.Latency,
				//param.Request.UserAgent(),
				param.ErrorMessage,
			)
		})
}

var Log *log.Logger

func checkdir( dir string)  {
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println(err.Error())
	}
}


func Setuplog(){
	logFilePath := ""
	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}
	if err := os.MkdirAll(logFilePath, 0755); err != nil {
		fmt.Println(err.Error())
	}
	logFileName := "WebFile.log"
	//日志文件
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println(err.Error())
		}
	}
	Log =log.New("")
	Log.SetLevel(1)
	Log.SetOutput(io.MultiWriter(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}, os.Stdout))
	Log.SetHeader("${time_rfc3339} || ${level} [${short_file}:${line}]")
}
func ExternalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}


func in(target string, str_array []string) bool {
	for _, element := range str_array{
		if target == element{
			return true
		}
	}
	return false
}
type ListFiles struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Cpath string `json:"cpath"`
	Filetype string `json:"filetype"`
}
func Index(c *gin.Context) {
	cpath := c.DefaultQuery("path","/")
	lm := make([]ListFiles, 0)
	//遍历目录，读出文件名、大小
	//Log.Info(path.Join(Storage, cpath))
	f, err := os.Open(path.Join(Storage, cpath))
	if err != nil {
		Log.Error(err)
	}
	defer f.Close()
	infos, _ := f.Readdir(-1)
	//Log.Infof("%T:%#v",infos,infos)
	for _, info := range infos {
		var m ListFiles
		//Log.Infof("%#v",info)
		if info.IsDir(){
			m.Filetype="folder.svg"
		}else {
			//Log.Infof("%T",m.Name)
			p:=strings.Split(info.Name(),".")
			m.Filetype="txt.svg"
			if len(p)==2{
				if in(p[1],Txt){
					m.Filetype="txt.svg"
				}else if in(p[1],Zip) {
					m.Filetype="zip.png"
				}
			}
		}
		m.Name = info.Name()
		m.Cpath = path.Join(cpath, "/",m.Name)
		//Log.Info(m.Name)
		m.Size = strconv.FormatInt(info.Size()/1024, 10)
		lm = append(lm, m)
	}
	//Log.Info(lm)
	c.HTML(http.StatusOK, "listfile.html", gin.H{"lm":lm,"Path":cpath})
}

func Downfile(c *gin.Context) {
	//cpath := c.DefaultQuery("path","")
	fname := c.DefaultQuery("fname","")
	//打开文件
	//Log.Info(path.Join(Storage,"/" ,fname))
	f, err := os.Open(path.Join(Storage,"/" ,fname))
	//非空处理
	if  err != nil {
		c.JSON(http.StatusOK, gin.H{
		   "success": false,
		   "message": "失败",
		   "error":   "资源不存在",
		})
		return
	}
	basename:=path.Base(f.Name())
	//Log.Info(basename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+basename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(path.Join(Storage, "/",fname))
	return
}


func Delfile(c *gin.Context) {
	cpath := c.DefaultQuery("path","")
	fname := c.DefaultQuery("fname","")
	//删除
	//Log.Info(cpath)
	//Log.Info(fname)
	//Log.Info(path.Join(Storage,"/" ,fname))
	err := os.RemoveAll(path.Join(Storage,"/" ,fname))
	if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "失败",
				"error":   "资源不存在",
			})
			return
	}
	c.Redirect(http.StatusMovedPermanently, "/?path="+cpath)
	return
}

func Renamefile(c *gin.Context) {
	cpath := c.DefaultQuery("cpath","")
	newname := c.DefaultQuery("newname","")
	//删除
	//Log.Info(path.Join(Storage,"/" ,cpath))
	err := os.Rename(path.Join(Storage,"/" ,cpath),path.Join(Storage,"/" ,newname))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "失败",
			"error":   "资源不存在",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成功",
		"error":   "",
	})
	return
}

func Mkdir(c *gin.Context) {
	cpath := c.DefaultQuery("path","")
	dirname := c.DefaultQuery("dirname","")
	//删除
	//Log.Info(cpath)
	//Log.Info(dirname)
	//Log.Info(path.Join(Storage,"/" ,cpath))
	err := os.Mkdir(path.Join(Storage,"/" ,cpath,dirname), os.ModePerm)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "失败",
			"error":   "资源不存在",
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/?path="+cpath)
	return
}

func Upfile1(c *gin.Context) {
	cpath := c.PostForm("path")
	//Log.Info(cpath)
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println(err)
		c.String(500, "上传图片出错")
	}
	//fmt.Println(file.Filename)
	// c.JSON(200, gin.H{"message": file.Header.Context})
	fname:= path.Join(Storage,"/" ,cpath,file.Filename)
	c.SaveUploadedFile(file, fname)
	//c.String(http.StatusOK, file.Filename)
	c.Redirect(http.StatusMovedPermanently, "/?path="+cpath)
	return
}

func Upfile(c *gin.Context) {
	cpath := c.PostForm("path")
	//Log.Info(cpath)
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err)
		c.String(500, "上传图片出错")
	}
	files := form.File["file"]
	for _, file := range files {
		// 逐个存
		fname:= path.Join(Storage,"/" ,cpath,file.Filename)
		log.Info(fname)
		if err := c.SaveUploadedFile(file, fname); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			return
		}
	}
	//fmt.Println(file.Filename)
	// c.JSON(200, gin.H{"message": file.Header.Context})
	//fname:= path.Join(Storage,"/" ,cpath,file.Filename)
	//c.SaveUploadedFile(file, fname)
	//c.String(http.StatusOK, file.Filename)
	c.Redirect(http.StatusMovedPermanently, "/?path="+cpath)
	return
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}
func Pardir(c *gin.Context) {
	cpath := c.DefaultQuery("path","/")
	//Log.Info(cpath)
	npath:=substr(cpath, 0, strings.LastIndex(cpath, "/"))
	//Log.Info(npath)
	c.Redirect(http.StatusMovedPermanently, "/?path="+npath)
	return
}