// Package service implements buisness-logic of app.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/darrior/gophermart/internal/mocks"
	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/repository"
	"github.com/darrior/gophermart/internal/utils/atomic"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)
	a := mocks.NewMockAccrualSystem(ctrl)

	type args struct {
		ctx     context.Context
		r       Repository
		a       AccrualSystem
		workers int
	}
	tests := []struct {
		name string
		args args
		want *service
	}{
		{
			name: "",
			args: args{
				ctx:     t.Context(),
				r:       r,
				a:       a,
				workers: 5,
			},
			want: &service{
				repository:    r,
				accrualSystem: a,
				orderCfg: struct {
					orderWorkers int
					sleepUntil   atomic.GenericValue[time.Time]
				}{
					orderWorkers: 5,
					sleepUntil:   atomic.NewGenericValue(time.Now()),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(tt.args.ctx, tt.args.r, tt.args.a, tt.args.workers)

			assert.Equal(t, tt.want.accrualSystem, got.accrualSystem)
			assert.Equal(t, tt.want.repository, got.repository)
			assert.Equal(t, tt.want.orderCfg.orderWorkers, got.orderCfg.orderWorkers)
		})
	}
}

func Test_service_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test0").
		Return(models.User{}, repository.ErrUserNotFound)
	r.EXPECT().
		AddUser(gomock.Any(),
			gomock.Any(),
			"test0",
			"03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4").
		Return(nil)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test1").
		Return(models.User{}, nil)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test2").
		Return(models.User{}, errors.New(""))

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test3").
		Return(models.User{}, repository.ErrUserNotFound)
	r.EXPECT().
		AddUser(gomock.Any(), gomock.Any(), "test3", gomock.Any()).
		Return(errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
		workers       int
	}
	type args struct {
		ctx      context.Context
		login    string
		password string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      models.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test case",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test0",
				password: "1234",
			},
			want: models.User{
				UUID:             "",
				Login:            "test0",
				PasswordHash:     "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
				CurrentBalance:   0,
				WithdrawnBalance: 0,
			},
			assertion: assert.NoError,
		},
		{
			name: "User exists",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test1",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
		{
			name: "Cannot get user",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test2",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
		{
			name: "Cannot get user",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test3",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, tt.fields.workers)
			got, err := s.RegisterUser(tt.args.ctx, tt.args.login, tt.args.password)

			tt.assertion(t, err)

			assert.Equal(t, tt.want.Login, got.Login)
			assert.Equal(t, tt.want.PasswordHash, got.PasswordHash)
		})
	}
}

func Test_service_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test0").
		Return(models.User{
			UUID:             "1234-1234",
			Login:            "test0",
			PasswordHash:     "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
			CurrentBalance:   10.0,
			WithdrawnBalance: 1.0,
		}, nil)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test1").
		Return(models.User{

			UUID:             "1234-1234",
			Login:            "test1",
			PasswordHash:     "1231241212",
			CurrentBalance:   10.0,
			WithdrawnBalance: 1.0,
		}, nil)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test2").
		Return(models.User{}, repository.ErrUserNotFound)

	r.EXPECT().
		GetUserByLogin(gomock.Any(), "test3").
		Return(models.User{}, errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
		workers       int
	}
	type args struct {
		ctx      context.Context
		login    string
		password string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      models.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Correct test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test0",
				password: "1234",
			},
			want: models.User{
				UUID:             "1234-1234",
				Login:            "test0",
				PasswordHash:     "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
				CurrentBalance:   10.0,
				WithdrawnBalance: 1.0,
			},
			assertion: assert.NoError,
		},
		{
			name: "Invalid password",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test1",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
		{
			name: "Invalid login",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test2",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
		{
			name: "Can not get user",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:      context.TODO(),
				login:    "test3",
				password: "1234",
			},
			want:      models.User{},
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, tt.fields.workers)
			got, err := s.LoginUser(tt.args.ctx, tt.args.login, tt.args.password)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		AddOrder(gomock.Any(), "18e9249f-6e4c-4e2a-8af1-06333a960783", "1234", gomock.Any()).
		Return(nil)

	r.EXPECT().
		AddOrder(gomock.Any(), "4a632ab7-8911-49e8-a743-a583826c1ea9", "4321", gomock.Any()).
		Return(&repository.ErrorOrderExists{
			UUID: "4a632ab7-8911-49e8-a743-a583826c1ea9",
		})

	r.EXPECT().
		AddOrder(gomock.Any(), "8aa414a5-fd61-4e6b-85cc-d5bdd45fe1f2", "4321", gomock.Any()).
		Return(&repository.ErrorOrderExists{
			UUID: "4a632ab7-8911-49e8-a743-a583826c1ea9",
		})

	r.EXPECT().
		AddOrder(gomock.Any(), "aae4eda3-800d-457b-8763-08e019f93e9d", "1111", gomock.Any()).
		Return(errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
		workers       int
	}
	type args struct {
		ctx   context.Context
		uuid  string
		order string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Correct test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "18e9249f-6e4c-4e2a-8af1-06333a960783",
				order: "1234",
			},
			assertion: assert.NoError,
		},
		{
			name: "Duplicate order",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "4a632ab7-8911-49e8-a743-a583826c1ea9",
				order: "4321",
			},
			assertion: assert.Error,
		},
		{
			name: "Not owner of order",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "8aa414a5-fd61-4e6b-85cc-d5bdd45fe1f2",
				order: "4321",
			},
			assertion: assert.Error,
		},
		{
			name: "Cannot add order",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "aae4eda3-800d-457b-8763-08e019f93e9d",
				order: "1111",
			},
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, tt.fields.workers)
			tt.assertion(t, s.AddOrder(tt.args.ctx, tt.args.uuid, tt.args.order))
		})
	}
}

