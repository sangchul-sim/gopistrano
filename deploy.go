package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

type deploy struct {
	cl *ssh.Client
}

//returns a new deployment
func newDeploy(User string, Pwd string, Hostname string, Port string, SshPath string) (d *deploy, err error) {
	if deployConfig.Login.Pwd != "" {
		cfg := &ssh.ClientConfig{
			User: User,
			Auth: []ssh.AuthMethod{
				ssh.Password(Pwd),
			},
		}
		fmt.Println("SSH-ing into " + Hostname + ":" + Port)
		cl, err := ssh.Dial("tcp", Hostname+":"+Port, cfg)
		if err != nil {
			return nil, err
		}
		d = &deploy{cl: cl}
	}
	if SshPath != "" {
		sshConfig := &ssh.ClientConfig{
			User: User,
			Auth: []ssh.AuthMethod{
				PublicKeyFile(SshPath),
			},
		}
		fmt.Println("SSH-ing into " + Hostname + ":" + Port)
		cl, err := ssh.Dial("tcp", Hostname+":"+Port, sshConfig)
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
	deployCmd := "if [ ! -d " + deployConfig.Deploy.GoProjectPath + " ]; then mkdir " + deployConfig.Deploy.GoProjectPath + "; fi && " +
		"if [ ! -d " + remotePath.utils + " ]; then exit 1; fi && " +
		"if [ ! -f " + remotePath.utils + "/deploy.pl ]; then exit 1; fi && " +
		remotePath.utils + "/deploy.pl " + deployConfig.Login.User + " " + deployConfig.Deploy.GoProjectPath + " " +
		deployConfig.Deploy.Package + " " + deployConfig.Deploy.Repository + " " + deployConfig.Deploy.KeepRelease

	if err := d.runCmd(deployCmd); err != nil {
		return err
	}

	//fmt.Println("deployCmd", deployCmd)
	fmt.Println("Project Deployed!")
	fmt.Println("Restarting Tmux at " + remotePath.deployment)

	restartCmdD := "if [ ! -d " + remotePath.utils + " ]; then exit 1; fi && " +
		"if [ ! -f " + remotePath.utils + "/run_app.pl ]; then exit 1; fi && " +
		remotePath.utils + "/run_app.pl " + deployConfig.Deploy.GoProjectPath + " " +
		deployConfig.Deploy.Package + " " + deployConfig.Deploy.App

	//fmt.Println("restartCMD", restartCmdD)
	if err := d.runCmd(restartCmdD); err != nil {
		return err
	}

	fmt.Println("Tmux Restarted!")
	return nil
}

// sets up directories for deployment a la capistrano
func (d *deploy) Setup() error {
	cdPathCmd := "if [ ! -d " + remotePath.backup + " ]; then mkdir " + remotePath.backup + "; fi &&" +
		"if [ ! -d " + remotePath.utils + " ]; then mkdir " + remotePath.utils + "; fi"

	if err := d.runCmd(cdPathCmd); err != nil {
		return err
	}

	fmt.Println("running scp connection")

	cpy := `echo -n '` + string(deploymentScript) + `' > ` + remotePath.utils + `/deploy.pl ; chmod +x ` + remotePath.utils + `/deploy.pl`
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
