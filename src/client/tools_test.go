package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"packet"
	"transport/flow"
)

func TestClearSession(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = "test"

	broker := flow.New().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := ClearSession(NewConfigWithClientID("tcp://localhost:"+port, "test"), 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClearRetainedMessage(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = nil
	publish.Message.Retain = true

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := ClearRetainedMessage(NewConfig("tcp://localhost:"+port), "test", 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestPublishMessage(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := PublishMessage(NewConfig("tcp://localhost:"+port), &publish.Message, 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestReceiveMessage(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}

	suback := packet.NewSubackPacket()
	suback.ID = 1
	suback.ReturnCodes = []uint8{0}

	publish := packet.NewPublishPacket()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	msg, err := ReceiveMessage(NewConfig("tcp://localhost:"+port), "test", 0, 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, publish.Message.String(), msg.String())

	safeReceive(done)
}