func Test_service_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		GetUser(gomock.Any(), "95002e63-f934-473e-918a-ce5cf307b440").
		Return(models.User{
			UUID:             "95002e63-f934-473e-918a-ce5cf307b440",
			Login:            "test0",
			PasswordHash:     "12345",
			CurrentBalance:   15.0,
			WithdrawnBalance: 0,
		}, nil)
	r.EXPECT().
		AddWithdrawal(gomock.Any(), "95002e63-f934-473e-918a-ce5cf307b440", "1111", 10.0, 5.0, gomock.Any()).
		Return(nil)

	r.EXPECT().
		GetUser(gomock.Any(), "66530a8b-b76d-486c-8364-297b20a295f5").
		Return(models.User{
			UUID:             "66530a8b-b76d-486c-8364-297b20a295f5",
			Login:            "test0",
			PasswordHash:     "12345",
			CurrentBalance:   15.0,
			WithdrawnBalance: 0,
		}, nil)
	r.EXPECT().
		AddWithdrawal(gomock.Any(), "66530a8b-b76d-486c-8364-297b20a295f5", "1111", 10.0, 5.0, gomock.Any()).
		Return(errors.New(""))

	r.EXPECT().
		GetUser(gomock.Any(), "558b32e7-5af9-4f5d-82a9-eadb3ee1c84e").
		Return(models.User{}, errors.New(""))

	r.EXPECT().
		GetUser(gomock.Any(), "8a10bcf1-0bf3-4cec-b80b-6f96d97d158c").
		Return(models.User{
			UUID:             "8a10bcf1-0bf3-4cec-b80b-6f96d97d158c",
			Login:            "test0",
			PasswordHash:     "12345",
			CurrentBalance:   15.0,
			WithdrawnBalance: 0,
		}, nil)

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
		workers       int
	}
	type args struct {
		ctx   context.Context
		uuid  string
		order string
		sum   float64
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Correct test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "95002e63-f934-473e-918a-ce5cf307b440",
				order: "1111",
				sum:   5.0,
			},
			assertion: assert.NoError,
		},
		{
			name: "Cannot add withdraw",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "66530a8b-b76d-486c-8364-297b20a295f5",
				order: "1111",
				sum:   5.0,
			},
			assertion: assert.Error,
		},
		{
			name: "Cannot add withdraw",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "558b32e7-5af9-4f5d-82a9-eadb3ee1c84e",
				order: "1111",
				sum:   5.0,
			},
			assertion: assert.Error,
		},
		{
			name: "Balance lower than sum",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:   context.TODO(),
				uuid:  "8a10bcf1-0bf3-4cec-b80b-6f96d97d158c",
				order: "1111",
				sum:   20.0,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, tt.fields.workers)
			tt.assertion(t, s.Withdraw(tt.args.ctx, tt.args.uuid, tt.args.order, tt.args.sum))
		})
	}
}

