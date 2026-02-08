package cartservice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	cartservice "github.com/codepnw/go-starter-kit/internal/features/cart/service"
	"github.com/codepnw/go-starter-kit/internal/features/product"
	productservice "github.com/codepnw/go-starter-kit/internal/features/product/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	mockUserID       = "mock-uuid-user-id-1"
	mockCartID int64 = 100
)

var (
	mockProductData = &product.Product{ID: 1, Name: "IPhone-17", Price: 43900, Stock: 10}
	ErrDB           = errors.New("DB Error")
)

func TestAddItem(t *testing.T) {
	type inputData struct {
		userID    string
		productID int64
		quantity  int
	}

	type testCase struct {
		name        string
		input       inputData
		mockFn      func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:  "success",
			input: inputData{userID: mockUserID, productID: 1, quantity: 2},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockProd.EXPECT().GetProduct(gomock.Any(), input.productID).Return(mockProductData, nil).Times(1)

				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(mockCartID, nil).Times(1)

				mockRepo.EXPECT().AddItem(gomock.Any(), mockCartID, input.productID, input.quantity).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:  "fail product not found",
			input: inputData{userID: mockUserID, productID: 1, quantity: 20},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockProd.EXPECT().GetProduct(gomock.Any(), input.productID).Return(nil, errs.ErrProductNotFound).Times(1)
			},
			expectedErr: errs.ErrProductNotFound,
		},
		{
			name:  "fail stock not enough",
			input: inputData{userID: mockUserID, productID: 1, quantity: 20},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockProd.EXPECT().GetProduct(gomock.Any(), input.productID).Return(mockProductData, nil).Times(1)
			},
			expectedErr: errs.ErrStockNotEnough,
		},
		{
			name:  "fail get cart id",
			input: inputData{userID: mockUserID, productID: 1, quantity: 2},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockProd.EXPECT().GetProduct(gomock.Any(), input.productID).Return(mockProductData, nil).Times(1)

				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(int64(0), ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:  "fail add item",
			input: inputData{userID: mockUserID, productID: 1, quantity: 2},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockProd.EXPECT().GetProduct(gomock.Any(), input.productID).Return(mockProductData, nil).Times(1)

				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(mockCartID, nil).Times(1)

				mockRepo.EXPECT().AddItem(gomock.Any(), mockCartID, input.productID, input.quantity).Return(ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		service, mockRepo, mockProd := setup(t)

		tc.mockFn(mockRepo, mockProd, tc.input)

		err := service.AddItem(context.Background(), tc.input.userID, tc.input.productID, tc.input.quantity)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGetCart(t *testing.T) {
	type testCase struct {
		name        string
		userID      string
		mockFn      func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, userID string)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:   "success",
			userID: mockUserID,
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, userID string) {
				mockItems := []*cart.CartItemResult{
					{ID: 1, ProductID: 10, Quantity: 2, ProductName: "IPhone-17", Price: 36900, Stock: 12},
					{ID: 2, ProductID: 20, Quantity: 1, ProductName: "Macbook air M4", Price: 32900, Stock: 7},
				}
				mockRepo.EXPECT().GetCartItems(gomock.Any(), userID).Return(mockItems, nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:   "fail get items",
			userID: mockUserID,
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, userID string) {
				mockRepo.EXPECT().GetCartItems(gomock.Any(), userID).Return(nil, ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		service, mockRepo, mockProd := setup(t)

		tc.mockFn(mockRepo, mockProd, tc.userID)

		resp, err := service.GetCart(context.Background(), tc.userID)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		}
	}
}

func TestRemoveItem(t *testing.T) {
	type inputData struct {
		userID    string
		productID int64
	}

	type testCase struct {
		name        string
		input      inputData
		mockFn      func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData)
		expectedErr error
	}
	
	testCases := []testCase{
		{
			name: "success",
			input: inputData{userID: mockUserID, productID: int64(10)},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(mockCartID, nil).Times(1)
				
				mockRepo.EXPECT().RemoveItem(gomock.Any(), mockCartID, input.productID).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "fail get cart id",
			input: inputData{userID: mockUserID, productID: int64(10)},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(int64(0), ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name: "fail remove item",
			input: inputData{userID: mockUserID, productID: int64(10)},
			mockFn: func(mockRepo *cartrepository.MockCartRepository, mockProd *productservice.MockProductService, input inputData) {
				mockRepo.EXPECT().FindCartID(gomock.Any(), input.userID).Return(mockCartID, nil).Times(1)
				
				mockRepo.EXPECT().RemoveItem(gomock.Any(), mockCartID, input.productID).Return(ErrDB).Times(1)
			},
			expectedErr: ErrDB,
		},
	}
	
	for _, tc := range testCases {
		service, mockRepo, mockProd := setup(t)

		tc.mockFn(mockRepo, mockProd, tc.input)

		err := service.RemoveItem(context.Background(), tc.input.userID, tc.input.productID)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func setup(t *testing.T) (cartservice.CartService, *cartrepository.MockCartRepository, *productservice.MockProductService) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := cartrepository.NewMockCartRepository(ctrl)
	mockProd := productservice.NewMockProductService(ctrl)

	service := cartservice.NewCartService(mockRepo, mockProd)

	return service, mockRepo, mockProd
}
