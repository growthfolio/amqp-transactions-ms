package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger é uma instância global de logrus.Logger
var Logger = logrus.New()

func init() {
	// Configuração do formato do log
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.InfoLevel)
}
