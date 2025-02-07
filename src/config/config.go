package cf

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"net/http"
	"crypto/tls"

	ini "github.com/BurntSushi/toml"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	CENTER_PATH         string
	SERVER_PATH         string
	CENTER_LOG_PATH		string
	SERVER_LOG_PATH		string
	CENTER_ADDR			string
	TIME_FLAG			string

	PARTNER_KEY			string
	PROFILE_KEY			string
	API_SERVER			string
}

var Conf Config
var Stdlog *log.Logger
var BasePath string
var Client *resty.Client

func InitConfig() {
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	logDir := filepath.Join(dir, "logs")
	err := createDir(logDir)
	if err != nil {
		log.Fatalf("Failed to ensure log directory: %s", err)
	}
	path := filepath.Join(logDir, "DHNHealthchecker")
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
		rotatelogs.WithLocation(loc),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(7),
	)

	if err != nil {
		log.Fatalf("Failed to Initialize Log File %s", err)
	}

	log.SetOutput(writer)
	stdlog := log.New(os.Stdout, "INFO -> ", log.Ldate|log.Ltime)
	stdlog.SetOutput(writer)
	Stdlog = stdlog

	Conf = readConfig()
	BasePath = dir + "/"

	Client = resty.New().
		SetTimeout(100 * time.Second).
		SetTLSClientConfig(&tls.Config{MinVersion: tls.VersionTLS12}).
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetTransport(&http.Transport{
			IdleConnTimeout:	 90 * time.Second,
		})

}

func readConfig() Config {
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	var configfile = filepath.Join(dir, "config.ini")
	_, err := os.Stat(configfile)
	if err != nil {

		if err != nil {
			Stdlog.Println("Config file create fail")
		}
		Stdlog.Println("config.ini 생성완료 작성을 해주세요.")

		system_exit("DHNHealthchecker")
		fmt.Println("Config file is missing")
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}

func createDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func system_exit(service_name string) {
	cmd := exec.Command("systemctl", "stop", service_name)
	if err := cmd.Run(); err != nil {
		Stdlog.Println(service_name+" 서비스 종료 실패:", err)
	} else {
		Stdlog.Println(service_name + " 서비스가 성공적으로 종료되었습니다.")
	}
}
