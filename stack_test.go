package hastack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	var err error
	var s *Stack

	s, err = Connect("testqueue", "localhost:6379", "", 3, 100)
	assert.NoError(t, err)

	var inboxStart int64
	var outboxStart int64

	inboxStart, _ = s.InboxLength()
	outboxStart, _ = s.OutboxLength()

	err = s.Push("message")
	assert.NoError(t, err)
	var i int64
	i, err = s.InboxLength()
	assert.NoError(t, err)
	assert.Equal(t, inboxStart+1, i, "Inbox length should be 1")
	var recv string
	err = s.Pop(&recv)
	assert.NoError(t, err)
	assert.Equal(t, "message", recv, "sent and returned should be the same")
	i, err = s.InboxLength()
	assert.NoError(t, err)
	assert.Equal(t, inboxStart, i, "Inbox length should be O")
	i, err = s.OutboxLength()
	assert.NoError(t, err)
	assert.Equal(t, outboxStart+1, i, "Outbox length should be 1")
}

func TestBlockingPop(t *testing.T) {
	var err error
	var s *Stack

	s, err = Connect("testqueue", "localhost:6379", "", 3, 100)
	assert.NoError(t, err)

	go func() {
		time.Sleep(5 * time.Second)
		s.Push("message")
	}()

	var recv string
	err = s.BPop(&recv)
	assert.NoError(t, err)
	assert.Equal(t, "message", recv, "sent and returned should be the same")

}
