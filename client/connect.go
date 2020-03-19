package client

import (
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

//Endpoint .
type Endpoint struct {
	Host string
	Port int
}

//String exports string of endpoint
func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			logrus.Errorf("error while copy remote->local: %s", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			logrus.Errorf("error while copy remote->local: %s", err)
		}
		chDone <- true
	}()

	<-chDone
}

//CreateConnectionRemote create a -R ssh connection
func CreateConnectionRemote(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint) error {
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH remote server using serverEndpoint
	serverConn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Errorf("Dial INTO remote server error: %s", err)
		return err
	}

	// Listen on remote server port
	listener, err := serverConn.Listen("tcp", remoteEndpoint.String())
	if err != nil {
		logrus.Errorf("Listen open port ON remote server error: %s", err)
		return err
	}
	defer listener.Close()

	// handle incoming connections on reverse forwarded tunnel
	for {
		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		local, err := net.Dial("tcp", localEndpoint.String())
		if err != nil {
			logrus.Errorf("Dial INTO local service error: %s", err)
			return err
		}

		client, err := listener.Accept()
		if err != nil {
			logrus.Error(err)
			return err
		}

		handleClient(client, local)
	}
	return nil
}

//CreateConnectionLocal create a -L ssh connection
func CreateConnectionLocal(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint) error {
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer conn.Close()

	listener, err := net.Listen("tcp", localEndpoint.String())
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer listener.Close()

	for {
		remote, err := conn.Dial("tcp", remoteEndpoint.String())
		if err != nil {
			logrus.Errorf("Dial INTO remote service error: %s", err)
			return err
		}

		client, err := listener.Accept()
		if err != nil {
			logrus.Error(err)
			return err
		}

		handleClient(client, remote)
	}
	return nil
}
