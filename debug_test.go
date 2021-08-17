package geerpc

//func TestDebugHTTP_ServeHTTP(t *testing.T) {
//	ch := make(chan struct{})
//	addr := "127.0.0.1:9999"
//	go func() {
//		l, err := net.Listen("tcp", addr)
//		if err != nil {
//			t.Fatal("failed to listen tcp")
//		}
//		ch <- struct{}{}
//		Accept(l)
//	}()
//	<-ch
//	_, err := XDial("tcp@" + addr)
//	assert.Nil(t, err, "failed to connect tcp")
//}
