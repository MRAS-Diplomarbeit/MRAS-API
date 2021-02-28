package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func (env *Env) LogMessage(c *gin.Context) {

	if Log.GetLevel() != logrus.DebugLevel {
		return
	}

	type logmessage struct {
		Lines []string `json:"lines"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request logmessage
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	for _, message := range request.Lines {
		Log.WithField("module", "speaker").Debug(message)
	}

}
