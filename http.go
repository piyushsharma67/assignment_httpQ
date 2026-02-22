package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi"
)

// HTTPQ is the main broker struct that manages pub/sub channels and statistics.
// TODO: Add synchronization primitives to make stats thread-safe
// TODO: Add a data structure to manage topics/channels

const waitTime= 5* time.Second
type quedMessage struct{
	data []byte
	ack chan struct{}
}
type topic struct{
	message []quedMessage
	waiting []chan []byte
	mu sync.Mutex
}
type HTTPQ struct {
	topics map[string]*topic
	mu sync.Mutex
	RxBytes  int64 // number of bytes (message body) consumed
	TxBytes  int64 // number of bytes (message body) published
	PubFails int64 // number of publish failures (e.g., timeout)
	SubFails int64 // number of subscribe failures (e.g., timeout)
}

func (h *HTTPQ) Handler() http.Handler {
	if h.topics == nil{
		h.topics=make(map[string]*topic)
	}
	r := chi.NewRouter()

	// Stats endpoint
	r.Get("/stats", h.Stats().ServeHTTP)

	// Pub/Sub endpoints with topic parameter
	r.Get("/{topic}", h.Consume().ServeHTTP)
	r.Post("/{topic}", h.Publish().ServeHTTP)

	return r
}

func (h * HTTPQ)getTopic(name string)*topic{
	h.mu.Lock()
	defer h.mu.Unlock()

	t,ok:=h.topics[name]
	if !ok{
		t=&topic{}
		h.topics[name]=t
	}
	return t
}

// Publish handles POST /{topic} requests.
// It should block until a consumer receives the message or timeout (30s).
func (h *HTTPQ) Publish() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract topic from URL using chi.URLParam(r, "topic")
		// TODO: Read message body
		// TODO: Block until consumer receives or timeout
		// TODO: Update TxBytes on success, PubFails on timeout
		name:=chi.URLParam(r,"topic")
		t:=h.getTopic(name)

		body,err:=io.ReadAll(r.Body)
		if err!=nil{
			http.Error(w,err.Error(),http.StatusBadRequest)
		}

		timeout:=time.After(waitTime)

		t.mu.Lock()
		if len(t.waiting)>0{
			ch:=t.waiting[0]
			t.waiting=t.waiting[1:]
			t.mu.Unlock()
			ch<-body
			atomic.AddInt64(&h.TxBytes,int64(len(body)))
			w.WriteHeader(http.StatusOK)
			return 
		}
		ack:=make(chan struct{})
		t.message=append(t.message, quedMessage{
			data:body,
			ack: ack,
		})
		t.mu.Unlock()
		select{
			case <-ack:
				
				w.WriteHeader(http.StatusOK)
			case <- timeout:
				atomic.AddInt64(&h.PubFails,1)
				http.Error(w,"timeout waiting for consumer",http.StatusGatewayTimeout)
		}


	})
}

// Consume handles GET /{topic} requests.
// It should block until a producer sends a message or timeout (30s).
func (h *HTTPQ) Consume() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract topic from URL using chi.URLParam(r, "topic")
		// TODO: Block until producer sends or timeout
		// TODO: Write message to response
		// TODO: Update RxBytes on success, SubFails on timeout

		name:=chi.URLParam(r,"topic")
		t:=h.getTopic(name)
		
		timeout:=time.After(waitTime)

		t.mu.Lock()
		if len(t.message)>0{
			msg:=t.message[0]
			t.message=t.message[1:]
			t.mu.Unlock()
			w.Write(msg.data)
			
			atomic.AddInt64(&h.RxBytes,int64(len(msg.data)))
			atomic.AddInt64(&h.TxBytes,int64(len(msg.data)))
			close(msg.ack)
			return
		}

		waitCh:=make(chan []byte,1)
		t.waiting=append(t.waiting,waitCh)
		t.mu.Unlock()

		select{
			case msg:=<- waitCh:
				w.Write(msg)
				atomic.AddInt64(&h.RxBytes,int64(len(msg)))
			case <-timeout:
				atomic.AddInt64(&h.SubFails,1)
				http.Error(w,"timeout waiting for producer",http.StatusGatewayTimeout)
				return 
		}
	})
}

// Stats handles GET /stats requests.
// Returns JSON with rx_bytes, tx_bytes, pub_fails, sub_fails.
func (h *HTTPQ) Stats() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Return JSON response with stats
		// TODO: Ensure thread-safe reads of counters

		stats:=map[string]int64{
			"rx_bytes":atomic.LoadInt64(&h.RxBytes),
			"tx_bytes":atomic.LoadInt64(&h.TxBytes),
			"pub_fails":atomic.LoadInt64(&h.PubFails),
			"sub_fails":atomic.LoadInt64(&h.SubFails),
		}

		w.Header().Set("Content-Type","application/json")
		json.NewEncoder(w).Encode(stats)
	})
}
