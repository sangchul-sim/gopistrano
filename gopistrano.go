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
)

type Path struct {
	deployment string
	shared     string
	utils      string
	release    string
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
	Port int
}

type deployInfo struct {
	Repository    string
	GoProjectPath string `toml:"go_project_path"`
	Package       string `toml:"package_name"`
	App           string `toml:"app_name"`
	KeepRelease   int    `toml:"keep_releases"`
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
	remotePath.release = remotePath.deployment + "/releases"
	remotePath.shared = remotePath.deployment + "/shared"
	remotePath.utils = remotePath.deployment + "/utils"

	for _, ip := range deployConfig.Servers[*serverEnv].Ip {
		deploy, err := newDeploy(ip, deployConfig.Servers[*serverEnv].Port, deployConfig.Login.SShPath)
		if err != nil {
			fmt.Println("Failed to start: " + err.Error())
			return
		}

		switch strings.ToLower(*deployAction) {
		case "setup":
			err = deploy.Setup()
		case "deploy":
			err = deploy.Run()
		default:
			fmt.Println("Invalid command!")
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
