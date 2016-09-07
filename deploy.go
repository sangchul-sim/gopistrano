package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
	"strconv"
)

type deploy struct {
	cl *ssh.Client
}

//returns a new deployment
func newDeploy(Hostname string, Port int, SshPath string) (d *deploy, err error) {
	fmt.Println("Port", Port)
	if deployConfig.Login.Pwd != "" {
		cfg := &ssh.ClientConfig{
			User: deployConfig.Login.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(deployConfig.Login.Pwd),
			},
		}
		fmt.Println("SSH-ing into " + Hostname+":"+strconv.Itoa(Port))
		cl, err := ssh.Dial("tcp", Hostname+":"+strconv.Itoa(Port), cfg)
		if err != nil {
			return nil, err
		}
		d = &deploy{cl: cl}
	}
	if SshPath != "" {
		sshConfig := &ssh.ClientConfig{
			User: deployConfig.Login.User,
			Auth: []ssh.AuthMethod{
				PublicKeyFile(SshPath),
			},
		}
		fmt.Println("SSH-ing into " + Hostname+":"+strconv.Itoa(Port))
		cl, err := ssh.Dial("tcp", Hostname+":"+strconv.Itoa(Port), sshConfig)
		if err != nil {
			return nil, err
		}
		d = &deploy{cl: cl}
	}

	return
}

//
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// runs the deployment script remotely
func (d *deploy) Run() error {
	deployCmd := "if [ ! -d " + remotePath.release + " ]; then exit 1; fi &&" +
		"if [ ! -d " + remotePath.shared + " ]; then exit 1; fi &&" +
		"if [ ! -d " + remotePath.utils + " ]; then exit 1; fi &&" +
		"if [ ! -f " + remotePath.utils + "/deploy.sh ]; then exit 1; fi &&" +
		"" + remotePath.utils + "/deploy.sh " + remotePath.deployment + " " + deployConfig.Deploy.Repository + " " + string(deployConfig.Deploy.KeepRelease) + " " + deployConfig.Deploy.App

	if err := d.runCmd(deployCmd); err != nil {
		return err
	}

	fmt.Println("Project Deployed!")
	fmt.Println("Restarting Tmux at " + remotePath.deployment)

	restartCmdD := "if [ ! -d " + remotePath.release + " ]; then exit 1; fi &&" +
		"if [ ! -d " + remotePath.shared + " ]; then exit 1; fi &&" +
		"if [ ! -d " + remotePath.utils + " ]; then exit 1; fi &&" +
		"if [ ! -f " + remotePath.utils + "/run_app.pl ]; then exit 1; fi &&" +
		" " + remotePath.utils + "/run_app.pl " + deployConfig.Deploy.GoProjectPath + " " + deployConfig.Deploy.Package + " " + deployConfig.Deploy.App

	fmt.Println("restartCMD", restartCmdD)
	if err := d.runCmd(restartCmdD); err != nil {
		return err
	}

	fmt.Println("Tmux Restarted!")
	return nil
}

// sets up directories for deployment a la capistrano
func (d *deploy) Setup() error {
	cdPathCmd := "if [ ! -d " + remotePath.release + " ]; then mkdir -p " + remotePath.release + "; fi &&" +
		"if [ ! -d " + remotePath.shared + " ]; then mkdir " + remotePath.shared + "; fi &&" +
		"if [ ! -d " + remotePath.utils + " ]; then mkdir " + remotePath.utils + "; fi &&" +
		"chmod g+w " + remotePath.release + " " + remotePath.shared + " " + remotePath.deployment + " " + remotePath.utils

	if err := d.runCmd(cdPathCmd); err != nil {
		return err
	}

	fmt.Println("running scp connection")

	cpy := `echo -n '` + string(deploymentScript) + `' > ` + remotePath.utils + `/deploy.sh ; chmod +x ` + remotePath.utils + `/deploy.sh`
	cpi := `echo -n '` + string(runScript) + `' > ` + remotePath.utils + `/run_app.pl ; chmod +x ` + remotePath.utils + `/run_app.pl`

	if err := d.runCmd(cpy); err != nil {
		return err
	}
	if err := d.runCmd(cpi); err != nil {
		return err
	}

	fmt.Println("Cool Beans! Gopistrano created the structure correctly!")
	return nil
}

// basic ssh cmd runner
func (d *deploy) runCmd(cmd string) (err error) {
	session, err := d.cl.NewSession()
	if err != nil {
		return err
	}

	//this *does* return an error (EOF of some sort), but I guess we don't care?
	//the ssh lib needs to send it and must return it or something
	defer session.Close()

	//send through to main stdout, stderr
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(cmd)
}