func Test_service_ListOrders(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		ListOrders(gomock.Any(), "82b6779d-557b-446c-a4e7-c67b62509119").
		Return([]models.Order{
			{
				UserUUID:   "82b6779d-557b-446c-a4e7-c67b62509119",
				Accrual:    10,
				Number:     "1111",
				Status:     models.OrderStatusProcessed,
				UploadedAt: time.Time{},
			},
			{
				UserUUID:   "82b6779d-557b-446c-a4e7-c67b62509119",
				Accrual:    0,
				Number:     "1112",
				Status:     models.OrderStatusInvalid,
				UploadedAt: time.Time{},
			},
			{
				UserUUID:   "82b6779d-557b-446c-a4e7-c67b62509119",
				Accrual:    0,
				Number:     "1113",
				Status:     models.OrderStatusNew,
				UploadedAt: time.Time{},
			},
			{
				UserUUID:   "82b6779d-557b-446c-a4e7-c67b62509119",
				Accrual:    0,
				Number:     "1114",
				Status:     models.OrderStatusProcessing,
				UploadedAt: time.Time{},
			},
		}, nil)

	r.EXPECT().
		ListOrders(gomock.Any(), "1e1a3858-46e2-40e9-a30d-47f8ecbb5c97").
		Return([]models.Order{}, nil)

	r.EXPECT().
		ListOrders(gomock.Any(), "97e19bd5-d4bd-4d26-a631-b3f5cc9272ff").
		Return([]models.Order{}, errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
		workers       int
	}
	type args struct {
		ctx  context.Context
		uuid string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []models.OrderResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "List of orders",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "82b6779d-557b-446c-a4e7-c67b62509119",
			},
			want: []models.OrderResponse{
				{
					Accrual:    10,
					Number:     "1111",
					Status:     models.OrderStatusProcessed,
					UploadedAt: time.Time{}.Format(timeLayout),
				},
				{
					Accrual:    0,
					Number:     "1112",
					Status:     models.OrderStatusInvalid,
					UploadedAt: time.Time{}.Format(timeLayout),
				},
				{
					Accrual:    0,
					Number:     "1113",
					Status:     models.OrderStatusNew,
					UploadedAt: time.Time{}.Format(timeLayout),
				},
				{
					Accrual:    0,
					Number:     "1114",
					Status:     models.OrderStatusProcessing,
					UploadedAt: time.Time{}.Format(timeLayout),
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "Empty list",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "1e1a3858-46e2-40e9-a30d-47f8ecbb5c97",
			},
			want:      []models.OrderResponse{},
			assertion: assert.NoError,
		},
		{
			name: "Cannot get list of orders",
			fields: fields{
				repository:    r,
				accrualSystem: a,
				workers:       1,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "97e19bd5-d4bd-4d26-a631-b3f5cc9272ff",
			},
			want:      nil,
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, tt.fields.workers)
			got, err := s.ListOrders(tt.args.ctx, tt.args.uuid)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_ListWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		ListWithdrawals(gomock.Any(), "72021a30-0469-4d37-b71c-e79da067d5ec").
		Return([]models.Withdrawal{}, nil)

	r.EXPECT().
		ListWithdrawals(gomock.Any(), "2cd35d02-3e9c-47c2-9dff-492b31cc01b3").
		Return([]models.Withdrawal{
			{
				ID:          12,
				UserUUID:    "2cd35d02-3e9c-47c2-9dff-492b31cc01b3",
				Order:       "12345",
				Sum:         124,
				ProcessedAt: time.Unix(0, 0),
			},
		}, nil)

	r.EXPECT().
		ListWithdrawals(gomock.Any(), "312a65eb-6295-4e7a-94e5-1cecee38aaef").
		Return([]models.Withdrawal{}, errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
	}
	type args struct {
		ctx  context.Context
		uuid string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []models.WithdrawResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Empty list",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "72021a30-0469-4d37-b71c-e79da067d5ec",
			},
			want:      []models.WithdrawResponse{},
			assertion: assert.NoError,
		},
		{
			name: "Valid test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "2cd35d02-3e9c-47c2-9dff-492b31cc01b3",
			},
			want: []models.WithdrawResponse{
				{
					Order:       "12345",
					Sum:         124,
					ProcessedAt: time.Unix(0, 0).Format(timeLayout),
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "Repository error",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "312a65eb-6295-4e7a-94e5-1cecee38aaef",
			},
			want:      []models.WithdrawResponse{},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := *NewService(t.Context(), r, a, 1)

			got, err := s.ListWithdrawals(tt.args.ctx, tt.args.uuid)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		GetUser(gomock.Any(), "cfaa5f89-3c16-4270-9800-686835fcd94f").
		Return(models.User{
			UUID:             "cfaa5f89-3c16-4270-9800-686835fcd94f",
			Login:            "test",
			PasswordHash:     "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			CurrentBalance:   1245.9,
			WithdrawnBalance: 123,
		}, nil)

	r.EXPECT().
		GetUser(gomock.Any(), "566a17a7-065b-40a7-b45f-b5e56f081caf").
		Return(models.User{}, errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
	}
	type args struct {
		ctx  context.Context
		uuid string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      models.BalanceResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "cfaa5f89-3c16-4270-9800-686835fcd94f",
			},
			want: models.BalanceResponse{
				Current:   1245.9,
				Withdrawn: 123,
			},
			assertion: assert.NoError,
		},
		{
			name: "Invalid test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "566a17a7-065b-40a7-b45f-b5e56f081caf",
			},
			want:      models.BalanceResponse{},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, 1)

			got, err := s.GetBalance(tt.args.ctx, tt.args.uuid)

			tt.assertion(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_GetPasswordHash(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		GetUser(gomock.Any(), "9d29b459-2d5a-466f-acbf-cf0d51c00804").
		Return(models.User{
			UUID:             "9d29b459-2d5a-466f-acbf-cf0d51c00804",
			Login:            "test",
			PasswordHash:     "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			CurrentBalance:   123.4,
			WithdrawnBalance: 12.3,
		}, nil)

	r.EXPECT().
		GetUser(gomock.Any(), "c654cb14-fb8e-4a0f-9766-f9481294ed5c").
		Return(models.User{}, errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
	}
	type args struct {
		ctx  context.Context
		uuid string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "9d29b459-2d5a-466f-acbf-cf0d51c00804",
			},
			want:      "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			assertion: assert.NoError,
		},
		{
			name: "An error",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx:  context.TODO(),
				uuid: "c654cb14-fb8e-4a0f-9766-f9481294ed5c",
			},
			want:      "",
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, 1)

			got, err := s.GetPasswordHash(tt.args.ctx, tt.args.uuid)

			tt.assertion(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_updateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := mocks.NewMockRepository(ctrl)

	r.EXPECT().
		UpdateOrderStatus(gomock.Any(), "0000", models.OrderStatusNew).
		Return(nil)

	r.EXPECT().
		UpdateOrderStatus(gomock.Any(), "0001", models.OrderStatusProcessing).
		Return(nil)

	r.EXPECT().
		UpdateOrderStatus(gomock.Any(), "0002", models.OrderStatusInvalid).
		Return(nil)

	r.EXPECT().
		UpdateOrder(gomock.Any(), models.Order{
			UserUUID:   "",
			Accrual:    123,
			Number:     "0003",
			Status:     models.OrderStatusProcessed,
			UploadedAt: time.Time{},
		}).
		Return(nil)

	r.EXPECT().
		UpdateOrderStatus(gomock.Any(), "0004", models.OrderStatusNew).
		Return(errors.New(""))

	a := mocks.NewMockAccrualSystem(ctrl)

	type fields struct {
		repository    Repository
		accrualSystem AccrualSystem
	}
	type args struct {
		ctx   context.Context
		order models.AccrualOrderState
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Status registered",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx: nil,
				order: models.AccrualOrderState{
					Order:   "0000",
					Status:  models.AccrualOrderStatusRegistered,
					Accrual: 0,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "Status processing",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx: nil,
				order: models.AccrualOrderState{
					Order:   "0001",
					Status:  models.AccrualOrderStatusProcessing,
					Accrual: 0,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "Status invalid",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx: nil,
				order: models.AccrualOrderState{
					Order:   "0002",
					Status:  models.AccrualOrderStatusInvalid,
					Accrual: 0,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "Status processed",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx: nil,
				order: models.AccrualOrderState{
					Order:   "0003",
					Status:  models.AccrualOrderStatusProcessed,
					Accrual: 123,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "An error",
			fields: fields{
				repository:    r,
				accrualSystem: a,
			},
			args: args{
				ctx: nil,
				order: models.AccrualOrderState{
					Order:   "0004",
					Status:  models.AccrualOrderStatusRegistered,
					Accrual: 0,
				},
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(t.Context(), tt.fields.repository, tt.fields.accrualSystem, 1)
			tt.assertion(t, s.updateOrder(tt.args.ctx, tt.args.order))
		})
	}
}
