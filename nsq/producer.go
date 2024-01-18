package nsq

import (
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// Logger 实现nsq日志接口
type Logger struct {
	zl *zap.Logger
}

// Output nsq错误输出
func (l *Logger) Output(calldepth int, s string) error {
	l.zl.Error("NSQ error", zap.Int("call_depth", calldepth), zap.String("error", s))
	return nil
}

// Init 使用默认配置初始化
func Init(nsqd string, lookupd []string, consumers ...Consumer) error {
	return InitWithCfg(nsqd, lookupd, nsq.NewConfig(), consumers...)
}

// InitWithCfg 指定配置初始化
func InitWithCfg(nsqd string, lookupd []string, cfg *nsq.Config, consumers ...Consumer) (err error) {
	producer, err = nsq.NewProducer(nsqd, cfg)
	if err != nil {
		return
	}
	if err = producer.Ping(); err != nil {
		return
	}

	// 设置消费者
	if len(consumers) != 0 {
		if err = consumerSet(lookupd, consumers...); err != nil {
			return
		}
	}

	return
}

// SetErrLogger 设置错误日志
func SetErrLogger(l *zap.Logger) {
	if producer == nil {
		return
	}

	producer.SetLogger(&Logger{zl: l}, nsq.LogLevelError)
}

// Publish 同步推送消息到指定Topic
func Publish(topic string, msg []byte) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten init?)")
	}

	return producer.Publish(topic, msg)
}

// DeferredPublish 同步推送延迟消息到指定Topic
func DeferredPublish(topic string, msg []byte, duration time.Duration) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten init?)")
	}

	return producer.DeferredPublish(topic, duration, msg)
}

// NextAttemptDelay 一个帮助方法，用于返回下一次尝试的等待时间
func NextAttemptDelay(attempts uint16) time.Duration {
	var d time.Duration

	switch attempts {
	case 0, 1:
		d = 5 * time.Second
	case 2:
		d = 10 * time.Second
	case 3:
		d = 15 * time.Second
	case 4:
		d = 30 * time.Second
	case 5:
		d = time.Minute
	case 6:
		d = 2 * time.Minute
	case 7:
		d = 5 * time.Minute
	case 8:
		d = 10 * time.Minute
	case 9:
		d = 15 * time.Minute
	case 10:
		d = 30 * time.Minute
	default:
		d = time.Hour
	}

	return d
}
