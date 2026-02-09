package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// FixedProductID is used for all operations (GET, UPDATE, DELETE by ID)
	// We don't care about 404/409 errors - we just want to test the API
	FixedProductID = 1
)

// executeRequest performs a single HTTP request and returns latency and error
func executeRequest(ctx *RequestContext, task RequestTask) (time.Duration, error) {
	start := time.Now()

	var err error
	var statusCode int
	var responseBody string

	// Use ProductID from context, fallback to FixedProductID if not set
	productID := int(ctx.ProductID)
	if productID == 0 {
		productID = FixedProductID
	}

	switch task.Type {
	case GetProducts:
		statusCode, responseBody, err = getProducts(ctx)
	case CreateProduct:
		statusCode, responseBody, err = createProduct(ctx)
	case GetProductByID:
		statusCode, responseBody, err = getProductByID(ctx, productID)
	case UpdateProduct:
		statusCode, responseBody, err = updateProduct(ctx, productID)
	case DeleteProduct:
		statusCode, responseBody, err = deleteProduct(ctx, productID)
	default:
		err = fmt.Errorf("unknown benchmark type: %s", task.Type)
		statusCode = 0
		responseBody = ""
	}

	latency := time.Since(start)

	// Record error if any
	if err != nil {
		operation := fmt.Sprintf("%s %s", getHTTPMethod(task.Type), getEndpoint(task.Type, productID))
		ctx.ErrorStats.RecordError(operation, "request_error", err.Error(), statusCode, responseBody)
	}

	return latency, err
}

// getProducts performs GET /api/products
func getProducts(ctx *RequestContext) (int, string, error) {
	url := ctx.Config.URL + "/api/products"
	return doRequest(ctx.Client, "GET", url, nil)
}

// createProduct performs POST /api/products
func createProduct(ctx *RequestContext) (int, string, error) {
	url := ctx.Config.URL + "/api/products"
	product := Product{
		Name:        "Benchmark Product",
		Description: "Created by benchmark tool",
		Price:       99.99,
		Quantity:    100,
	}
	body, err := json.Marshal(product)
	if err != nil {
		return 0, "", err
	}
	return doRequest(ctx.Client, "POST", url, body)
}

// createProductAndGetID performs POST /api/products and extracts the ID from response
func createProductAndGetID(ctx *RequestContext) (int64, error) {
	url := ctx.Config.URL + "/api/products"
	product := Product{
		Name:        "Benchmark Product",
		Description: "Created by benchmark tool",
		Price:       99.99,
		Quantity:    100,
	}
	body, err := json.Marshal(product)
	if err != nil {
		return 0, err
	}

	statusCode, responseBody, err := doRequest(ctx.Client, "POST", url, body)
	if err != nil || statusCode != 201 {
		return 0, fmt.Errorf("failed to create product: status=%d, error=%v", statusCode, err)
	}

	var createdProduct Product
	if err := json.Unmarshal([]byte(responseBody), &createdProduct); err != nil {
		return 0, fmt.Errorf("failed to parse created product: %v", err)
	}

	return createdProduct.ID, nil
}

// getProductByID performs GET /api/products/{id}
func getProductByID(ctx *RequestContext, id int) (int, string, error) {
	url := fmt.Sprintf("%s/api/products/%d", ctx.Config.URL, id)
	return doRequest(ctx.Client, "GET", url, nil)
}

// updateProduct performs PUT /api/products/{id}
func updateProduct(ctx *RequestContext, id int) (int, string, error) {
	url := fmt.Sprintf("%s/api/products/%d", ctx.Config.URL, id)
	product := Product{
		ID:          int64(id),
		Name:        "Updated Product",
		Description: "Updated by benchmark tool",
		Price:       149.99,
		Quantity:    200,
	}
	body, err := json.Marshal(product)
	if err != nil {
		return 0, "", err
	}
	return doRequest(ctx.Client, "PUT", url, body)
}

// deleteProduct performs DELETE /api/products/{id}
func deleteProduct(ctx *RequestContext, id int) (int, string, error) {
	url := fmt.Sprintf("%s/api/products/%d", ctx.Config.URL, id)
	return doRequest(ctx.Client, "DELETE", url, nil)
}

// doRequest is a helper function to perform HTTP request
func doRequest(client *http.Client, method, url string, body []byte) (int, string, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, "", err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	// Consider 2xx and 4xx as "successful" request (we reached the API)
	// We only care about network/server errors
	if resp.StatusCode >= 500 {
		return resp.StatusCode, string(responseBody), fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return resp.StatusCode, string(responseBody), nil
}

// getHTTPMethod returns HTTP method for benchmark type
func getHTTPMethod(benchType BenchmarkType) string {
	switch benchType {
	case GetProducts, GetProductByID:
		return "GET"
	case CreateProduct:
		return "POST"
	case UpdateProduct:
		return "PUT"
	case DeleteProduct:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// getEndpoint returns endpoint for benchmark type
func getEndpoint(benchType BenchmarkType, productID int) string {
	switch benchType {
	case GetProducts:
		return "/api/products"
	case CreateProduct:
		return "/api/products"
	case GetProductByID:
		return fmt.Sprintf("/api/products/%d", productID)
	case UpdateProduct:
		return fmt.Sprintf("/api/products/%d", productID)
	case DeleteProduct:
		return fmt.Sprintf("/api/products/%d", productID)
	default:
		return "/unknown"
	}
}