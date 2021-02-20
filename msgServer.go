package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	heapSlice "github.com/Brotchu/heap-slice"
	"github.com/Brotchu/msgQ/msgqpb/msgq"
	"google.golang.org/grpc"
)

type server struct {
	MapQ map[string]*heapSlice.Heap
	Mut  *sync.Mutex
}

func (s *server) Ping(ctx context.Context, req *msgq.PingRequest) (*msgq.PingResponse, error) {
	res := &msgq.PingResponse{
		Pong: &msgq.Ping{
			Ack: true,
		},
	}
	return res, nil
}

func (s *server) CreateQ(ctx context.Context, req *msgq.CreateQRequest) (*msgq.CreateQResponse, error) {
	//TODO:
	k := req.GetQname()

	s.Mut.Lock()
	//lock mutex
	//check if name present
	if _, ok := s.MapQ[k]; ok {
		//unlock mutex
		s.Mut.Unlock()
		return nil, errors.New("Err: Queue Already exists")
	} else {
		s.MapQ[k] = heapSlice.NewHeap()
	}
	//if not present make a new one
	//else return error
	res := &msgq.CreateQResponse{Ack: &msgq.Ack{Ack: true}}
	s.Mut.Unlock()
	//unlock mutex

	// res := &msgq.CreateQResponse{Ack: &msgq.Ack{Ack: true}}
	return res, nil
}

func (s *server) DeleteQ(ctx context.Context, req *msgq.DeleteQRequest) (*msgq.DeleteQResponse, error) {
	//get q name
	qname := req.GetQname()

	//lock mutex
	s.Mut.Lock()

	//delete item from map
	delete(s.MapQ, qname)

	s.Mut.Unlock()
	//unlock

	res := &msgq.DeleteQResponse{
		Ack: &msgq.Ack{
			Ack: true,
		},
	}
	return res, nil
}

func (s *server) AddMessage(ctx context.Context, req *msgq.AddMessageRequest) (*msgq.AddMessageResponse, error) {
	qname := req.GetQname()
	//TODO:
	//lock
	s.Mut.Lock()
	if _, ok := s.MapQ[qname]; !ok {
		s.Mut.Unlock()
		return nil, errors.New("Err: Queue Does not Exist")
	}
	//check if q exists
	//return err if not
	p := req.GetQmsg().GetPriority()
	m := req.GetQmsg().GetMsg()
	err := s.MapQ[qname].Insert(int(p), m)
	fmt.Printf("%+v\n", s.MapQ[qname].HeapList) //TODO: remove , this is testing

	s.Mut.Unlock()
	if err != nil {
		return nil, err
	}
	res := &msgq.AddMessageResponse{Ack: &msgq.Ack{Ack: true}}
	return res, nil
	//unlock
}

func (s *server) GetMessage(ctx context.Context, req *msgq.GetMessageRequest) (*msgq.GetMessageResponse, error) {
	qname := req.GetQname()

	s.Mut.Lock()
	if _, ok := s.MapQ[qname]; !ok {
		s.Mut.Unlock()
		return nil, errors.New("Err: Queue Does not Exist")
	}
	msg, err := s.MapQ[qname].Delete()
	s.Mut.Unlock()
	s.MapQ[qname].Test() //TODO: remove; this is testing
	if err != nil {
		return nil, err
	}
	res := &msgq.GetMessageResponse{
		Qmsg: msg.Message,
	}
	// s.Mut.Unlock()
	return res, nil
}

func main() {
	portNum := os.Args[1]
	addr := "0.0.0.0:" + portNum

	fmt.Println("Starting Message queue server")
	lis, err := net.Listen("tcp", addr)
	must(err)

	s := grpc.NewServer()
	msgq.RegisterMsgQServiceServer(s, &server{
		MapQ: make(map[string]*heapSlice.Heap),
		Mut:  &sync.Mutex{},
	})

	if err := s.Serve(lis); err != nil {
		log.Fatal("Error : ", err)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal("Error : ", err)
		os.Exit(1)
	}
}
