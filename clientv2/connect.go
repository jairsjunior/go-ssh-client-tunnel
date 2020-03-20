package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/function61/gokit/bidipipe"
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

func handleClientPipe(client net.Conn, remote net.Conn) {
	defer client.Close()

	err := bidipipe.Pipe(client, "client", remote, "remote")
	if err != nil {
		logrus.Infof(err.Error())
	}
}

func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			logrus.Tracef("error while copy remote->local: %s", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			logrus.Tracef("error while copy local->remote: %s", err)
		}
		chDone <- true
	}()

	<-chDone
}

//CreateConnectionRemoteV2 create a -R ssh connection
func CreateConnectionRemoteV2(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint) error {
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to SSH remote server using serverEndpoint
	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}
	connDialer, err := dialer.DialContext(context.Background(), "tcp", serverEndpoint.String())
	if err != nil {
		logrus.Error("Error at create dialer context")
	}
	sconn, chans, reqs, err := ssh.NewClientConn(connDialer, serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Error("Error at create new client conn")
	}
	conn := ssh.NewClient(sconn, chans, reqs)
	logrus.Info("Connection established with ssh server..")

	// Listen on remote server port
	listener, err := conn.Listen("tcp", remoteEndpoint.String())
	defer listener.Close()
	if err != nil {
		logrus.Fatalf("Listen open port ON remote server error: %s", err)
		return err
	}

	// handle incoming connections on reverse forwarded tunnel
	for {
		client, err := listener.Accept()
		if err != nil {
			logrus.Fatal(err)
			return err
		}

		local, err := conn.Dial("tcp", localEndpoint.String())
		if err != nil {
			logrus.Fatalf("Dial INTO remote service error: %s", err)
			return err
		}

		handleClientPipe(client, local)
	}
	logrus.Info("Exited for..")
	return nil
}

//CreateConnectionLocalV2 create a -L ssh connection
func CreateConnectionLocalV2(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint) error {
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Establish connection with SSH server
	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}
	connDialer, err := dialer.DialContext(context.Background(), "tcp", serverEndpoint.String())
	if err != nil {
		logrus.Error("Error at create dialer context")
	}
	sconn, chans, reqs, err := ssh.NewClientConn(connDialer, serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Error("Error at create new client conn")
	}
	conn := ssh.NewClient(sconn, chans, reqs)
	defer conn.Close()
	logrus.Info("Connection established with ssh server..")

	listener, err := net.Listen("tcp", localEndpoint.String())
	defer listener.Close()
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	for {
		client, err := listener.Accept()
		if err != nil {
			logrus.Fatal(err)
			return err
		}

		remote, err := conn.Dial("tcp", remoteEndpoint.String())
		if err != nil {
			logrus.Fatalf("Dial INTO remote service error: %s", err)
			return err
		}

		handleClientPipe(client, remote)
	}
	logrus.Info("Exited for..")
	return nil
}
