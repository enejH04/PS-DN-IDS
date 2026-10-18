// Komunikacija po protokolu gRPC
// odjemalec

package main

import (
	"api/grpc/protobufStorage"
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Client(url string) {
	// vzpostavimo povezavo s strežnikom
	fmt.Printf("gRPC client connecting to %v\n", url)
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// vzpostavimo izvajalno okolje
	// gRPC stream lives only as long as its context.Context is alive
	contextCRUD, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// vzpostavimo vmesnik gRPC
	grpcClient := protobufStorage.NewCRUDClient(conn)

	// pripravimo strukture, ki jih uporabljamo kot argumente pri klicu oddaljenih metod
	lecturesCreate := protobufStorage.Todo{Task: "predavanja", Completed: false}
	lecturesUpdate := protobufStorage.Todo{Task: "predavanja", Completed: true}
	homeworkCreate := protobufStorage.Todo{Task: "domaca naloga", Completed: false}
	homeworkUpdate := protobufStorage.Todo{Task: "domaca naloga", Completed: true}
	practicals := protobufStorage.Todo{Task: "vaje", Completed: false}
	// readAll := protobufStorage.Todo{Task: "", Completed: false}
	// Uporabljeno za sinhronizacijo
	subscribeChan := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(2)

	// Testiranje domace naloge 8
	go func() {
		defer wg.Done()
		<-subscribeChan
		if _, err := grpcClient.Create(contextCRUD, &lecturesCreate); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// ustvarimo zapis
		if _, err := grpcClient.Create(contextCRUD, &practicals); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// posodobimo zapis
		if _, err := grpcClient.Update(contextCRUD, &lecturesUpdate); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// izbrišemo zapis
		if _, err := grpcClient.Delete(contextCRUD, &practicals); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// ustvarimo zapis
		if _, err := grpcClient.Create(contextCRUD, &homeworkCreate); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// posodobimo zapis - konec DN
		if _, err := grpcClient.Update(contextCRUD, &homeworkUpdate); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}()

	go func() {
		defer wg.Done()
		// Narocimo se
		stream, err := grpcClient.Subscribe(contextCRUD, &emptypb.Empty{})

		if err != nil {
			panic(err)
		}

		// Sedaj lahko odjemalec zacne izvajati akcije
		close(subscribeChan)

		for {
			event, err := stream.Recv()

			// Ce je konec
			if err != nil {
				// Context deadline exceeded
				fmt.Println("KONEC POVEZAVE")
				break
			}

			fmt.Println(event)
		}
	}()

	wg.Wait()

	// // preberemo vse zapise
	// // fmt.Print("7. Read *: ")
	// // if response, err := grpcClient.Read(contextCRUD, &readAll); err == nil {
	// // 	fmt.Println(response.Todos, ": done")
	// // } else {
	// // 	panic(err)
	// // }

	// fmt.Print("7. Create: ")
	// if _, err := grpcClient.Create(contextCRUD, &homeworkCreate); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("done")

	// fmt.Println("8. Read *: ")
	// if stream, err := grpcClient.ReadAll(contextCRUD, &emptypb.Empty{}); err == nil {
	// 	for {
	// 		todo, err := stream.Recv()
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		fmt.Println(todo.Task, todo.Completed)
	// 	}
	// } else {
	// 	panic(err)
	// }
	// fmt.Println("Done")
}
