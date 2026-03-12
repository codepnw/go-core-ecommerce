package orderservice_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	"github.com/codepnw/go-starter-kit/internal/features/order"
	orderrepository "github.com/codepnw/go-starter-kit/internal/features/order/repository"
	orderservice "github.com/codepnw/go-starter-kit/internal/features/order/service"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var ErrDB = errors.New("database error")

func TestCreateOrder(t *testing.T) {
	type createOrderInput struct {
		userID  string
		address string
	}

	type testCase struct {
		name        string
		input       createOrderInput
		mockFn      func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockRedis *productrepository.MockProductRedisRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:  "success",
			input: createOrderInput{userID: "mock-uuid-1", address: "Bangkok, Thailand"},
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockRedis *productrepository.MockProductRedisRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput) {
				mockItems := []*cart.CartItemResult{
					{ID: 1, ProductID: int64(101), Quantity: 2, ProductName: "IPhone-17", Price: 44900, Stock: 10},
					{ID: 2, ProductID: int64(102), Quantity: 1, ProductName: "Macbook-air-M4", Price: 34900, Stock: 5},
				}
				mockCart.EXPECT().GetCartItems(gomock.Any(), input.userID).Return(mockItems, nil).Times(1)
				
				for _, item := range mockItems {
					mockRedis.EXPECT().CheckStockExists(gomock.Any(), item.ProductID).Return(true, nil).Times(1)
					
					mockRedis.EXPECT().DecreaseStock(gomock.Any(), item.ProductID, item.Quantity).Return(int64(1), nil).Times(1)
				}

				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				mockOrder.EXPECT().InsertOrderTx(gomock.Any(), gomock.Any(), input.userID, gomock.Any(), input.address).Return(int64(101), time.Time{}, nil).Times(1)

				for _, i := range mockItems {
					mockProd.EXPECT().DecreaseStockTx(gomock.Any(), gomock.Any(), i.ProductID, i.Quantity).Return(nil).Times(1)

					mockOrder.EXPECT().InsertOrderItemTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				}

				mockCart.EXPECT().ClearCartTx(gomock.Any(), gomock.Any(), input.userID).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:  "fail cart empty",
			input: createOrderInput{userID: "mock-uuid-1", address: "Bangkok, Thailand"},
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockRedis *productrepository.MockProductRedisRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput) {
				mockItems := []*cart.CartItemResult{}
				mockCart.EXPECT().GetCartItems(gomock.Any(), input.userID).Return(mockItems, nil).Times(1)
			},
			expectedErr: errs.ErrCartEmpty,
		},
	}

	for _, tc := range testCases {
		deps := setup(t)

		tc.mockFn(deps.MockTx, deps.MockOrd, deps.MockProd, deps.MockProdRedis, deps.MockCart, tc.input)

		orderNo, err := deps.Serivce.CreateOrder(context.Background(), tc.input.userID, tc.input.address)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotEmpty(t, orderNo)
		}
	}
}

