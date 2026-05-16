package presence

import (
	"sync"
	"testing"
	"time"
)

func TestMultiConnectionSameUser(t *testing.T) {
	tr := NewTracker()
	userID := "u1"

	c1 := Connection{ConnID: "c1", UserID: userID, RemoteAddr: "1.1.1.1", ConnectedAt: time.Now()}
	c2 := Connection{ConnID: "c2", UserID: userID, RemoteAddr: "2.2.2.2", ConnectedAt: time.Now().Add(time.Second)}

	tr.Register(c1)
	tr.Register(c2)

	if tr.Count() != 1 {
		t.Fatalf("expected 1 online user, got %d", tr.Count())
	}

	list := tr.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 user in list, got %d", len(list))
	}
	u := list[0]
	if u.UserID != userID {
		t.Errorf("expected userID=%s, got %s", userID, u.UserID)
	}
	if u.ConnCount != 2 {
		t.Errorf("expected connCount=2, got %d", u.ConnCount)
	}
}

func TestDisconnectOneConnectionStillOnline(t *testing.T) {
	tr := NewTracker()
	userID := "u1"

	tr.Register(Connection{ConnID: "c1", UserID: userID, RemoteAddr: "1.1.1.1", ConnectedAt: time.Now()})
	tr.Register(Connection{ConnID: "c2", UserID: userID, RemoteAddr: "2.2.2.2", ConnectedAt: time.Now()})

	tr.Unregister(userID, "c1")

	if tr.Count() != 1 {
		t.Fatalf("expected still 1 online user after disconnecting one conn, got %d", tr.Count())
	}

	list := tr.List()
	if len(list) != 1 || list[0].ConnCount != 1 {
		t.Fatalf("expected 1 user with 1 conn, got %+v", list)
	}
}

func TestDisconnectAllConnectionsOffline(t *testing.T) {
	tr := NewTracker()
	userID := "u1"

	tr.Register(Connection{ConnID: "c1", UserID: userID, RemoteAddr: "1.1.1.1", ConnectedAt: time.Now()})
	tr.Register(Connection{ConnID: "c2", UserID: userID, RemoteAddr: "2.2.2.2", ConnectedAt: time.Now()})

	tr.Unregister(userID, "c1")
	tr.Unregister(userID, "c2")

	if tr.Count() != 0 {
		t.Fatalf("expected 0 online users, got %d", tr.Count())
	}
	if len(tr.List()) != 0 {
		t.Fatal("expected empty list")
	}
}

func TestDifferentUsersIndependent(t *testing.T) {
	tr := NewTracker()

	tr.Register(Connection{ConnID: "c1", UserID: "u1", RemoteAddr: "1.1.1.1", ConnectedAt: time.Now()})
	tr.Register(Connection{ConnID: "c2", UserID: "u2", RemoteAddr: "2.2.2.2", ConnectedAt: time.Now()})

	if tr.Count() != 2 {
		t.Fatalf("expected 2 online users, got %d", tr.Count())
	}

	tr.Unregister("u1", "c1")

	if tr.Count() != 1 {
		t.Fatalf("expected 1 online user after u1 disconnected, got %d", tr.Count())
	}

	list := tr.List()
	if len(list) != 1 || list[0].UserID != "u2" {
		t.Fatalf("expected only u2 remaining, got %+v", list)
	}
}

func TestUnregisterNonexistentConnNoPanic(t *testing.T) {
	tr := NewTracker()
	tr.Unregister("no-such-user", "no-such-conn")
	if tr.Count() != 0 {
		t.Fatal("expected 0")
	}
}

func TestRegisterSameConnTwiceUpdates(t *testing.T) {
	tr := NewTracker()
	now := time.Now()

	tr.Register(Connection{ConnID: "c1", UserID: "u1", RemoteAddr: "1.1.1.1", ConnectedAt: now})
	tr.Register(Connection{ConnID: "c1", UserID: "u1", RemoteAddr: "2.2.2.2", ConnectedAt: now.Add(time.Hour)})

	list := tr.List()
	if len(list) != 1 || list[0].ConnCount != 1 {
		t.Fatalf("expected 1 user with 1 conn after re-register, got %+v", list)
	}
}

func TestListAggregationEarliestConnectedAt(t *testing.T) {
	tr := NewTracker()
	userID := "u1"
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)

	tr.Register(Connection{ConnID: "c1", UserID: userID, RemoteAddr: "1.1.1.1", ConnectedAt: early})
	tr.Register(Connection{ConnID: "c2", UserID: userID, RemoteAddr: "2.2.2.2", ConnectedAt: late})

	u := tr.List()[0]
	if !u.ConnectedAt.Equal(early) {
		t.Errorf("expected connected_at=%v (earliest), got %v", early, u.ConnectedAt)
	}
}

func TestThreadSafety(t *testing.T) {
	tr := NewTracker()
	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		id := string(rune('a' + i%26))
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				connID := uid + "-" + string(rune('0'+j%10))
				tr.Register(Connection{ConnID: connID, UserID: uid, ConnectedAt: time.Now()})
				if j%3 == 0 {
					tr.Unregister(uid, connID)
				}
			}
		}(id)
	}
	wg.Wait()
	// Just verify no panic and Count() doesn't crash
	_ = tr.Count()
	_ = tr.List()
}
