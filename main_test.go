package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockedSnClient struct {
	mock.Mock
}

func (mock *MockedSnClient) CreateIncident(incident Incident) (*Incident, error) {
	args := mock.Called(incident)
	return args.Get(0).(*Incident), args.Error(1)
}

func (mock *MockedSnClient) GetIncidents(params map[string]string) ([]Incident, error) {
	args := mock.Called(params)
	return args.Get(0).([]Incident), args.Error(1)
}

func (mock *MockedSnClient) UpdateIncident(incident Incident) (*Incident, error) {
	args := mock.Called(incident)
	return args.Get(0).(*Incident), args.Error(1)
}

func TestWebhookHandler_OK(t *testing.T) {
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(&basicIncident, nil)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_body.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_BadRequest(t *testing.T) {
	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	expected := `{"Status":400,"Message":"EOF"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_InternalServerError(t *testing.T) {
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(&Incident{}, errors.New("Error"))

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_body.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expected := `{"Status":500,"Message":"Error"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}
