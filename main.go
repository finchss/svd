package main

import (
	"flag"
	"github.com/google/shlex"
	"github.com/k0kubun/pp"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func readConfig(file string) ([]string, error) {
	log.Println("trying to load config from", file)
	f, err := os.ReadFile(file)
	if err == nil {
		fd := string(f)
		y := strings.Split(fd, "\n")
		log.Println("Loaded", len(y), "lines from", file)
		return y, nil
	} else {
		return nil, err
	}
}
func loadConfig() []string {

	configFileName := "svd.conf"
	var location = make([]string, 0)

	cwd, err := os.Getwd()
	if err == nil {
		location = append(location, cwd)
	}
	location = append(location, "/etc")

	for _, l := range location {
		if l == "/" {
			l = ""
		}
		tloc := l + "/" + configFileName
		if y, err := readConfig(tloc); err == nil {
			return y
		}
	}
	os.Exit(0x01)
	return nil
}

func doStuff(l string) {
	defer wg.Done()

	a, err := shlex.Split(l)
	if len(a) < 1 {
		return
	}

	if err != nil {
		log.Println(err.Error())
		return
	}

	if l[0] == '#' || a[0][0] == '#' || a[0] == "#" {
		return
	}

	for true {
		log.Println("Exec ", l)
		cmd := exec.Command(a[0], a[1:]...)
		err := cmd.Run()
		if err != nil {
			log.Println(l, "-->", err.Error())
		}
		time.Sleep(1 * time.Second)
	}
}

type pcT struct {
	flagInstall   bool
	installUser   string
	installGroup  string
	configFile    string
	showDebugInfo bool
}

var pc pcT

func installService() {
	var cmd *exec.Cmd
	var err error
	err = os.WriteFile("/etc/systemd/system/svd.service", []byte(""+
		"[Unit]\n"+
		"Description=Supervisor demon\n\n"+
		"[Service]\n\n"+
		"Restart=always\n"+
		"RestartSec=5s\n"+
		"User="+pc.installUser+"\n"+
		"Group="+pc.installGroup+"\n"+
		"ExecStart=/bin/svd\n"+
		"[Install]\n"+
		"WantedBy=multi-user.target"+
		"\n"), 0644)
	if err != nil {
		return
	}

	log.Println("Copy ", os.Args[0], "to /bin/svd")
	d, err := os.ReadFile(os.Args[0])
	if err != nil {
		return
	}
	err = os.WriteFile("/bin/svd", d, 0755)
	if err != nil {
		return
	}

	log.Println("Exec \"/usr/bin/systemctl daemon-reload\" ")
	cmd = exec.Command("/usr/bin/systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return
	}

	log.Println("Exec \"/usr/bin/systemctl enable svd\" ")
	cmd = exec.Command("/usr/bin/systemctl", "enable", "svd")
	err = cmd.Run()
	if err != nil {
		return
	}

	log.Println("Exec \"/usr/bin/systemctl start svd\" ")
	cmd = exec.Command("/usr/bin/systemctl", "start", "svd")
	err = cmd.Run()
	if err != nil {
		return
	}

}

func showDebugInfo() {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		_, _ = pp.Println(bi)
	}
	os.Exit(0)
}
func main() {

	log.SetFlags(log.LstdFlags)
	var lines []string
	var err error
	flag.BoolVar(&pc.flagInstall, "install", false, "Install systemd service")
	flag.StringVar(&pc.installUser, "iuser", "steve", "Install systemd service - run as user")
	flag.StringVar(&pc.installGroup, "igroup", "steve", "Install systemd service - run as group")
	flag.StringVar(&pc.configFile, "f", "", "config file")
	flag.BoolVar(&pc.showDebugInfo, "info", false, "show debug info")
	flag.Parse()

	if pc.showDebugInfo {
		showDebugInfo()
	}

	if pc.flagInstall {
		installService()
		os.Exit(0)
	}

	if pc.configFile == "" {
		lines = loadConfig()
	} else {
		lines, err = readConfig(pc.configFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, l := range lines {
		wg.Add(1)
		time.Sleep(1 * time.Second)
		go doStuff(l)
	}
	wg.Wait()
}
