package iot
/*
import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/sirupsen/logrus"
	"sync"
)

type RocketMq struct {
	ctx          context.Context
	cancel       context.CancelFunc
	ns           []string
	broker       []string
	accessKey    string
	accessSecret string
	cluster      string
	logger       *logrus.Logger

	admin    admin.Admin
	producer rocketmq.Producer
	retry    int
	once     sync.Once
}

func NewRocketMq(parent context.Context, l *logrus.Logger, logLevel string,
	cluster, accessKey, accessSecret string, ns, broker []string) (Iot, error) {
	admin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(ns)))
	if err != nil {
		return nil, fmt.Errorf("newRocketMq.NewAdmin:%v", err.Error())
	}
	rlog.SetLogLevel(logLevel)
	ctx, cancel := context.WithCancel(parent)
	return &RocketMq{
		logger:       l,
		ctx:          ctx,
		cancel:       cancel,
		cluster:      cluster,
		retry:        3,
		ns:           ns,
		broker:       broker,
		accessKey:    accessKey,
		accessSecret: accessSecret,
		admin:        admin,
	}, nil
}

func (rmq *RocketMq) Close() {
	rmq.once.Do(func() {
		rmq.cancel()
		rmq.producer.Shutdown()
		rmq.admin.Close()
	})
}

func (rmq *RocketMq) CreateTopic(topics ...string) error {
	for _, topic := range topics {
		return rmq.admin.CreateTopic(rmq.ctx,
			admin.WithTopicCreate(topic),
			admin.WithBrokerAddrCreate(rmq.broker[0]), //fixme
		)
	}
	return nil
}

func (rmq *RocketMq) CreateProducer(group string) error {
	options := make([]producer.Option, 0)
	options = append(options, producer.WithNsResolver(primitive.NewPassthroughResolver(rmq.ns)),
		producer.WithRetry(rmq.retry),
		producer.WithGroupName(group),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: rmq.accessKey,
			SecretKey: rmq.accessSecret,
		}))

	var err error
	rmq.producer, err = rocketmq.NewProducer(options...)
	if err != nil {
		return fmt.Errorf("rmqRegister.NewProducer:%v", err.Error())
	}
	if err = rmq.producer.Start(); err != nil {
		return fmt.Errorf("rmqRegister.rocketProducer.start: %v", err.Error())
	}
	return nil
}

func (rmq *RocketMq) Subscribe(topic, group string, cb func(*Message)) {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNsResolver(primitive.NewPassthroughResolver(rmq.ns)),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithGroupName(group),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: rmq.accessKey,
			SecretKey: rmq.accessSecret,
		}),
	)
	if err != nil {
		rmq.logger.Fatal("subscribe: %v: %v", topic, err.Error())
	}
	err = c.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context,
		messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, m := range messages {
			rmq.logger.Debug("subscribe:", m.String())
			mb := &Message{
				Topic:    topic,
				Id:       m.MsgId,
				BornHost: m.BornHost,
				Body:     m.Body,
			}
			cb(mb)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		rmq.logger.Error("subscribe: %v: %v", topic, err.Error())
		return
	}

	ctx, cancel := context.WithCancel(rmq.ctx)
	defer cancel()

	if err = c.Start(); err != nil {
		rmq.logger.Error("Subscribe.Start.", topic, ":", err.Error())
		return
	}

	rmq.logger.Info("subscribe: ", topic)
	<-ctx.Done()
	_ = c.Shutdown()
	//理论上 这个时候服务要重启下
	rmq.logger.Warn("subscribe stop ", topic)
}

func (rmq *RocketMq) Forward(aReply []byte, forwarders ...string) {
	for _, t := range forwarders {
		if err := rmq.Send(aReply, false, t); err != nil {
			rmq.logger.Error("forward", err.Error())
		}
	}
}

func (rmq *RocketMq) Send(aReply []byte, createTopic bool, topic string) error {
	//发送配置到路由器
	if createTopic {
		if err := rmq.CreateTopic(topic); err != nil {
			return fmt.Errorf("send.createTopic:%v", err.Error())
		}
	}

	if rmq.producer == nil {
		return fmt.Errorf("no producer")
	}
	m := primitive.NewMessage(topic, aReply)
	rmq.logger.Info("Send:", m.String())
	if err := rmq.producer.SendOneWay(rmq.ctx, m); err != nil {
		return fmt.Errorf("send.SendOneWay:%v", err.Error())
	}
	return nil
}

 */
