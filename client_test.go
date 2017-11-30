package notbearclient

import (
	"context"
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errChan := make(chan error)
	client := NewClient(3, 10, ctx, errChan)
	go client.Run()
	// req, err := NewRequest("POST", "http://www.motorcycle.com/specs/", "", "", map[string][]string{
	// 	"MakeId":      []string{"20"},
	// 	"ModelType":   []string{"Sport"},
	// 	"year":        []string{"2016"},
	// 	"TrimId":      []string{"183936"},
	// 	"get_specs.x": []string{"113"},
	// 	"get_specs.y": []string{"12"},
	// })
	req, err := NewRequest("GET", "https://www.autoevolution.com/moto/", "", "", map[string][]string{})
	if err != nil {
		fmt.Println(err)
		return
	}
	client.Input <- req
	// go func() {
	// 	err := <-client.Error
	// 	fmt.Println(err)
	// }()
	go func() {
		result := <-client.Output
		fmt.Println(result)
	}()
	close(client.Input)
	_ = <-client.Done
}
