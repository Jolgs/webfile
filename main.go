package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

var Port = ":80"      //端口
var Storage = "./data" //上传文件目录


var CurrentDirectory string

func main() {
	ip, err := ExternalIP()
	if err != nil {
		fmt.Println(err)
	}
	//h := http.FileServer(http.Dir("."))
	fmt.Printf(`
	_______                   __    ___________                  .__              .__   
 \      \   ____ ___  ____/  |_  \__    ___/__________  _____ |__| ____ _____  |  |  
 /   |   \_/ __ \\  \/  /\   __\   |    |_/ __ \_  __ \/     \|  |/    \\__  \ |  |  
/    |    \  ___/ >    <  |  |     |    |\  ___/|  | \/  Y Y  \  |   |  \/ __ \|  |__
\____|__  /\___  >__/\_ \ |__|     |____| \___  >__|  |__|_|  /__|___|  (____  /____/
        \/     \/      \/                     \/            \/        \/     \/      


		======================================WebFile==========================

`)
	fmt.Printf("请在浏览器访问：http://%s%s\n",ip.String(),Port)
	checkdir(Storage)
	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.ReleaseMode)
	//r := gin.Default()
	r := gin.New()
	r.Use(Cors())
	r.Use(Ginlog())
	r.Static("/static", "./static")
	r.Static("/files", "./files")
	r.LoadHTMLGlob("temple/*")
	r.GET("/", Index)
	r.GET("/downfile", Downfile)
	r.GET("/delfile", Delfile)
	r.GET("/renamefile", Renamefile)
	r.GET("/newdir", Mkdir)
	r.GET("/pardir", Pardir)
	r.POST("/upfile", Upfile)
	r.Run(Port)

}


