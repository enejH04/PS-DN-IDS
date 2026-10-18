// Komunikacija po protokolu gRPC
// strežnik

package main

import (
	"api/grpc/protobufStorage"
	"api/storage"
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Server(url string) {
	// pripravimo strežnik gRPC
	grpcServer := grpc.NewServer()

	// pripravimo strukuro za streženje metod CRUD na shrambi TodoStorage
	crudServer := NewServerCRUD()

	// streženje metod CRUD na shrambi TodoStorage povežemo s strežnikom gRPC
	protobufStorage.RegisterCRUDServer(grpcServer, crudServer)

	// izpišemo ime strežnika
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	// odpremo vtičnico
	listener, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	fmt.Printf("gRPC server listening at %v%v\n", hostName, url)
	// začnemo s streženjem
	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}

// stuktura za strežnik CRUD za shrambo TodoStorage
type serverCRUD struct {
	protobufStorage.UnimplementedCRUDServer
	todoStore storage.TodoStorage
	// Dodatni vnosi za implementacijo naročanja na dogodke
	subsMutex sync.Mutex
	// Rezina kanalov za obveščanje naročnikov
	subscribers map[int]chan *protobufStorage.TodoEvent
	nextSubID   int
}

// pripravimo nov strežnik CRUD za shrambo TodoStorage
func NewServerCRUD() *serverCRUD {
	todoStorePtr := storage.NewTodoStorage()
	// CRUD streznik
	return &serverCRUD{
		UnimplementedCRUDServer: protobufStorage.UnimplementedCRUDServer{},
		todoStore:               *todoStorePtr,
		// Narocniki
		subscribers: make(map[int]chan *protobufStorage.TodoEvent),
		// ID naslednjega narocnika
		nextSubID: 0,
	}
}

// metode strežnika CRUD za shrambo TodoStorage
// odtis metode (argumenti, vrnjene vrednosti) so podane tako, kot zahtevajo paketi, ki jih je ustvaril prevajalnik protoc
// vse metode so samo ovojnice za klic metod v TodoStorage
// metode se izvajajo v okviru omejitev izvajalnega okolju ctx
func (s *serverCRUD) Create(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todoPtr := &storage.Todo{Task: in.Task, Completed: in.Completed}
	err := s.todoStore.Create(todoPtr, &ret)

	if err == nil {
		event := &protobufStorage.TodoEvent{T: &protobufStorage.Todo{Task: in.Task, Completed: in.Completed}, Action: "create"}
		s.broadcastEvent(event)
	}

	return &emptypb.Empty{}, err
}

func (s *serverCRUD) Read(ctx context.Context, in *protobufStorage.Todo) (*protobufStorage.TodoStorage, error) {
	dict := make(map[string](storage.Todo))
	err := s.todoStore.Read(&storage.Todo{Task: in.Task, Completed: in.Completed}, &dict)
	pbDict := protobufStorage.TodoStorage{}

	for k, v := range dict {
		pbDict.Todos = append(pbDict.Todos, &protobufStorage.Todo{Task: k, Completed: v.Completed})
	}

	return &pbDict, err
}

func (s *serverCRUD) Update(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todoPtr := &storage.Todo{Task: in.Task, Completed: in.Completed}
	err := s.todoStore.Update(todoPtr, &ret)

	if err == nil {
		event := &protobufStorage.TodoEvent{T: &protobufStorage.Todo{Task: in.Task, Completed: in.Completed}, Action: "update"}
		s.broadcastEvent(event)
	}

	return &emptypb.Empty{}, err
}

func (s *serverCRUD) Delete(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todoPtr := &storage.Todo{Task: in.Task, Completed: in.Completed}
	err := s.todoStore.Delete(todoPtr, &ret)

	if err == nil {
		event := &protobufStorage.TodoEvent{T: &protobufStorage.Todo{Task: in.Task, Completed: in.Completed}, Action: "delete"}
		s.broadcastEvent(event)
	}

	return &emptypb.Empty{}, err
}

func (s *serverCRUD) ReadAll(e *emptypb.Empty, stream grpc.ServerStreamingServer[protobufStorage.Todo]) error {
	dict := make(map[string](storage.Todo))
	// Read data from the storage
	err := s.todoStore.Read(&storage.Todo{}, &dict)

	for _, todo := range dict {
		todo_temp := &protobufStorage.Todo{Task: todo.Task, Completed: todo.Completed}
		// Send data to client in a stream
		if err := stream.Send(todo_temp); err != nil {
			return err
		}
	}

	return err
}

// Broadcastamo event vsem naročnikom
func (s *serverCRUD) broadcastEvent(event *protobufStorage.TodoEvent) {
	// Rezina dolzine 0 s kapaciteto, ki je enaka stevilu narocnikov
	subs := make([]chan *protobufStorage.TodoEvent, 0, len(s.subscribers))
	s.subsMutex.Lock()
	for _, ch := range s.subscribers {
		subs = append(subs, ch)
	}
	s.subsMutex.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *serverCRUD) Subscribe(e *emptypb.Empty, stream grpc.ServerStreamingServer[protobufStorage.TodoEvent]) error {
	// Dodamo novega narocnika
	streamChan := make(chan *protobufStorage.TodoEvent, 20)
	s.subsMutex.Lock()
	subId := s.nextSubID
	s.nextSubID++
	s.subscribers[subId] = streamChan
	s.subsMutex.Unlock()

	// Na koncu se pocistimo za sabo
	defer func() {
		s.subsMutex.Lock()
		delete(s.subscribers, subId)
		s.subsMutex.Unlock()
	}()

	// Sedaj moramo posiljati
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-streamChan:
			// Ko preberemo ven
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}
