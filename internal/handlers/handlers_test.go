// Package handlers provides handlers for server's routes
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/darrior/gophermart/internal/mocks"
	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_handlers_postAPIUserRegister(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		RegisterUser(gomock.Any(), "test", "123").
		Return(models.User{
			UUID:             "db3e6683-6ed8-4f1d-86c3-a77234dea94c",
			Login:            "test",
			PasswordHash:     "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			CurrentBalance:   0,
			WithdrawnBalance: 0,
		}, nil)

	s.EXPECT().
		RegisterUser(gomock.Any(), "test1", "123").
		Return(models.User{}, service.ErrLoginExists)

	s.EXPECT().
		RegisterUser(gomock.Any(), "test2", "123").
		Return(models.User{}, errors.New(""))

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code       int
		cookiesNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:       http.StatusOK,
				cookiesNum: 1,
			},
		},
		{
			name: "Existing user",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test1",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:       http.StatusConflict,
				cookiesNum: 0,
			},
		},
		{
			name: "Internal server error",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test2",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:       http.StatusInternalServerError,
				cookiesNum: 0,
			},
		},
		{
			name: "Invalid content-type",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bytes.NewReader(data))
					req.Header.Add("content-type", "text/plain")

					return req
				}(),
			},
			want: want{
				code:       http.StatusBadRequest,
				cookiesNum: 0,
			},
		},
		{
			name: "Random data",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := []byte("123414")
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:       http.StatusBadRequest,
				cookiesNum: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.postAPIUserRegister(tt.args.w, tt.args.req)

			response := tt.args.w.Result()
			defer func() { _ = response.Body.Close() }()

			assert.Equal(t, tt.want.code, tt.args.w.Code)
			assert.Len(t, response.Cookies(), tt.want.cookiesNum)
		})
	}
}

func Test_handlers_postAPIUserLogin(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		LoginUser(gomock.Any(), "test", "123").
		Return(models.User{
			UUID:             "db3e6683-6ed8-4f1d-86c3-a77234dea94c",
			Login:            "test",
			PasswordHash:     "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			CurrentBalance:   0,
			WithdrawnBalance: 0,
		}, nil)

	s.EXPECT().
		LoginUser(gomock.Any(), "test1", "123").
		Return(models.User{}, service.ErrInvalidCredentials)

	s.EXPECT().
		LoginUser(gomock.Any(), "test2", "123").
		Return(models.User{}, errors.New(""))

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code      int
		cookieNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/login", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:      http.StatusOK,
				cookieNum: 1,
			},
		},
		{
			name: "Invalid password",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test1",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/login", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:      http.StatusUnauthorized,
				cookieNum: 0,
			},
		},
		{
			name: "Internal server error",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test2",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/login", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: want{
				code:      http.StatusInternalServerError,
				cookieNum: 0,
			},
		},
		{
			name: "Bad request",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					a := models.AuthenticationDataRequest{
						Login:    "test3",
						Password: "123",
					}
					data, _ := json.Marshal(a)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/login", bytes.NewReader(data))
					req.Header.Add("content-type", "text/plain")

					return req
				}(),
			},
			want: want{
				code:      http.StatusBadRequest,
				cookieNum: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.postAPIUserLogin(tt.args.w, tt.args.req)

			response := tt.args.w.Result()
			defer func() { _ = response.Body.Close() }()

			assert.Equal(t, tt.want.code, tt.args.w.Code)
			assert.Len(t, response.Cookies(), tt.want.cookieNum)
		})
	}
}

