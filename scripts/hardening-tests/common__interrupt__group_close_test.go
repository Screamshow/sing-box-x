package interrupt

import (
 "net"
 "testing"
 "time"
 "github.com/stretchr/testify/require"
)

type stalledClose struct {
 net.Conn
 entered chan struct{}
 release chan struct{}
}
func (c *stalledClose) Close() error {
 close(c.entered)
 <-c.release
 return nil
}

func TestStalledCloseDoesNotLockGroup(t *testing.T) {
 for _, interrupt := range []bool{false,true} {
  t.Run(map[bool]string{false:"close",true:"interrupt"}[interrupt],func(t *testing.T) {
   group:=NewGroup()
   underlying:=&stalledClose{entered:make(chan struct{}),release:make(chan struct{})}
   defer close(underlying.release)
   conn:=group.NewConn(underlying,false)
   done:=make(chan struct{})
   go func() { if interrupt { group.Interrupt(false) } else { conn.Close() }; close(done) }()
   select { case <-underlying.entered: case <-time.After(time.Second): t.Fatal("close did not begin") }
   added:=make(chan struct{})
   go func() { group.NewConn(nil,true); close(added) }()
   select { case <-added: case <-time.After(time.Second): t.Fatal("stalled close blocked group registration") }
   group.access.Lock()
   require.Equal(t,1,group.connections.Len())
   group.access.Unlock()
  })
 }
}
