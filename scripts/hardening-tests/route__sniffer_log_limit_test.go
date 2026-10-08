package route

import (
 "bytes"
 "context"
 "strings"
 "testing"
 "time"

 "github.com/sagernet/sing-box/common/sniff"
 "github.com/sagernet/sing-box/log"
 "github.com/stretchr/testify/require"
)

func TestSnifferStackLoggingIsBounded(t *testing.T) {
 var output bytes.Buffer
 factory:=log.NewDefaultFactory(context.Background(),log.Formatter{BaseTime:time.Now(),DisableColors:true},&output,"",nil,false)
 require.NoError(t,factory.Start())
 defer factory.Close()
 router:=&Router{logger:factory.NewLogger("router")}
 for range 5 {
  err:=&sniff.SnifferPanicError{Value:"test failure",Stack:[]byte("unique-stack-marker")}
  router.logSnifferPanic(context.Background(),err)
  require.Nil(t,err.Stack,"connection metadata must not retain a stack")
 }
 require.Equal(t,1,strings.Count(output.String(),"unique-stack-marker"))
 require.Equal(t,4,router.snifferPanics.skipped)
 router.snifferPanics.logged=time.Now().Add(-2*time.Minute)
 router.logSnifferPanic(context.Background(),&sniff.SnifferPanicError{Value:"test failure",Stack:[]byte("unique-stack-marker")})
 require.Equal(t,2,strings.Count(output.String(),"unique-stack-marker"))
 require.Contains(t,output.String(),"4 more since the last stack was logged")
}
