package main

import(
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"strings"

	cf "hc/src/config"
	send "hc/src/send"

	"github.com/takama/daemon"
)

const (
	name        = "DHNHealthcheck"
	description = "대형네트웍스 센터, 서버 헬스체커"
	certEmail   = "dhn@dhncorp.co.kr"
)

var dependencies = []string{name+".service"}

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: "+name+" install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}
	resultProc()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case killSignal := <-interrupt:
			cf.Stdlog.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

func main(){
	cf.InitConfig()

	var rLimit syscall.Rlimit

	rLimit.Max = 50000
	rLimit.Cur = 50000

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)

	if err != nil {
		cf.Stdlog.Println("Error Setting Rlimit ", err)
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)

	if err != nil {
		cf.Stdlog.Println("Error Getting Rlimit ", err)
	}

	cf.Stdlog.Println("Rlimit Final", rLimit)

	srv, err := daemon.New(name, description, daemon.SystemDaemon, dependencies...)
	if err != nil {
		cf.Stdlog.Println("Error: ", err)
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		cf.Stdlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}

func resultProc(){
	go logChecker(cf.Conf.CENTER_LOG_PATH, cf.Conf.TIME_FLAG)
	go logChecker(cf.Conf.SERVER_LOG_PATH, cf.Conf.TIME_FLAG)
}

func logChecker(logPath, timeFlag string) {
	preCenterLogString := ""
	curCenterLogString := ""

	for {
		preCenterLogString = curCenterLogString

		now := time.Now()                  // 현재 시간 가져오기
		logDate := now.Format("2006-01-02") // 기본 날짜 (오늘 날짜)

		if (now.Hour() < 8 || now.Hour() > 22) && strings.EqualFold(timeFlag, "Y") {
			time.Sleep(1 * time.Minute)
			continue
		}

		if now.Hour() < 9 {
			yesterday := now.AddDate(0, 0, -1) // 하루 전 날짜
			logDate = yesterday.Format("2006-01-02")
		}

		logFileName := fmt.Sprintf("%s%s.log", logPath, logDate)

		cmd := exec.Command("tail", "-n", "1", logFileName) // Windows: exec.Command("cmd", "/C", "dir")

		// 명령어 실행 및 결과 가져오기
		output, err := cmd.Output()
		if err != nil {
			cf.Stdlog.Println("Error:", err)
		}

		curCenterLogString = string(output)

		if strings.EqualFold(preCenterLogString, curCenterLogString){
			send.SendAlimtalk("821093440043")
			send.SendAlimtalk("821055537431")
			send.SendAlimtalk("821077709980")
		}

		// 응답 출력
		time.Sleep(5 * time.Minute)
	}
}