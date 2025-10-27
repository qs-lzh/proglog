package discovery_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"

	. "github.com/qs-lzh/proglog/internal/discovery"
)

func TestMembership(t *testing.T) {
	members, handler := setupMember(t, nil)
	members, _ = setupMember(t, members)
	members, _ = setupMember(t, members)

	require.Eventually(
		t,
		func() bool {
			return 2 == len(handler.joins) && 3 == len(members[0].Members()) && 0 == len(handler.leaves)
		},
		3*time.Second,
		250*time.Millisecond,
	)
	require.NoError(t, members[2].Leave())
	require.Eventually(
		t,
		func() bool {
			return 2 == len(handler.joins) && 3 == len(members[0].Members()) && serf.StatusLeft == members[0].Members()[2].Status && 1 == len(handler.leaves)
		},
		3*time.Second,
		250*time.Millisecond,
	)
	require.Equal(t, fmt.Sprintf("%d", 2), <-handler.leaves)
}

func setupMember(t *testing.T, members []*Membership) ([]*Membership, *handler) {
	id := len(members)
	ports := dynaport.Get(1)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
	tags := map[string]string{
		"rpc_addr": addr,
	}
	config := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}
	handler := &handler{}
	if len(members) == 0 {
		handler.joins = make(chan map[string]string, 3)
		handler.leaves = make(chan string, 3)
	} else {
		config.StartJoinAddrs = []string{members[0].BindAddr}
	}
	member, err := New(handler, config)
	require.NoError(t, err)
	members = append(members, member)
	return members, handler
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(id, addr string) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}
	return nil
}

func (h *handler) Leave(id string) error {
	if h.leaves != nil {
		h.leaves <- id
	}
	return nil
}