func Test_handlers_postAPIUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		AddOrder(gomock.Any(), "a93683be-24da-4e0b-9890-3a20346b3538", "79927398713").
		Return(nil)

	s.EXPECT().
		AddOrder(gomock.Any(), "4fc6cf83-5c12-41f8-baf2-f4464cae458b", "79927398713").
		Return(service.ErrOrderAlreadyExists)

	s.EXPECT().
		AddOrder(gomock.Any(), "b227ff55-b9ee-4924-ae48-b995b077fe20", "79927398713").
		Return(service.ErrOrderOwnedByOtherUser)

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "79927398713"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, "a93683be-24da-4e0b-9890-3a20346b3538"))
				}(),
			},
			want: http.StatusAccepted,
		},
		{
			name: "Order already exists",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "79927398713"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, "4fc6cf83-5c12-41f8-baf2-f4464cae458b"))
				}(),
			},
			want: http.StatusOK,
		},
		{
			name: "Order uploaded by other user",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "79927398713"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, "b227ff55-b9ee-4924-ae48-b995b077fe20"))
				}(),
			},
			want: http.StatusConflict,
		},
		{
			name: "UUID is not string",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "79927398713"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, 123))
				}(),
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "Number not valid",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "79827398713"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, "c771847e-b3ca-4d49-8f16-800c35cedd8d"))
				}(),
			},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "Not a number",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "abcde"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req.WithContext(context.WithValue(req.Context(), userUUIDKey, "c771847e-b3ca-4d49-8f16-800c35cedd8d"))
				}(),
			},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "No UUID",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "abcde"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "text/plain")

					return req
				}(),
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "Invalid content type",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := "abcde"

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/orders", strings.NewReader(data))

					req.Header.Add("content-type", "application/json")

					return req
				}(),
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.postAPIUserOrders(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want, tt.args.w.Code)
		})
	}
}

func Test_handlers_getAPIUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		ListOrders(gomock.Any(), "72bd90b0-6649-43d4-ab4a-a43b9c37783a").
		Return([]models.OrderResponse{
			{
				Number:     "1111",
				Status:     models.OrderStatusInvalid,
				Accrual:    0,
				UploadedAt: "123",
			},
			{
				Number:     "1112",
				Status:     models.OrderStatusProcessed,
				Accrual:    12.5,
				UploadedAt: "123",
			},
		}, nil)

	s.EXPECT().
		ListOrders(gomock.Any(), "a02f089c-c3a6-48d8-9f95-988d66132420").
		Return([]models.OrderResponse{}, nil)

	s.EXPECT().
		ListOrders(gomock.Any(), "0a767833-356c-4b9a-ae4b-c44483e29ad9").
		Return([]models.OrderResponse{}, errors.New(""))

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code int
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/orders", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "72bd90b0-6649-43d4-ab4a-a43b9c37783a"))

					return req
				}(),
			},
			want: want{
				code: http.StatusOK,
				data: func() []byte {
					res := []models.OrderResponse{
						{
							Number:     "1111",
							Status:     models.OrderStatusInvalid,
							Accrual:    0,
							UploadedAt: "123",
						},
						{
							Number:     "1112",
							Status:     models.OrderStatusProcessed,
							Accrual:    12.5,
							UploadedAt: "123",
						},
					}
					data, _ := json.Marshal(res)
					return data
				}(),
			},
		},
		{
			name: "No content",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/orders", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "a02f089c-c3a6-48d8-9f95-988d66132420"))

					return req
				}(),
			},
			want: want{
				code: http.StatusNoContent,
				data: nil,
			},
		},
		{
			name: "Internal server error",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/orders", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "0a767833-356c-4b9a-ae4b-c44483e29ad9"))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
		{
			name: "No UUID",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/orders", nil)

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
		{
			name: "UUID is not string",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/orders", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, 1234))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.getAPIUserOrders(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want.code, tt.args.w.Code)

			assert.Equal(t, tt.want.data, tt.args.w.Body.Bytes())
		})
	}
}

func Test_handlers_getAPIUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		GetBalance(gomock.Any(), "8df540a7-ee66-4d14-b440-4b4d5fe636ef").
		Return(models.BalanceResponse{
			Current:   12.3,
			Withdrawn: 32.1,
		}, nil)

	s.EXPECT().
		GetBalance(gomock.Any(), "3ba50d77-7b61-409e-ba23-fad249ed7c68").
		Return(models.BalanceResponse{}, errors.New(""))

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code int
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/balance", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "8df540a7-ee66-4d14-b440-4b4d5fe636ef"))

					return req
				}(),
			},
			want: want{
				code: http.StatusOK,
				data: func() []byte {
					balance := models.BalanceResponse{
						Current:   12.3,
						Withdrawn: 32.1,
					}

					data, _ := json.Marshal(balance)
					return data
				}(),
			},
		},
		{
			name: "Cannot get balance",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/balance", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "3ba50d77-7b61-409e-ba23-fad249ed7c68"))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
		{
			name: "Login not string",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/balance", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, 3231))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
		{
			name: "No login",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/balance", nil)

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.getAPIUserBalance(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want.code, tt.args.w.Code)
			assert.Equal(t, tt.want.data, tt.args.w.Body.Bytes())
		})
	}
}

