package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"runtime"

	"github.com/BurntSushi/toml"
)

var (
	deployConfig *Config
	remotePath   Path
	deployCh     chan *deploy
)

type Path struct {
	deployment string
	backup     string
	utils      string
}

type Config struct {
	Login   loginInfo
	Servers map[string]serverInfo
	Deploy  deployInfo
}

type loginInfo struct {
	User    string `toml:"username"`
	Pwd     string `toml:"password"`
	SShPath string `toml:"ssh_path"`
}

type serverInfo struct {
	Ip   []string
	Port string
}

type deployInfo struct {
	Repository    string
	GoProjectPath string `toml:"go_project_path"`
	Package       string `toml:"package_name"`
	App           string `toml:"app_name"`
	KeepRelease   string `toml:"keep_releases"`
	UseSudo       bool   `toml:"use_sudo"`
	WebUser       string `toml:"webserver_user"`
}

func ReadConfig(configfile string) (*Config, error) {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config *Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)

		return config, err
	}
	return config, nil
}

func init() {
	deployCh = make(chan *deploy)
}

func main() {
	// 모든 CPU 사용
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 현재 설정값 리턴
	fmt.Println(runtime.GOMAXPROCS(0))

	configFile := flag.String("config", "", "")
	deployAction := flag.String("action", "", "")
	serverEnv := flag.String("env", "", "")

	flag.Parse()

	if *configFile == "" || *deployAction == "" || *serverEnv == "" {
		fmt.Println("Usage:", os.Args[0], "-config=file -action=[deploy|setup] -env=[development|staging|production]")
		os.Exit(1)
	}

	var err error
	deployConfig, err = ReadConfig(*configFile)

	if err != nil {
		fmt.Println("Failed to read config")
		os.Exit(1)
	}

	remotePath.deployment = deployConfig.Deploy.GoProjectPath + "/src/" + deployConfig.Deploy.Package
	remotePath.backup = "/home/" + deployConfig.Login.User + "/backup"
	remotePath.utils = "/home/" + deployConfig.Login.User + "/utils"

	dpCh := receiver(
		deployConfig.Login.User,
		deployConfig.Login.Pwd,
		deployConfig.Servers[*serverEnv].Port,
		deployConfig.Login.SShPath,
		*deployAction,
		deployConfig.Servers[*serverEnv].Ip,
	)

	errCh := producer(dpCh)
	for err := range errCh {
		fmt.Println(err)
	}
}

func setupFnc(ch *deploy) error {
	err := ch.Setup()

	if err != nil {
		return err
	}

	return nil
}

func runFnc(ch *deploy) error {
	err := ch.Run()

	if err != nil {
		return err
	}

	return nil
}

func someOtherFunc(ch *deploy, f func(*deploy) error) error {
	return f(ch)
}

func producer(in <-chan *deploy) <-chan error {
	out := make(chan error)

	go func() {
		for ch := range in {
			switch strings.ToLower(ch.Action()) {
			case "setup":
				out <- someOtherFunc(ch, setupFnc)
			case "deploy":
				out <- someOtherFunc(ch, runFnc)
			}
		}
		close(out)
	}()
	return out
}

func receiver(User string, Pwd string, Port string, SShPath string, deployAction string, hostname []string) <-chan *deploy {
	out := make(chan *deploy)

	go func() {
		for _, ip := range hostname {
			dp, err := newDeploy(
				User,
				Pwd,
				ip,
				Port,
				SShPath,
				deployAction,
			)

			if err != nil {
				fmt.Println("Failed to start: " + err.Error())
				return
			}

			out <- dp
		}

		close(out)
	}()

	return out
}
