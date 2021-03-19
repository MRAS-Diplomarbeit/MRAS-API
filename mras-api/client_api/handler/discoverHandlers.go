package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"net/http"
	"time"
)

func (env *Env) Discover(c *gin.Context) {

	var speaker mysql.Speaker
	speaker.IPAddress = c.ClientIP()

	var exists int64

	result := env.db.Model(&speaker).Where("ip_address", speaker.IPAddress).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		speaker.Name = "NEW SPEAKER"
		speaker.Description = "SETUP NEW SPEAKER"
		speaker.LastLifesign = time.Now()
		speaker.Alive = true

		result = env.db.Save(&speaker)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	} else {
		result = env.db.Find(&speaker).Update("last_lifesign", time.Now())
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

	}

}
