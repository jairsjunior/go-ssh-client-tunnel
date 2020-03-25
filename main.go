package main

import (
	client "github.com/jairsjunior/go-ssh-client-tunnel/clientv2"
	"github.com/jairsjunior/go-ssh-client-tunnel/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.Info("=========> Starting SSH TUNNEL Application <=========")
	logrus.Trace("Loading Environments using Viper")
	util.ConfigViper()
	logrus.Trace("Configure Logrus")
	util.ConfigLogrus()

	// local service to be forwarded
	var localEndpoint = client.Endpoint{
		Host: viper.GetString("LOCAL_HOST"),
		Port: viper.GetInt("LOCAL_PORT"),
	}

	// remote SSH server
	var serverEndpoint = client.Endpoint{
		Host: viper.GetString("SERVER_HOST"),
		Port: viper.GetInt("SERVER_PORT"),
	}

	// remote forwarding port (on remote SSH server network)
	var remoteEndpoint = client.Endpoint{
		Host: viper.GetString("REMOTE_HOST"),
		Port: viper.GetInt("REMOTE_PORT"),
	}

	user := viper.GetString("USER")
	password := viper.GetString("PASSWORD")

	mode := viper.GetString("MODE")

	logrus.Infof("Server: %s", serverEndpoint.String())
	logrus.Infof("Local: %s", localEndpoint.String())
	logrus.Infof("Remote: %s", remoteEndpoint.String())

	isConnected := make(chan bool)

	if mode == "remote" {
		logrus.Info("MODE: REMOTE")
		for {
			go func() {
				err := client.CreateConnectionRemoteV2(user, password, localEndpoint, remoteEndpoint, serverEndpoint, isConnected)
				if err != nil {
					return
				}
			}()
			connected := <-isConnected
			logrus.Infof("Connected: %s", connected)
			if connected {
				connected = <-isConnected
			} else {
				logrus.Info("Retry to connect..")
			}
		}
	} else if mode == "local" {
		logrus.Info("MODE: LOCAL")
		for {
			go func() {
				err := client.CreateConnectionLocalV2(user, password, localEndpoint, remoteEndpoint, serverEndpoint, isConnected)
				if err != nil {
					return
				}
			}()
			connected := <-isConnected
			logrus.Infof("Connected: %s", connected)
			if connected {
				connected = <-isConnected
			} else {
				logrus.Info("Retry to connect..")
			}
		}
	}
}
