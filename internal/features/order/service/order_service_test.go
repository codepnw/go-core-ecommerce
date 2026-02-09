package orderservice_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	orderrepository "github.com/codepnw/go-starter-kit/internal/features/order/repository"
	orderservice "github.com/codepnw/go-starter-kit/internal/features/order/service"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	type createOrderInput struct {
		userID  string
		address string
	}

	type testCase struct {
		name        string
		input       createOrderInput
		mockFn      func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:  "success",
			input: createOrderInput{userID: "mock-uuid-1", address: "Bangkok, Thailand"},
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput) {
				mockItems := []*cart.CartItemResult{
					{ID: 1, ProductID: 101, Quantity: 2, ProductName: "IPhone-17", Price: 44900, Stock: 10},
					{ID: 2, ProductID: 102, Quantity: 1, ProductName: "Macbook-air-M4", Price: 34900, Stock: 5},
				}
				mockCart.EXPECT().GetCartItems(gomock.Any(), input.userID).Return(mockItems, nil).Times(1)

				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				mockOrder.EXPECT().InsertOrderTx(gomock.Any(), gomock.Any(), input.userID, gomock.Any(), input.address).Return(int64(101), nil).Times(1)

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
			mockFn: func(mockTx *database.MockTxManager, mockOrder *orderrepository.MockOrderRepository, mockProd *productrepository.MockProductRepository, mockCart *cartrepository.MockCartRepository, input createOrderInput) {
				mockItems := []*cart.CartItemResult{}
				mockCart.EXPECT().GetCartItems(gomock.Any(), input.userID).Return(mockItems, nil).Times(1)
			},
			expectedErr: errs.ErrCartEmpty,
		},
	}

	for _, tc := range testCases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTx := database.NewMockTxManager(ctrl)
		mockOrd := orderrepository.NewMockOrderRepository(ctrl)
		mockProd := productrepository.NewMockProductRepository(ctrl)
		mockCart := cartrepository.NewMockCartRepository(ctrl)

		tc.mockFn(mockTx, mockOrd, mockProd, mockCart, tc.input)

		service := orderservice.NewOrderService(mockTx, mockOrd, mockProd, mockCart)

		orderNo, err := service.CreateOrder(context.Background(), tc.input.userID, tc.input.address)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotEmpty(t, orderNo)
		}
	}
}
