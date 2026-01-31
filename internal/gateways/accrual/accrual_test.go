// Package accrual provides methods to interact with accural system.
package accrual

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darrior/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_accrual_GetOrder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/orders/{number}", func(w http.ResponseWriter, req *http.Request) {
		number := req.PathValue("number")
		switch number {
		case "1111":
			res := models.AccrualOrderState{
				Order:   "1111",
				Status:  models.AccrualOrderStatusProcessed,
				Accrual: 123.0,
			}
			data, _ := json.Marshal(res)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		case "1234":
			http.Error(w, "", http.StatusTooManyRequests)
		case "12345":
			http.Error(w, "", http.StatusNoContent)
		case "4321ad":
			http.Error(w, "", http.StatusInternalServerError)
		}
	})
	srv := httptest.NewServer(mux)

	type fields struct {
		client  *http.Client
		baseURL string
	}
	type args struct {
		number string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      models.AccrualOrderState
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			fields: fields{
				client:  &http.Client{},
				baseURL: srv.URL,
			},
			args: args{
				number: "1111",
			},
			want: models.AccrualOrderState{
				Order:   "1111",
				Status:  models.AccrualOrderStatusProcessed,
				Accrual: 123.0,
			},
			assertion: assert.NoError,
		},
		{
			name: "Too many requests",
			fields: fields{
				client:  &http.Client{},
				baseURL: srv.URL,
			},
			args: args{
				number: "1234",
			},
			want:      models.AccrualOrderState{},
			assertion: assert.Error,
		},
		{
			name: "No content",
			fields: fields{
				client:  &http.Client{},
				baseURL: srv.URL,
			},
			args: args{
				number: "12345",
			},
			want:      models.AccrualOrderState{},
			assertion: assert.Error,
		},
		{
			name: "Internal server error",
			fields: fields{
				client:  &http.Client{},
				baseURL: srv.URL,
			},
			args: args{
				number: "4321ad",
			},
			want:      models.AccrualOrderState{},
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &accrual{
				client:  tt.fields.client,
				baseURL: tt.fields.baseURL,
			}
			got, err := a.GetOrder(tt.args.number)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
