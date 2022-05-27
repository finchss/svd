package main

import (
	"flag"
	"fmt"
	"github.com/google/shlex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)
var wg sync.WaitGroup


func loadConfig() []string {
	configFileName:="svd.conf"
	var location=make([]string,0)

	cwd, err := os.Getwd()
	if err==nil {
		location = append(location,cwd)
	}
	location = append(location,".",filepath.Dir(os.Args[0]),"/etc")

	for _,l :=range location{
		if l == "/" {
			l = ""
		}
		tloc:= l+"/"+configFileName
		fmt.Println("Trying to load config from "+tloc)
		f,err:=os.ReadFile(tloc)
		if err == nil {
			fd := string(f)
			y:=strings.Split(fd,"\n")
			fmt.Println("Loaded ", len(y)," lines")
			return y
		}
	}
	os.Exit(0x01)
	return nil
}

func doStuff(l string) {
	defer wg.Done()
	a,err := shlex.Split(l)
	if (len(a)<1) {
		return
	}


	if err != nil {
		fmt.Println(err.Error())
		return ;
	}
	for true {
		fmt.Println("Exec ",l)
		cmd:=exec.Command(a[0],a[1:]...)
		err:=cmd.Run()
		if err!=nil {
			fmt.Println(l,"-->",err.Error())
		}
		time.Sleep(1*time.Second)
	}
}
type  pcT struct {
	flagInstall bool
	installUser string
	installGroup string
}

var pc pcT

func installService(){
	var cmd *exec.Cmd
	var err error
	err = os.WriteFile("/etc/systemd/system/svd.service",[]byte("" +
		"[Unit]\n" +
		"Description=Supervisor demon\n\n" +
		"[Service]\n\n" +
		"Restart=always\n" +
		"RestartSec=5s\n" +
		"User="+pc.installUser+"\n" +
		"Group="+pc.installGroup+"\n" +
		"ExecStart=/bin/svd\n" +
		"[Install]\n" +
		"WantedBy=multi-user.target" +
		"\n"),0644)
	if err != nil {
		return
	}

	fmt.Println("Copy ",os.Args[0],"to /bin/svd")
	d, err :=os.ReadFile(os.Args[0])
	if err !=nil{
		return
	}
	err = os.WriteFile("/bin/svd",d,0755)
	if err != nil {
		return
	}


	fmt.Println("Exec \"/usr/bin/systemctl daemon-reload\" ")
	cmd=exec.Command("/usr/bin/systemctl","daemon-reload")
	err=cmd.Run()
	if err!=nil{
		return
	}

	fmt.Println("Exec \"/usr/bin/systemctl enable svd\" ")
	cmd=exec.Command("/usr/bin/systemctl","enable","svd")
	err=cmd.Run()
	if err!=nil{
		return
	}

	fmt.Println("Exec \"/usr/bin/systemctl start svd\" ")
	cmd=exec.Command("/usr/bin/systemctl","start","svd")
	err=cmd.Run()
	if err!=nil{
		return
	}



}
func main()  {



	flag.BoolVar(&pc.flagInstall,"install",false,"Install systemd service")
	flag.StringVar(&pc.installUser,"iuser" ,"steve","Install systemd service - run as user")
	flag.StringVar(&pc.installGroup,"igroup","steve","Install systemd service - run as group")
	flag.Parse()

	if pc.flagInstall {
		installService()
		os.Exit(0)
	}


	lines:=loadConfig()
	for _,l:=range lines {
		wg.Add(1)
		time.Sleep(1*time.Second)
		go doStuff(l)
	}
	wg.Wait()
}
