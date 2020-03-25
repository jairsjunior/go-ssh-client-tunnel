package util

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

//PublicKeyFile read public key file
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}
	return ssh.PublicKeys(key)
}

//ConfigViper configure viper
func ConfigViper() {
	viper.SetEnvPrefix("ssh")

	viper.SetDefault("SERVER_HOST", "tunnel-convid.sandman.cloud")
	viper.SetDefault("SERVER_PORT", "2432")
	viper.SetDefault("LOCAL_HOST", "localhost")
	viper.SetDefault("LOCAL_PORT", "5000")
	viper.SetDefault("REMOTE_HOST", "localhost")
	viper.SetDefault("REMOTE_PORT", "11290")
	viper.SetDefault("USER", "MZL2321")
	viper.SetDefault("PASSWORD", "13199857000124")

	viper.SetDefault("MODE", "local")

	viper.SetDefault("LOG_LEVEL", "info")

	viper.AutomaticEnv()
}

//ConfigLogrus .
func ConfigLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	level, e := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if e != nil {
		logrus.Errorf("Error parsing log level, setting log level to 'info'. error: %s", e.Error())
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}