func TestGetOrderDetails(t *testing.T) {
	type testCase struct {
		name        string
		orderID     int64
		mockFn      func(mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, orderID int64)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:    "success",
			orderID: 101,
			mockFn: func(mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, orderID int64) {
				mockOrderData := &order.Order{
					Items: []order.OrderItem{
						{OrderID: orderID, ProductID: 101, ProductName: "IPhone-17", Quantity: 1, Price: 35000},
						{OrderID: orderID, ProductID: 102, ProductName: "IPhone-17-Pro", Quantity: 2, Price: 45000},
					},
				}
				mockOrder.EXPECT().FindOrderDetails(gomock.Any(), orderID).Return(mockOrderData, nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:    "fail get order",
			orderID: 101,
			mockFn: func(mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, orderID int64) {
				mockOrder.EXPECT().FindOrderDetails(gomock.Any(), orderID).Return(nil, ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		deps := setup(t)

		tc.mockFn(deps.MockOrd, deps.MockProd, deps.MockCart, tc.orderID)

		resp, err := deps.Serivce.GetOrderDetails(context.Background(), tc.orderID)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		}
	}
}

func TestMyOrders(t *testing.T) {
	type testCase struct {
		name        string
		userID      string
		mockFn      func(mockOrder *orderrepository.MockOrderRepository, userID string)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:   "success",
			userID: "mock-uuid-01",
			mockFn: func(mockOrder *orderrepository.MockOrderRepository, userID string) {
				mockOrdersResp := []*order.Order{
					{ID: 1, TotalAmount: 2000, Status: order.StatusPending, CreatedAt: time.Now()},
					{ID: 2, TotalAmount: 500, Status: order.StatusPending, CreatedAt: time.Now()},
				}
				mockOrder.EXPECT().FindMyOrders(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(mockOrdersResp, int64(10), nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:   "fail",
			userID: "mock-uuid-01",
			mockFn: func(mockOrder *orderrepository.MockOrderRepository, userID string) {
				mockOrder.EXPECT().FindMyOrders(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(nil, int64(0), ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		deps := setup(t)

		tc.mockFn(deps.MockOrd, tc.userID)

		resp, err := deps.Serivce.MyOrders(context.Background(), "mock-uuid-01", 0, 0)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	type testCase struct {
		name        string
		orderID     int64
		newStatus   order.OrderStatus
		mockFn      func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, orderID int64, newStatus order.OrderStatus)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:      "success",
			orderID:   101,
			newStatus: order.StatusPaid,
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, orderID int64, newStatus order.OrderStatus) {
				mockData := &order.Order{Status: "PENDING"}
				mockOrder.EXPECT().FindOrderDetails(gomock.Any(), orderID).Return(mockData, nil).Times(1)

				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				)

				mockOrder.EXPECT().UpdateStatusTx(gomock.Any(), gomock.Any(), orderID, newStatus).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:      "fail status transition",
			orderID:   101,
			newStatus: order.StatusShipped,
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, orderID int64, newStatus order.OrderStatus) {
				mockData := &order.Order{Status: "PENDING"}
				mockOrder.EXPECT().FindOrderDetails(gomock.Any(), orderID).Return(mockData, nil).Times(1)
			},
			expectedErr: errors.New("invalid status transition"),
		},
		{
			name:      "fail update status",
			orderID:   101,
			newStatus: order.StatusPaid,
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, orderID int64, newStatus order.OrderStatus) {
				mockData := &order.Order{Status: "PENDING"}
				mockOrder.EXPECT().FindOrderDetails(gomock.Any(), orderID).Return(mockData, nil).Times(1)

				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				)

				mockOrder.EXPECT().UpdateStatusTx(gomock.Any(), gomock.Any(), orderID, newStatus).Return(ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		deps := setup(t)

		tc.mockFn(deps.MockTx, deps.MockOrd, tc.orderID, tc.newStatus)

		err := deps.Serivce.UpdateStatus(context.Background(), tc.orderID, tc.newStatus)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

type setupDeps struct {
	Serivce       orderservice.OrderService
	MockTx        *database.MockTxManager
	MockOrd       *orderrepository.MockOrderRepository
	MockProd      *productrepository.MockProductRepository
	MockProdRedis *productrepository.MockProductRedisRepository
	MockCart      *cartrepository.MockCartRepository
}

func setup(t *testing.T) *setupDeps {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := database.NewMockTxManager(ctrl)
	mockOrd := orderrepository.NewMockOrderRepository(ctrl)
	mockProd := productrepository.NewMockProductRepository(ctrl)
	mockProdRedis := productrepository.NewMockProductRedisRepository(ctrl)
	mockCart := cartrepository.NewMockCartRepository(ctrl)

	service := orderservice.NewOrderService(&orderservice.OrderServiceDeps{
		Tx:        mockTx,
		OrderRepo: mockOrd,
		ProdRepo:  mockProd,
		ProdRedis: mockProdRedis,
		CartRepo:  mockCart,
	})

	return &setupDeps{
		Serivce:       service,
		MockTx:        mockTx,
		MockOrd:       mockOrd,
		MockProd:      mockProd,
		MockProdRedis: mockProdRedis,
		MockCart:      mockCart,
	}
}
