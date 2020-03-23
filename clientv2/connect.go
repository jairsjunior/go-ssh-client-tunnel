package client

import (
	"context"
	"fmt"
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

func handleClientPipe(client net.Conn, remote net.Conn, isConnected chan bool) error {
	defer closeClient(client)

	logrus.Info(">>>>>>>>>>>>>>>>>>>>>> before piping")
	err := bidipipe.Pipe(client, "client", remote, "remote")
	logrus.Info(">>>>>>>>>>>>>>>>>>>>>> after piping")

	if err != nil {
		logrus.Debugf("Error at handling copy between clients: %s ", err.Error())
		isConnected <- false
		return err
	}

	isConnected <- true
	return nil
}

//CreateConnectionRemoteV2 create a -R ssh connection
func CreateConnectionRemoteV2(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint, isConnected chan bool) error {
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
		isConnected <- false
		return err
	}
	sconn, chans, reqs, err := ssh.NewClientConn(connDialer, serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Error("Error at create new client conn")
		isConnected <- false
		return err
	}
	conn := ssh.NewClient(sconn, chans, reqs)
	logrus.Info("Connection established with ssh server..")

	// Listen on remote server port
	listener, err := conn.Listen("tcp", remoteEndpoint.String())
	defer closeListener(listener)
	if err != nil {
		logrus.Fatalf("Listen open port ON remote server error: %s", err)
		isConnected <- false
		return err
	}

	// handle incoming connections on reverse forwarded tunnel
	for {
		client, err := listener.Accept()
		if err != nil {
			logrus.Fatal(err)
			isConnected <- false
			return err
		}

		local, err := net.Dial("tcp", localEndpoint.String())
		if err != nil {
			logrus.Fatalf("Dial INTO remote service error: %s", err)
			isConnected <- false
			return err
		}

		go handleClientPipe(client, local, isConnected)
		// if err != nil {
		// 	isConnected <- false
		// 	return err
		// }
		// isConnected <- true

		break
	}
	logrus.Info("Exited for..")
	return nil
}

//CreateConnectionLocalV2 create a -L ssh connection
func CreateConnectionLocalV2(user string, password string, localEndpoint Endpoint, remoteEndpoint Endpoint, serverEndpoint Endpoint, isConnected chan bool) error {
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
		isConnected <- false
		return err
	}
	sconn, chans, reqs, err := ssh.NewClientConn(connDialer, serverEndpoint.String(), sshConfig)
	if err != nil {
		logrus.Error("Error at create new client conn")
		isConnected <- false
		return err
	}
	conn := ssh.NewClient(sconn, chans, reqs)
	defer closeConn(conn)
	logrus.Info("Connection established with ssh server..")

	listener, err := net.Listen("tcp", localEndpoint.String())
	logrus.Info("AFTER LISTEN COMMAND..")
	defer closeListener(listener)
	if err != nil {
		logrus.Fatal(err)
		isConnected <- false
		return err
	}

	isConnected <- true

	for {
		client, err := listener.Accept()
		logrus.Info("AFTER LISTENER ACCEPT")
		if err != nil {
			logrus.Fatal(err)
			// isConnected <- false
			return err
		}

		remote, err := conn.Dial("tcp", remoteEndpoint.String())
		logrus.Info("AFTER CONN DIALING")
		if err != nil {
			logrus.Fatalf("Dial INTO remote service error: %s", err)
			// isConnected <- false
			return err
		}

		// go handleClientPipe(client, remote)

		logrus.Info("]]]]]]]]]]]]]]]]]]]]]BEFORE HANDLING PIPE")
		go handleClientPipe(client, remote, isConnected)
		logrus.Info("]]]]]]]]]]]]]]]]]]]]]AFTER HANDLING PIPE")
		// if err != nil {
		// 	isConnected <- false
		// 	return err
		// }
		// isConnected <- true

		break
	}
	logrus.Info("Exited for..")
	return nil
}

func closeClient(client net.Conn) {
	defer recoveryFunction("closeClient()")
	defer client.Close()
}

func closeListener(listener net.Listener) {
	defer recoveryFunction("closeListener()")
	defer listener.Close()
}

func closeConn(conn ssh.Conn) {
	defer recoveryFunction("closeListener()")
	defer conn.Close()
}

func recoveryFunction(rss string) {
	logrus.Info("Error closing resources ... recovering from closing " + rss)
}
