package handler

import (
	"bufio"
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (env *Env) GetLog(c *gin.Context) {

	count, err := strconv.Atoi(c.Param("lines"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}

	var file *os.File
	file, err = os.Open(config.LogLocation)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	linecount, err := lineCounter(file)

	err = file.Close()
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	file, err = os.Open(config.LogLocation)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	var fileTextLines []string

	pos := linecount - count

	var n int
	for fileScanner.Scan() {
		n = n + 1
		if n >= pos {
			fileTextLines = append(fileTextLines, fileScanner.Text())
		}
	}

	Log.Debug(n)

	c.JSON(http.StatusOK, fileTextLines)
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
