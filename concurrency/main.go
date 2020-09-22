package main

import (
	"bufio"
	"fmt"
	"microGo/repo"
	"microGo/services"
	_ "microGo/src/api/domain/repositories"
	"microGo/src/api/utils/errors"
	"os"
	"sync"
)

var (
	success = make(map[string]string, 0)
	failed  = make(map[string]errors.ApiError, 0)
)

type createRepoResult struct {
	Request repo.CreateRepoRequest
	Result  *repo.CreateRepoResponse
	Error   errors.ApiError
}

func getRequests() []repo.CreateRepoRequest {
	result := make([]repo.CreateRepoRequest, 0)

	file, err := os.Open("/Users/fleon/Desktop/requests/requests.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		request := repo.CreateRepoRequest{
			Name: line,
		}
		result = append(result, request)
	}
	return result
}

func main() {
	requests := getRequests()

	fmt.Println(fmt.Sprintf("about to process %d requests", len(requests)))

	input := make(chan createRepoResult)
	buffer := make(chan bool, 10)
	var wg sync.WaitGroup

	go handleResults(&wg, input)

	for _, request := range requests {
		buffer <- true
		wg.Add(1)
		go createRepo(buffer, input, request)
	}

	wg.Wait()
	close(input)

	// Now you can write success and failed maps to disk or notify them via email or anything you need to do.
}

func handleResults(wg *sync.WaitGroup, input chan createRepoResult) {
	for result := range input {
		if result.Error != nil {
			failed[result.Request.Name] = result.Error
		} else {
			success[result.Request.Name] = result.Result.Name
		}
		wg.Done()
	}
}

func createRepo(buffer chan bool, output chan createRepoResult, request repo.CreateRepoRequest) {
	result, err := services.RepositoryService.CreateRepo("your_client_id", request)

	output <- createRepoResult{
		Request: request,
		Result:  result,
		Error:   err,
	}
	<-buffer
}
