package main

import (
	"bytes"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHWorker - Initialize thread for Windows Management Interface
func SSHWorker(server string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	timeStart := time.Now()
	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Thread started")
	if !validateServer(server) {
		ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Thread finished")
		return
	}
	ApplicationLogger.Output(3, "[INFO] Secure Shell ("+server+"): Validation passed, establishing connection")
	if !runOnServer(server) {
		ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): Connection or execution failed, see errors above this line.")
		ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Thread finished in "+time.Since(timeStart).String())
		return
	}
	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Thread finished in "+time.Since(timeStart).String())
}

func validateServer(server string) bool {
	if ApplicationConfig.Server[server].DestHost == "" {
		ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): DestHost cannot be empty")
		return false
	} else if ApplicationConfig.Server[server].DestPort == 0 {
		ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): DestPort cannot be null")
		return false
	} else if ApplicationConfig.Server[server].HostUser == "" {
		ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): HostUser cannot be empty")
		return false
	}
	if ApplicationConfig.Server[server].ExecCommands == nil {
		ApplicationLogger.Output(3, "[WARNING] Secure Shell ("+server+"): ExecCommands array empty")
	}
	if ApplicationConfig.Server[server].HostPass == "" && ApplicationConfig.Server[server].SSHPrivateKey == "" {
		ApplicationLogger.Output(3, "[WARNING] Secure Shell ("+server+"): HostPass AND SSHPrivateKey are empty")
	}
	return true
}

func runOnServer(server string) bool {
	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Entering in connection state")
	var sshClientConfig *ssh.ClientConfig
	if ApplicationConfig.Server[server].SSHPrivateKey == "" {
		ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): SSHPrivate key not defined, using password authentication.")
		sshClientConfig = &ssh.ClientConfig{
			User: ApplicationConfig.Server[server].HostUser,
			Auth: []ssh.AuthMethod{
				ssh.Password(ApplicationConfig.Server[server].HostPass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): SSHPrivate key defined, trying it first.")
		sshClientPrivateKey := []byte(ApplicationConfig.Server[server].SSHPrivateKey)
		signer, err := ssh.ParsePrivateKey(sshClientPrivateKey)
		if err != nil {
			ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): Unable to parse private key, error: +"+err.Error()+"")
			return false
		}
		ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Private key parsed successfuly.")
		sshClientConfig = &ssh.ClientConfig{
			User: ApplicationConfig.Server[server].HostUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Attempting to establish connection to server")
	sshClient, err := ssh.Dial("tcp", ApplicationConfig.Server[server].DestHost+":"+strconv.Itoa(ApplicationConfig.Server[server].DestPort), sshClientConfig)
	if err != nil {
		ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"): Unable to establish connection, error: "+err.Error())
		return false
	}
	defer sshClient.Close()
	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Connection established successfuly")

	var waitGroup sync.WaitGroup
	for index, execCommand := range ApplicationConfig.Server[server].ExecCommands {
		waitGroup.Add(1)
		go func(index int, execCommand string) {
			defer waitGroup.Done()
			ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"."+strconv.Itoa(index)+"): Executing command "+execCommand)
			sshSession, err := sshClient.NewSession()
			if err != nil {
				ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"."+strconv.Itoa(index)+"): Unable to open session: "+err.Error())
				return
			}
			defer sshSession.Close()
			ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"."+strconv.Itoa(index)+"): Session opened successfuly")

			var (
				stdoutBuffer bytes.Buffer
				stderrBuffer bytes.Buffer
			)

			sshSession.Stdout = &stdoutBuffer
			sshSession.Stderr = &stderrBuffer

			err = sshSession.Run(execCommand)
			if err != nil {
				ApplicationLogger.Output(3, "[ERROR] Secure Shell ("+server+"."+strconv.Itoa(index)+"): Command execution failed with error: "+err.Error())
			}
			//ApplicationLogger.Output(3, "[TRACE] Secure Shell ("+server+"."+strconv.Itoa(index)+"): Execution results\nCommand stdout:\n"+stdoutBuffer.String()+"Command stderr:\n"+stderrBuffer.String())

			ExecResultSave(server, execCommand, stdoutBuffer.String(), stderrBuffer.String())
		}(index, execCommand)
	}
	waitGroup.Wait()

	ApplicationLogger.Output(3, "[DEBUG] Secure Shell ("+server+"): Execution complete.")
	return true
}
