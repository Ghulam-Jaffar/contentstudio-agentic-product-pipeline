package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	apimodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Constants for test strings
// ============================================================================

const (
	contentTypeJSON              = "application/json"
	competitorWorkEndpoint       = "/competitor-work"
	kafkaSendFailedMsg           = "Kafka send failed"
	competitorWorkOrderInstagram = "competitor-work-order-instagram"
	competitorWorkOrderFacebook  = "competitor-work-order-facebook"
	competitorWorkOrderLinkedIn  = "competitor-work-order-linkedin"
)

// ============================================================================
// Helper Functions
// ============================================================================

// makeCompetitorRequestBody creates a JSON request body from CompetitorWorkRequest
func makeCompetitorRequestBody(t *testing.T, req apimodels.CompetitorWorkRequest) *bytes.Buffer {
	t.Helper()
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	return bytes.NewBuffer(body)
}

// ============================================================================
// Test: HandleCompetitorWork
// ============================================================================

func TestHandleCompetitorWork(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           *bytes.Buffer
		setup          func(prod *MockProducer)
		expectedStatus int
		expectedBody   string // substring match
	}{
		{
			name:   "Invalid method - GET",
			method: http.MethodGet,
			body:   bytes.NewBuffer(nil),
			setup: func(*MockProducer) {
				// No setup needed - HTTP method validation occurs before any producer interaction
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid method - PUT",
			method: http.MethodPut,
			body:   bytes.NewBuffer(nil),
			setup: func(*MockProducer) {
				// No setup needed - HTTP method validation occurs before any producer interaction
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid method - DELETE",
			method: http.MethodDelete,
			body:   bytes.NewBuffer(nil),
			setup: func(*MockProducer) {
				// No setup needed - HTTP method validation occurs before any producer interaction
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid method - PATCH",
			method: http.MethodPatch,
			body:   bytes.NewBuffer(nil),
			setup: func(*MockProducer) {
				// No setup needed - HTTP method validation occurs before any producer interaction
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid JSON body",
			method: http.MethodPost,
			body:   bytes.NewBufferString("{invalid json"),
			setup: func(*MockProducer) {
				// No setup needed - JSON parsing occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"Invalid request body"`,
		},
		{
			name:   "Empty JSON body",
			method: http.MethodPost,
			body:   bytes.NewBufferString("{}"),
			setup: func(*MockProducer) {
				// No setup needed - validation occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"MISSING_FIELD"`,
		},
		{
			name:   "Missing page_id",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				Channel:  "facebook",
			}),
			setup: func(*MockProducer) {
				// No setup needed - validation occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"page_id is required"`,
		},
		{
			name:   "Missing channel",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				PageID:   "page_456",
			}),
			setup: func(*MockProducer) {
				// No setup needed - validation occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"channel is required"`,
		},
		{
			name:   "Empty page_id",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				PageID:   "",
				Channel:  "facebook",
			}),
			setup: func(*MockProducer) {
				// No setup needed - validation occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"page_id is required"`,
		},
		{
			name:   "Empty channel",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				PageID:   "page_456",
				Channel:  "",
			}),
			setup: func(*MockProducer) {
				// No setup needed - validation occurs before Kafka operation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"channel is required"`,
		},
		{
			name:   "Successful Facebook request",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				PageID:   "fb_page_456",
				Channel:  "facebook",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("fb_page_456"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful Instagram request",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_789",
				PageID:   "insta_page_123",
				Channel:  "instagram",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte("insta_page_123"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful request without report_id",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				PageID:  "page_999",
				Channel: "facebook",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_999"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Kafka producer failure - Facebook",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_abc",
				PageID:   "page_def",
				Channel:  "facebook",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_def"),
					mock.Anything,
				).Return(errors.New("kafka connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   kafkaSendFailedMsg,
		},
		{
			name:   "Kafka producer failure - Instagram",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_xyz",
				PageID:   "insta_789",
				Channel:  "instagram",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte("insta_789"),
					mock.Anything,
				).Return(errors.New("kafka timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   kafkaSendFailedMsg,
		},
		{
			name:   "Verify response contains all fields - Facebook",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_full",
				PageID:   "page_full",
				Channel:  "facebook",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_full"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Special characters in page_id",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: "report_special",
				PageID:   "page-with_special.chars123",
				Channel:  "facebook",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page-with_special.chars123"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Long page_id and report_id",
			method: http.MethodPost,
			body: makeCompetitorRequestBody(t, apimodels.CompetitorWorkRequest{
				ReportID: strings.Repeat("a", 100),
				PageID:   strings.Repeat("b", 100),
				Channel:  "instagram",
			}),
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte(strings.Repeat("b", 100)),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			s := makeAPIServer(repo, prod, nil)

			if tt.setup != nil {
				tt.setup(prod)
			}

			req := httptest.NewRequest(tt.method, competitorWorkEndpoint, tt.body)
			w := httptest.NewRecorder()
			s.HandleCompetitorWork(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			responseBody := buf.String()
			assert.Contains(t, responseBody, tt.expectedBody)

			// Verify response structure for successful requests
			if tt.expectedStatus == http.StatusOK {
				var jsonResp map[string]interface{}
				err := json.Unmarshal([]byte(responseBody), &jsonResp)
				assert.NoError(t, err, "response should be valid JSON")
				assert.Equal(t, "success", jsonResp["status"])
				assert.NotEmpty(t, jsonResp["timestamp"])

				// Verify timestamp is RFC3339
				if timestamp, ok := jsonResp["timestamp"].(string); ok {
					_, err := time.Parse(time.RFC3339, timestamp)
					assert.NoError(t, err, "timestamp should be RFC3339 format")
				}
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: processCompetitorWork (internal function)
// ============================================================================

func TestProcessCompetitorWork(t *testing.T) {
	tests := []struct {
		name        string
		request     apimodels.CompetitorWorkRequest
		setup       func(prod *MockProducer)
		expectedErr string
	}{
		{
			name: "Success - Facebook with all fields",
			request: apimodels.CompetitorWorkRequest{
				ReportID:  "report_123",
				PageID:    "fb_page_456",
				Channel:   "facebook",
				StartDate: "2025-01-01",
				EndDate:   "2025-01-31",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("fb_page_456"),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.CompetitorWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ReportID == "report_123" &&
							workOrder.PageID == "fb_page_456" &&
							workOrder.Channel == "facebook" &&
							workOrder.Mode == "full_refresh" &&
							workOrder.StartDate == "2025-01-01" &&
							workOrder.EndDate == "2025-01-31"
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - Instagram with all fields",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_789",
				PageID:   "insta_page_123",
				Channel:  "instagram",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte("insta_page_123"),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.CompetitorWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ReportID == "report_789" &&
							workOrder.PageID == "insta_page_123" &&
							workOrder.Channel == "instagram" &&
							workOrder.Mode == "full_refresh"
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - without report_id",
			request: apimodels.CompetitorWorkRequest{
				PageID:  "page_no_report",
				Channel: "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_no_report"),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.CompetitorWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ReportID == "" &&
							workOrder.PageID == "page_no_report" &&
							workOrder.Channel == "facebook" &&
							workOrder.Mode == "full_refresh"
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Kafka error - Facebook",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_error",
				PageID:   "page_error",
				Channel:  "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_error"),
					mock.Anything,
				).Return(errors.New("kafka broker not available"))
			},
			expectedErr: kafkaSendFailedMsg,
		},
		{
			name: "Kafka error - Instagram",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_timeout",
				PageID:   "page_timeout",
				Channel:  "instagram",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte("page_timeout"),
					mock.Anything,
				).Return(errors.New("request timeout"))
			},
			expectedErr: kafkaSendFailedMsg,
		},
		{
			name: "Verify topic naming - Facebook",
			request: apimodels.CompetitorWorkRequest{
				PageID:  "page_topic_test",
				Channel: "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook, // Topic must match exactly
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Verify topic naming - Instagram",
			request: apimodels.CompetitorWorkRequest{
				PageID:  "page_topic_test2",
				Channel: "instagram",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram, // Topic must match exactly
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Verify key is page_id",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_key_test",
				PageID:   "unique_page_id_12345",
				Channel:  "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("unique_page_id_12345"), // Key must be page_id
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Multiple channels supported - linkedin (edge case)",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_linkedin",
				PageID:   "linkedin_page",
				Channel:  "linkedin",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderLinkedIn,
					[]byte("linkedin_page"),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Special characters in IDs",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_with-special_chars.123",
				PageID:   "page_with-special_chars.456",
				Channel:  "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_with-special_chars.456"),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Very long IDs",
			request: apimodels.CompetitorWorkRequest{
				ReportID: strings.Repeat("r", 200),
				PageID:   strings.Repeat("p", 200),
				Channel:  "instagram",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderInstagram,
					[]byte(strings.Repeat("p", 200)),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Empty report_id is valid",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "",
				PageID:   "page_empty_report",
				Channel:  "facebook",
			},
			setup: func(prod *MockProducer) {
				prod.On("Produce",
					mock.Anything,
					competitorWorkOrderFacebook,
					[]byte("page_empty_report"),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.CompetitorWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ReportID == "" &&
							workOrder.Mode == "full_refresh"
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			runProcessCompetitorWorkTest(t, tt)
		})
	}
}

// runProcessCompetitorWorkTest executes a single test case for ProcessCompetitorWork.
// This helper reduces cognitive complexity by extracting the test execution logic.
func runProcessCompetitorWorkTest(t *testing.T, tt struct {
	name        string
	request     apimodels.CompetitorWorkRequest
	setup       func(prod *MockProducer)
	expectedErr string
}) {
	server, producer := setupProcessCompetitorWorkTest(t, tt.setup)
	executeProcessCompetitorWorkRequest(t, server, producer, tt.request, tt.expectedErr)
}

// executeProcessCompetitorWorkRequest handles HTTP request/response for ProcessCompetitorWork testing.
// Extracted to reduce cognitive complexity in the main test function.
func executeProcessCompetitorWorkRequest(t *testing.T, server *api.APIServer, producer *MockProducer, request apimodels.CompetitorWorkRequest, expectedErr string) {
	req := makeCompetitorRequestBody(t, request)
	httpReq := httptest.NewRequest(http.MethodPost, competitorWorkEndpoint, req)
	w := httptest.NewRecorder()
	server.HandleCompetitorWork(w, httpReq)

	verifyProcessCompetitorWorkResponse(t, w, expectedErr)
	producer.AssertExpectations(t)
}

// setupProcessCompetitorWorkTest initializes test dependencies and returns the server and mock producer.
// Extracting this reduces cognitive complexity in the main test function.
func setupProcessCompetitorWorkTest(t *testing.T, setupFn func(*MockProducer)) (*api.APIServer, *MockProducer) {
	repo := mongodb.NewSeedRepo()
	prod := new(MockProducer)
	s := makeAPIServer(repo, prod, nil)

	if setupFn != nil {
		setupFn(prod)
	}

	return s, prod
}

// verifyProcessCompetitorWorkResponse checks the HTTP response for success or expected errors.
// This separation further reduces cognitive complexity in the main test loop.
func verifyProcessCompetitorWorkResponse(t *testing.T, w *httptest.ResponseRecorder, expectedErr string) {
	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if expectedErr == "" {
		verifySuccessResponse(t, resp, body)
	} else {
		verifyErrorResponse(t, resp, body, expectedErr)
	}
}

// verifySuccessResponse checks a successful (200 OK) response.
func verifySuccessResponse(t *testing.T, resp *http.Response, body []byte) {
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// verifyErrorResponse checks an error (500 InternalServerError) response.
func verifyErrorResponse(t *testing.T, resp *http.Response, body []byte, expectedErr string) {
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), expectedErr)
}

// ============================================================================
// Test: Work Order Structure Validation
// ============================================================================

func TestCompetitorWorkOrderStructure(t *testing.T) {
	tests := []struct {
		name    string
		request apimodels.CompetitorWorkRequest
		verify  func(t *testing.T, data []byte)
	}{
		{
			name: "Work order contains correct mode",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_mode",
				PageID:   "page_mode",
				Channel:  "facebook",
			},
			verify: func(t *testing.T, data []byte) {
				var workOrder kafkaModels.CompetitorWorkOrder
				err := json.Unmarshal(data, &workOrder)
				assert.NoError(t, err)
				assert.Equal(t, "full_refresh", workOrder.Mode)
			},
		},
		{
			name: "Work order marshals to valid JSON",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_json",
				PageID:   "page_json",
				Channel:  "instagram",
			},
			verify: func(t *testing.T, data []byte) {
				var workOrder kafkaModels.CompetitorWorkOrder
				err := json.Unmarshal(data, &workOrder)
				assert.NoError(t, err)
				assert.True(t, json.Valid(data))
			},
		},
		{
			name: "Work order preserves all request fields",
			request: apimodels.CompetitorWorkRequest{
				ReportID:  "preserve_report",
				PageID:    "preserve_page",
				Channel:   "facebook",
				StartDate: "2025-02-01",
				EndDate:   "2025-02-28",
			},
			verify: func(t *testing.T, data []byte) {
				var workOrder kafkaModels.CompetitorWorkOrder
				err := json.Unmarshal(data, &workOrder)
				assert.NoError(t, err)
				assert.Equal(t, "preserve_report", workOrder.ReportID)
				assert.Equal(t, "preserve_page", workOrder.PageID)
				assert.Equal(t, "facebook", workOrder.Channel)
				assert.Equal(t, "2025-02-01", workOrder.StartDate)
				assert.Equal(t, "2025-02-28", workOrder.EndDate)
			},
		},
		{
			name: "Work order handles empty report_id",
			request: apimodels.CompetitorWorkRequest{
				PageID:  "page_no_report",
				Channel: "instagram",
			},
			verify: func(t *testing.T, data []byte) {
				var workOrder kafkaModels.CompetitorWorkOrder
				err := json.Unmarshal(data, &workOrder)
				assert.NoError(t, err)
				assert.Equal(t, "", workOrder.ReportID)
				assert.Equal(t, "page_no_report", workOrder.PageID)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			var capturedData []byte
			prod.On("Produce",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Run(func(args mock.Arguments) {
				capturedData = args.Get(3).([]byte)
			}).Return(nil)

			s := makeAPIServer(repo, prod, nil)

			req := makeCompetitorRequestBody(t, tt.request)
			httpReq := httptest.NewRequest(http.MethodPost, competitorWorkEndpoint, req)
			w := httptest.NewRecorder()
			s.HandleCompetitorWork(w, httpReq)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.NotNil(t, capturedData)

			if tt.verify != nil {
				tt.verify(t, capturedData)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: Edge Cases and Error Handling
// ============================================================================

func TestCompetitorWorkEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		contentType    string
		setupMock      bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Nil request body",
			method:         http.MethodPost,
			body:           "",
			contentType:    contentTypeJSON,
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"Invalid request body"`,
		},
		{
			name:           "JSON array instead of object",
			method:         http.MethodPost,
			body:           `[{"page_id":"test","channel":"facebook"}]`,
			contentType:    contentTypeJSON,
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"Invalid request body"`,
		},
		{
			name:           "JSON with extra fields",
			method:         http.MethodPost,
			body:           `{"report_id":"r1","page_id":"p1","channel":"facebook","extra":"field"}`,
			contentType:    contentTypeJSON,
			setupMock:      true,
			expectedStatus: http.StatusOK, // Extra fields should be ignored
			expectedBody:   `"status":"success"`,
		},
		{
			name:           "Unicode characters in IDs",
			method:         http.MethodPost,
			body:           `{"report_id":"报告","page_id":"页面","channel":"facebook"}`,
			contentType:    contentTypeJSON,
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setupMock {
				prod.On("Produce",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			}

			s := makeAPIServer(repo, prod, nil)

			req := httptest.NewRequest(tt.method, competitorWorkEndpoint, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			s.HandleCompetitorWork(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), tt.expectedBody)

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: Response Format Validation
// ============================================================================

func TestCompetitorWorkResponseFormat(t *testing.T) {
	tests := []struct {
		name    string
		request apimodels.CompetitorWorkRequest
		verify  func(t *testing.T, resp *http.Response, body string)
	}{
		{
			name: "Response contains all required fields",
			request: apimodels.CompetitorWorkRequest{
				ReportID: "report_123",
				PageID:   "page_456",
				Channel:  "facebook",
			},
			verify: func(t *testing.T, resp *http.Response, body string) {
				var jsonResp map[string]interface{}
				err := json.Unmarshal([]byte(body), &jsonResp)
				assert.NoError(t, err)

				assert.Equal(t, "success", jsonResp["status"])
				assert.Equal(t, "report_123", jsonResp["report_id"])
				assert.Equal(t, "page_456", jsonResp["page_id"])
				assert.Equal(t, "facebook", jsonResp["channel"])
				assert.NotEmpty(t, jsonResp["timestamp"])
			},
		},
		{
			name: "Response timestamp is valid RFC3339",
			request: apimodels.CompetitorWorkRequest{
				PageID:  "page_time",
				Channel: "instagram",
			},
			verify: func(t *testing.T, resp *http.Response, body string) {
				var jsonResp map[string]interface{}
				err := json.Unmarshal([]byte(body), &jsonResp)
				assert.NoError(t, err)

				timestamp, ok := jsonResp["timestamp"].(string)
				assert.True(t, ok)

				parsedTime, err := time.Parse(time.RFC3339, timestamp)
				assert.NoError(t, err)
				assert.True(t, time.Since(parsedTime) < 5*time.Second, "timestamp should be recent")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			prod.On("Produce",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil)

			s := makeAPIServer(repo, prod, nil)

			req := makeCompetitorRequestBody(t, tt.request)
			httpReq := httptest.NewRequest(http.MethodPost, competitorWorkEndpoint, req)
			w := httptest.NewRecorder()
			s.HandleCompetitorWork(w, httpReq)

			resp := w.Result()
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			if tt.verify != nil {
				tt.verify(t, resp, string(body))
			}

			prod.AssertExpectations(t)
		})
	}
}
