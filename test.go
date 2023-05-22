package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func test(t *testing.T) {
	main()
	app := App

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}

	fmt.Println(resp)

	/*if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	expectedBody := "Hello, World!"
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}

	if string(body) != expectedBody {
		t.Errorf("Expected response body %q, got %q", expectedBody, string(body))
	}

	err := app.Listen(os.Getenv("PUERTO"))
	if err != nil {
		fmt.Println(err)
	}*/
}
