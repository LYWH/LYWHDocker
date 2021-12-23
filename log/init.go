package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	Mylog.Out = os.Stdout
	Mylog.SetLevel(logrus.WarnLevel)

	Mylog.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Mylog.SetReportCaller(true)
}
