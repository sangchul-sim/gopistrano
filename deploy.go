package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"io"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type deploy struct {
	cl     *ssh.Client
	action string
}

//returns a new deployment
func NewDeploy(User string, Pwd string, Hostname string, Port string, SshPath string, DeployAction string) (d *deploy, err error) {
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
		d = &deploy{
			cl:     cl,
			action: DeployAction,
		}
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
		d = &deploy{
			cl:     cl,
			action: DeployAction,
		}
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

func (d *deploy) Backup() error {
	backupCmd := "if [ ! -d " + remotePath.utils + " ]; then exit 1; fi && " +
		"if [ ! -f " + remotePath.utils + "/backup.pl ]; then exit 1; fi && " +
		remotePath.utils + "/backup.pl " + deployConfig.Login.User + " " + deployConfig.Deploy.KeepRelease

	if err := d.runCmd(backupCmd); err != nil {
		return err
	}

	fmt.Println("backup Completed!")
	return nil
}

// runs the deployment script
func (d *deploy) Deploy() error {
	deployCmd := "if [ ! -d " + deployConfig.Deploy.GoProjectPath + " ]; then mkdir " + deployConfig.Deploy.GoProjectPath + "; fi && " +
		"if [ ! -d " + remotePath.utils + " ]; then exit 1; fi && " +
		"if [ ! -f " + remotePath.utils + "/deploy.pl ]; then exit 1; fi && " +
		remotePath.utils + "/deploy.pl " + deployConfig.Login.User + " " + deployConfig.Deploy.GoProjectPath + " " +
		//deployConfig.Deploy.Package + " " + deployConfig.Deploy.Repository + " " + deployConfig.Deploy.KeepRelease
		deployConfig.Deploy.Package + " " + deployConfig.Deploy.Repository

	if err := d.runCmd(deployCmd); err != nil {
		return err
	}

	fmt.Println("deploy Completed!")
	return nil
}

// runs the restart script remotely
func (d *deploy) Restart() error {
	restartCmd := "if [ ! -d " + remotePath.utils + " ]; then exit 1; fi && " +
		"if [ ! -f " + remotePath.utils + "/run_app.pl ]; then exit 1; fi && " +
		remotePath.utils + "/run_app.pl " + deployConfig.Deploy.GoProjectPath + " " +
		deployConfig.Deploy.Package + " " + deployConfig.Deploy.App

	if err := d.runCmd(restartCmd); err != nil {
		return err
	}

	fmt.Println("Tmux Restarted!")
	return nil
}

func (d *deploy) Transafer(localPath string, remotePath string) (err error) {
	c, err := sftp.NewClient(d.cl, sftp.MaxPacket(1<<15))
	if err != nil {
		return fmt.Errorf("sftp.NewClient: %v", err)
	}
	defer c.Close()

	remoteFile, err := c.Create(remotePath)
	if err != nil {
		return fmt.Errorf("remoteFile create: %v", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("localFile read: %v", err)
	}
	defer localFile.Close()

	//const size int64 = 1e9
	fileInfo, err := localFile.Stat()
	if err != nil {
		return err
	}
	localSize := fileInfo.Size()

	timeStamp := time.Now()
	remoteSize, err := io.Copy(remoteFile, io.LimitReader(localFile, localSize))
	if err != nil {
		return fmt.Errorf("copy: %v", err)
	}

	if remoteSize != localSize {
		return fmt.Errorf("copy: expected %v bytes, got %d", localSize, remoteSize)
	}
	fmt.Printf("wrote %s %v bytes in %s\n", remotePath, localSize, time.Since(timeStamp))
	return err
}

// sets up directories for deployment a la capistrano
// copy deploy.pl and run_app.pl to utils directory
func (d *deploy) Setup() error {
	cdPathCmd := "if [ ! -d " + remotePath.backup + " ]; then mkdir " + remotePath.backup + "; fi &&" +
		"if [ ! -d " + remotePath.utils + " ]; then mkdir " + remotePath.utils + "; fi"

	if err := d.runCmd(cdPathCmd); err != nil {
		return err
	}

	fmt.Println("running scp connection")

	copyDeploymentScript := `echo -n '` + string(deploymentScript) + `' > ` + remotePath.utils + `/deploy.pl ; chmod +x ` + remotePath.utils + `/deploy.pl`
	copyBackupScript := `echo -n '` + string(backupScript) + `' > ` + remotePath.utils + `/backup.pl ; chmod +x ` + remotePath.utils + `/backup.pl`
	copyRestartScript := `echo -n '` + string(restartScript) + `' > ` + remotePath.utils + `/run_app.pl ; chmod +x ` + remotePath.utils + `/run_app.pl`

	if err := d.runCmd(copyDeploymentScript); err != nil {
		return err
	}

	if err := d.runCmd(copyBackupScript); err != nil {
		return err
	}

	if err := d.runCmd(copyRestartScript); err != nil {
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

func (d *deploy) Action() string {
	return d.action
}
