package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	deployConfig *Config
	remotePath   Path
	deployCh     chan *deploy
	configFile   = flag.String("config", "", "/path/to/config")
	deployAction = flag.String("action", "", "deploy|setup|deploy_file|deploy_list")
	serverEnv    = flag.String("env", "", "development|staging|production")
	deployFile   = flag.String("deploy_file", "", "/path/to/file")
	deployList   = flag.String("deploy_list", "", "/path/to/list")
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
	flag.Parse()
}

func inputServerEnv() {
	environmentMap := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	for {
		if environmentMap[*serverEnv] {
			break
		} else {
			var envString string
			for env, _ := range environmentMap {
				envString += env + "|"
			}
			fmt.Print("\nEnter server environment [" + envString + "]: ")
			fmt.Scanln(serverEnv)
		}
	}
}

func inputDeployConfig() {
	for {
		config, err := os.Open(*configFile)
		if err != nil {
			fmt.Print("\nEnter deploy config file: ")
			fmt.Scanln(configFile)
		} else {
			break
		}
		defer config.Close()
	}
}

func inputDeployAction() {
	deployActionMap := map[string]bool{
		"setup":       true,
		"deploy":      true,
		"deploy_file": true,
		"deploy_list": true,
	}

	for {
		if deployActionMap[*deployAction] {
			break
		} else {
			var deployActionString string
			for env, _ := range deployActionMap {
				deployActionString += env + "|"
			}
			fmt.Print("\nEnter deploy action [" + deployActionString + "]: ")
			fmt.Scanln(deployAction)
		}
	}
}

func inputDeployFile() {
	if *deployAction == "deploy_file" {
		for {
			dpFile, err := os.Open(*deployFile)
			if err != nil {
				fmt.Print("\nEnter deploy file: ")
				fmt.Scanln(deployFile)
			} else {
				break
			}
			defer dpFile.Close()
		}
	}
}

func inputDeployList() {
	if *deployAction == "deploy_list" {
		for {
			dpFile, err := os.Open(*deployList)
			if err != nil {
				fmt.Print("\nEnter deploy list file: ")
				fmt.Scanln(deployList)
			} else {
				break
			}
			defer dpFile.Close()
		}
	}
}

func main() {
	// 모든 CPU 사용
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 현재 설정값 리턴
	//fmt.Println(runtime.GOMAXPROCS(0))

	inputServerEnv()
	inputDeployConfig()
	inputDeployAction()
	inputDeployFile()
	inputDeployList()

	var err error
	deployConfig, err = ReadConfig(*configFile)

	if err != nil {
		fmt.Println("Failed to read config")
		os.Exit(1)
	}

	remotePath.deployment = deployConfig.Deploy.GoProjectPath + "/src/" + deployConfig.Deploy.Package
	remotePath.backup = "/home/" + deployConfig.Login.User + "/backup"
	remotePath.utils = "/home/" + deployConfig.Login.User + "/utils"

	// spawn worker goroutines
	// http://stackoverflow.com/questions/18405023/how-would-you-define-a-pool-of-goroutines-to-be-executed-at-once-in-golang
	var wg sync.WaitGroup
	for range deployConfig.Servers[*serverEnv].Ip {
		// Add adds delta, which may be negative, to the WaitGroup counter.
		// If the counter becomes zero, all goroutines blocked on Wait are released.
		// If the counter goes negative, Add panics.
		//
		// the calls to Add should execute before the statement creating the goroutine or other event to be waited for.
		// If a WaitGroup is reused to wait for several independent sets of events,
		// new Add calls must happen after all previous Wait calls have returned.
		wg.Add(1)
		go func() {
			for dp := range deployCh {
				switch strings.ToLower(dp.Action()) {
				case "setup":
					err = dp.Setup()
				case "deploy":
					err = dp.Deploy()
					if err == nil {
						err = dp.Run()
					}
				case "deploy_file":
					if strings.Index(*deployFile, deployConfig.Deploy.Package) == -1 {
						err = errors.New("invalid path")
					} else {
						// []string
						path := strings.SplitAfter(*deployFile, deployConfig.Deploy.Package)
						remoteFile := deployConfig.Deploy.GoProjectPath + "/src/" + deployConfig.Deploy.Package + path[1]

						err = dp.Transafer(*deployFile, remoteFile)
						if err == nil {
							err = dp.Run()
						}
					}
				case "deploy_list":
					localFile, err := os.Open(*deployList)
					if err != nil {
						err = fmt.Errorf("localFile read: %v", err)
					}
					defer localFile.Close()

					fileInfo, err := localFile.Stat()
					if err != nil {
						err = fmt.Errorf("localFile stat: %v", err)
					}
					localSize := fileInfo.Size()

					if localSize > 0 {
						var data = make([]byte, localSize)
						_, err := localFile.Read(data)
						if err != nil {
							err = fmt.Errorf("localFile read: %v", err)
						}

						splitData := strings.Split(string(data), "\n")
						for _, localFile := range splitData {
							if localFile != "" {
								// []string
								path := strings.SplitAfter(localFile, deployConfig.Deploy.Package)
								remoteFile := deployConfig.Deploy.GoProjectPath + "/src/" + deployConfig.Deploy.Package + path[1]
								err = dp.Transafer(localFile, remoteFile)
								if err != nil {
									break
								}
							}
						}

						if err == nil {
							err = dp.Run()
						}
					} else {
						err = errors.New("localFile size is 0")
					}

				default:
					err = errors.New("Invalid command!")
				}

				if err != nil {
					fmt.Println(err.Error())
				}

				defer dp.cl.Close()
			}
			// Done decrements the WaitGroup counter.
			wg.Done()
		}()
	}

	// generate some tasks
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

	// Wait blocks until the WaitGroup counter is zero.
	// wait for the workers to finish
	wg.Wait()
}
