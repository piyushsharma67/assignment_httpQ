package main_test

import (
	"assignment"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func NewTestServer()(*main.HTTPQ,*httptest.Server){
	h:=&main.HTTPQ{}
	srv:=httptest.NewServer(h.Handler())
	return h,srv
	
}

// TestPublish verifies that publishing blocks when no consumer is waiting
// and times out appropriately.
func TestPublish(t *testing.T) {
	// TODO: Setup test server
	// TODO: Publish without a consumer and verify it blocks
	// TODO: Verify timeout behavior (PubFails incremented)
	h,srv:=NewTestServer()
	defer srv.Close()
	client:=&http.Client{}
	req,_:=http.NewRequest(http.MethodPost,srv.URL+"/topic1",bytes.NewBufferString("hello"))
	start:=time.Now()

	resp,err:=client.Do(req)

	if err!=nil{
		t.Fatal(err)
	}

	defer resp.Body.Close()

	elapsedTime:=time.Since(start)
	if elapsedTime< 4*time.Second{
		t.Fatalf("publish did not block long enough required 30 seconds but blocked only for %v",elapsedTime)
	}

	if resp.StatusCode!=http.StatusGatewayTimeout{
		t.Fatalf("expected timeout status but got %d",resp.StatusCode)
	}

	if h.PubFails ==0{
		t.Fatalf("expected pubFail increment")
	}
}

// TestConsume verifies that consuming blocks when no producer is waiting
// and times out appropriately.
func TestConsume(t *testing.T) {
	// TODO: Setup test server
	// TODO: Consume without a producer and verify it blocks
	// TODO: Verify timeout behavior (SubFails incremented)
	h,srv:=NewTestServer()
	defer srv.Close()
	start:=time.Now()
	client:=&http.Client{}
	resp,err:=client.Get(srv.URL+"/topic1")
	if err!=nil{
		t.Fatal(err)
	}


	defer resp.Body.Close()

	elapsed:=time.Since(start)

	if elapsed< 4*time.Second{
		t.Fatalf("consumer did not block long enough required 30 seconds but blocked only for %v",elapsed)
	}

	if resp.StatusCode!=http.StatusGatewayTimeout{
		t.Fatalf("expected timeout status but got %d",resp.StatusCode)
	}

	if h.SubFails ==0{
		t.Fatalf("expected pubFail increment")
	}

}

// TestPublishAndConsume verifies the basic pub/sub flow:
// producer and consumer should rendezvous and exchange the message.
func TestPublishAndConsume(t *testing.T) {
	// TODO: Setup test server
	// TODO: Start consumer in goroutine (will block)
	// TODO: Publish message
	// TODO: Verify consumer receives correct message
	// TODO: Verify stats (RxBytes, TxBytes)
	h,srv:=NewTestServer()
	defer srv.Close()
	client:=&http.Client{}

	resultCh:=make(chan string)

	go func(){
		resp,err:=client.Get(srv.URL+"/topic1")
		if err!=nil{
			t.Errorf("consumer failed %s",err.Error())
			return
		}

		defer resp.Body.Close()
		body,_:=io.ReadAll(resp.Body)

		resultCh<-string(body)
	}()

	time.Sleep(100*time.Millisecond)

	msg:="hello"

	req,_:=http.NewRequest(
		http.MethodPost,
		srv.URL+"/topic1",
		bytes.NewBufferString(msg),
	)

	resp,err:=client.Do(req)
	if err!=nil{
		t.Fatalf("publisher failed with : %v",err)
	}

	resp.Body.Close()
	select{
		case result:=<-resultCh:
			if result != msg{
				t.Fatalf("wanted %s, recieved %s",msg,result)
			}
		case <-time.After(2*time.Second):
			t.Fatalf("did not recieved msg")
	}

	if h.RxBytes ==0{
		t.Fatalf("expected rxByte >0",)
	}
	if h.TxBytes ==0{
		t.Fatalf("exptected txByte > 0")

	}
}

// TestMultipleTopics verifies that different topics are independent.
func TestMultipleTopics(t *testing.T) {
	// TODO: Publish to topic A
	// TODO: Consume from topic B (should block/timeout)
	// TODO: Consume from topic A (should receive message)
	h,srv:=NewTestServer()
	defer srv.Close()
	client:=&http.Client{}
	req,_:=http.NewRequest(
		http.MethodPost,
		srv.URL+"/topicA",
		bytes.NewBufferString("messageA"),
	)

	go func(){
		resp,err:=client.Do(req)
		if err!=nil{
			resp.Body.Close()
		}
	}()
	time.Sleep(100*time.Millisecond)

	start:=time.Now()
	respB,err:=client.Get(srv.URL+"/topicB")

	if err!=nil{
		t.Errorf("consume topicB failed %v",err)
	}

	defer respB.Body.Close()
	elapsed:=time.Since(start)

	if elapsed< 4*time.Second{
		t.Fatalf("topicB should have blocked and waited for 4 second but it returned in %v",elapsed)
	}
	if respB.StatusCode != http.StatusGatewayTimeout{
		t.Fatalf("expected timeout status but got %v",respB.StatusCode)
	}

	respA,err:=client.Get(srv.URL+"/topicA")
	if err!=nil{
		t.Errorf("consume topicA failed %v",err)
	}
	
	bodyA,_:=io.ReadAll(respA.Body)
	respA.Body.Close()

	if string(bodyA)!="messageA" {
		t.Fatalf("expected messageA got %s",string(bodyA))
	}

	if h.RxBytes == 0 || h.TxBytes==0{
		t.Fatalf("expected stats update")
	}

}

// TestFIFOOrdering verifies that messages are delivered in order.
func TestFIFOOrdering(t *testing.T) {
	// TODO: Setup multiple producers sending messages 1, 2, 3
	// TODO: Consume and verify order is preserved

	_,srv:=NewTestServer()

	defer srv.Close()
	client:=&http.Client{}

	messagesArr:=[]string{"1","2","3"}

	for _, message:=range messagesArr{
		req,_:=http.NewRequest(
			http.MethodPost,
			srv.URL+"/topicA",
			bytes.NewBufferString(message),
		)

		
		resp,err:=client.Do(req)
		if err!=nil{
			t.Fatal(err)
		}
		resp.Body.Close()
		
	}

	time.Sleep(100*time.Millisecond)

	for i,expectedMsg:=range messagesArr{
		resp,err:=client.Get(srv.URL+"/topicA")
		if err!=nil{
			t.Fatalf("consumed %d failed %v",i,err)
		}

		body,_:=io.ReadAll(resp.Body)
		resp.Body.Close()

		got := string(body)

		if got != expectedMsg{
			t.Fatalf("FIFO ordering viiolated!!")
		}
	}
}

// TestConcurrentAccess verifies thread-safety under concurrent load.
func TestConcurrentAccess(t *testing.T) {
	// TODO: Launch multiple producers and consumers concurrently
	// TODO: Verify no race conditions (run with -race flag)
	// TODO: Verify stats are accurate

	h,srv:=NewTestServer()
	defer srv.Close()

	client:=&http.Client{}

	producers:=20
	consumers:=20

	done:=make(chan bool)

	for _=range consumers{
		go func(){
			resp,err:=client.Get(srv.URL+"/topicA")

			if err!=nil{
				resp.Body.Close()
			}

			done<-true
		}()
	}
	for _=range consumers{
		go func(){
			req,_:=http.NewRequest(
				http.MethodPost,
				srv.URL+"/topicA",
				bytes.NewBufferString("msg"),
			)
			resp,err:=client.Do(req)
			if err!=nil{
				resp.Body.Close()

			}
			done<-true
		}()
	}

	for i:=0;i<producers+consumers;i++{
		<-done
	}

	if h.TxBytes==0 || h.RxBytes ==0{
		t.Fatalf("expected stats updated under concurrent load")	
	}
	
}

// TestStats verifies the /stats endpoint returns correct values.
func TestStats(t *testing.T) {
	// TODO: Perform some pub/sub operations
	// TODO: GET /stats and verify JSON response
	// TODO: Verify rx_bytes, tx_bytes, pub_fails, sub_fails

	h,srv:=NewTestServer()
	defer srv.Close()

	client:=&http.Client{}

	go func(){
		resp,err:=client.Get(srv.URL+"/topicA")
		if err!=nil{
			resp.Body.Close()
		}
	}()

	time.Sleep(100*time.Millisecond)
	req,_:=http.NewRequest(
		http.MethodPost,
		srv.URL+"/topicA",
		bytes.NewBufferString("message"),
	)

	resp,err :=  client.Do(req)
	if err!=nil{
		t.Fatal(err)
	}
	resp.Body.Close()

	statusResp,err:=client.Get(srv.URL+"/stats")

	if err!=nil{
		t.Fatal(err)
	}

	defer statusResp.Body.Close()
	if statusResp.StatusCode!=http.StatusOK{
		t.Fatalf("expected 200 recieved %d",statusResp.StatusCode)
	}

	body,_:=io.ReadAll(statusResp.Body)

	var stats map[string]int64
	err=json.Unmarshal(body,&stats)

	if err!=nil{
		t.Fatalf("invalid json provided")
	}

	if stats["rx_bytes"]==0{
		t.Fatalf("expected rx_bytes > 0")
	}
	if stats["tx_bytes"]==0{
		t.Fatalf("expected tx_bytes > 0")
	}
	if stats["rx_bytes"]!=h.RxBytes{
		t.Fatalf("rx_bytes mismatchj")
	}
	if stats["tx_bytes"]!=h.TxBytes{
		t.Fatalf("tx_bytes mismatch")
	}
}