func Test_handlers_postAPIUserBalanceWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		Withdraw(gomock.Any(), "847d2409-33d1-4513-8bef-6377b67ab92a", "61234", 12.3).
		Return(nil)

	s.EXPECT().
		Withdraw(gomock.Any(), "6ecfaa14-4a7e-4717-8d6d-688b57c02e2c", "61234", 12.3).
		Return(service.ErrInsufficientFunds)

	s.EXPECT().
		Withdraw(gomock.Any(), "9bafa41d-c2d0-4b51-8508-548223a386e4", "61234", 12.3).
		Return(errors.New(""))

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data, _ := json.Marshal(models.WithdrawRequest{
						Order: "61234",
						Sum:   12.3,
					})

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")
					req.Header.Add("content-length", strconv.Itoa(len(data)))
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "847d2409-33d1-4513-8bef-6377b67ab92a"))

					return req
				}(),
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Wrong content type",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data, _ := json.Marshal(models.WithdrawRequest{
						Order: "61234",
						Sum:   12.3,
					})

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "text/plain")
					req.Header.Add("content-length", strconv.Itoa(len(data)))
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "847d2409-33d1-4513-8bef-6377b67ab92a"))

					return req
				}(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong data",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data := []byte(`{ "test"`)
					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")
					req.Header.Add("content-length", strconv.Itoa(len(data)))
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "847d2409-33d1-4513-8bef-6377b67ab92a"))

					return req
				}(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "No UUID",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data, _ := json.Marshal(models.WithdrawRequest{
						Order: "61234",
						Sum:   12.3,
					})

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")
					req.Header.Add("content-length", strconv.Itoa(len(data)))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data, _ := json.Marshal(models.WithdrawRequest{
						Order: "61234",
						Sum:   12.3,
					})

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")
					req.Header.Add("content-length", strconv.Itoa(len(data)))
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "6ecfaa14-4a7e-4717-8d6d-688b57c02e2c"))

					return req
				}(),
			},
			want: want{
				code: http.StatusPaymentRequired,
			},
		},
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					data, _ := json.Marshal(models.WithdrawRequest{
						Order: "61234",
						Sum:   12.3,
					})

					req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/balance/withdraw", bytes.NewReader(data))
					req.Header.Add("content-type", "application/json")
					req.Header.Add("content-length", strconv.Itoa(len(data)))
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "9bafa41d-c2d0-4b51-8508-548223a386e4"))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.postAPIUserBalanceWithdraw(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want.code, tt.args.w.Code)
		})
	}
}

func Test_handlers_getAPIUserWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		ListWithdrawals(gomock.Any(), "fc4003b1-9477-48ab-a56e-a68a5059ae7c").
		Return([]models.WithdrawResponse{
			{
				Order:       "1234",
				Sum:         12.3,
				ProcessedAt: "now",
			},
		}, nil)

	s.EXPECT().
		ListWithdrawals(gomock.Any(), "2db6af16-a19b-4e9c-8dd1-407cc94f30ba").
		Return([]models.WithdrawResponse{}, nil)

	type fields struct {
		s Service
	}
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	type want struct {
		code int
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Valid test",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/withdrawals", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "fc4003b1-9477-48ab-a56e-a68a5059ae7c"))

					return req
				}(),
			},
			want: want{
				code: http.StatusOK,
				data: []byte(`[{"order":"1234","sum":12.3,"processed_at":"now"}]`),
			},
		},
		{
			name: "Valid test no content",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/withdrawals", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, "2db6af16-a19b-4e9c-8dd1-407cc94f30ba"))

					return req
				}(),
			},
			want: want{
				code: http.StatusNoContent,
				data: nil,
			},
		},
		{
			name: "No UUID",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/withdrawals", nil)

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
		{
			name: "Authentication error",
			fields: fields{
				s: s,
			},
			args: args{
				w: httptest.NewRecorder(),
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/api/user/withdrawals", nil)
					req = req.WithContext(context.WithValue(req.Context(), userUUIDKey, true))

					return req
				}(),
			},
			want: want{
				code: http.StatusInternalServerError,
				data: []byte(http.StatusText(http.StatusInternalServerError) + "\n"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			h.getAPIUserWithdrawals(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want.code, tt.args.w.Code)
			assert.Equal(t, tt.want.data, tt.args.w.Body.Bytes())
		})
	}
}
