package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/darrior/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_handlers_authMiddlware(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		GetPasswordHash(gomock.Any(), "7ddf6f86-3a47-40de-a7b5-13cb9d09f431").
		Return("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rawUUID := req.Context().Value(userUUIDKey)
		if rawUUID == nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		uuid, ok := rawUUID.(string)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		data := []byte(uuid)

		w.Header().Add("content-type", "text/plain")
		w.Header().Add("content-length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(data)
	})

	type fields struct {
		s Service
	}
	type args struct {
		handler http.Handler
		req     *http.Request
		w       *httptest.ResponseRecorder
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
				handler: h,
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
					tokenString, _ := signClaims("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", "7ddf6f86-3a47-40de-a7b5-13cb9d09f431")

					cookie := &http.Cookie{
						Name:  authCookieName,
						Value: tokenString,
					}

					req.AddCookie(cookie)

					return req
				}(),
				w: &httptest.ResponseRecorder{},
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "No cookie",
			fields: fields{
				s: s,
			},
			args: args{
				handler: h,
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

					return req
				}(),
				w: &httptest.ResponseRecorder{},
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{

			name: "Invalid cookie",
			fields: fields{
				s: s,
			},
			args: args{
				handler: h,
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

					cookie := &http.Cookie{
						Name:  authCookieName,
						Value: "KEHktXK5WpCWQ4yD8e4zLt8j9DqMbsHc2RVgTn51L88FKteHhFuhyRnTeEtn1SWkb1r7hGTj0mogGOk9vK79oGJvhQPSkbBhLnGSvjUESOB9ANRRb8uGhnWA9JBYwEpG",
					}

					req.AddCookie(cookie)

					return req
				}(),
				w: &httptest.ResponseRecorder{},
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}

			gotHandler := h.authMiddlware(tt.args.handler)

			gotHandler.ServeHTTP(tt.args.w, tt.args.req)

			assert.Equal(t, tt.want.code, tt.args.w.Code)
		})
	}
}
