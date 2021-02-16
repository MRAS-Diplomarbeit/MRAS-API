package mysql

// github.com/onrik/gorm-logrus fork

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"regexp"
	"time"

	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type logger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func NewLogger() *logger {
	return &logger{
		SkipErrRecordNotFound: true,
	}
}

func (l *logger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *logger) Info(ctx context.Context, s string, args ...interface{}) {
	Log.WithFields(logrus.Fields{"module": "gorm"}).WithContext(ctx).Infof(s, args)
}

func (l *logger) Warn(ctx context.Context, s string, args ...interface{}) {
	Log.WithFields(logrus.Fields{"module": "gorm"}).WithContext(ctx).Warnf(s, args)
}

func (l *logger) Error(ctx context.Context, s string, args ...interface{}) {
	Log.WithFields(logrus.Fields{"module": "gorm"}).WithContext(ctx).Errorf(s, args)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{"module": "gorm"}

	//remove JWT-Tokens from Log
	var re = regexp.MustCompile(`(='([A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=])+')`)
	cleansql := re.ReplaceAllString(sql, `TOKEN`)

	//remove reset-code from Log
	re = regexp.MustCompile(`('(\w+(-|')){10})`)
	cleansql = re.ReplaceAllString(cleansql, `RESETCODE`)

	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[logrus.ErrorKey] = err
		Log.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		Log.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", cleansql, elapsed)
		return
	}

	Log.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", cleansql, elapsed)
}
