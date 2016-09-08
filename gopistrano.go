package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
	deployCh = make(chan *deploy, 100)

	//go func() {
	//	var err error
	//
	//	for {
	//		dp := <-deployCh
	//		switch strings.ToLower(dp.Action()) {
	//		case "setup":
	//			err = dp.Setup()
	//		case "deploy":
	//			err = dp.Run()
	//		default:
	//			fmt.Println("Invalid command!")
	//		}
	//
	//		if err != nil {
	//			fmt.Println(err.Error())
	//		}
	//	}
	//}()
}

func main() {
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

	go func() {
		for _, ip := range deployConfig.Servers[*serverEnv].Ip {
			deploy, err := newDeploy(
				deployConfig.Login.User,
				deployConfig.Login.Pwd,
				ip,
				deployConfig.Servers[*serverEnv].Port,
				deployConfig.Login.SShPath,
				*deployAction,
			)

			if err != nil {
				fmt.Println("Failed to start: " + err.Error())
				return
			}

			deployCh <- deploy
		}
		close(deployCh)
	}()

	for ch := range deployCh {
		switch strings.ToLower(ch.Action()) {
		case "setup":
			err = ch.Setup()
		case "deploy":
			err = ch.Run()
		default:
			fmt.Println("Invalid command!")
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
