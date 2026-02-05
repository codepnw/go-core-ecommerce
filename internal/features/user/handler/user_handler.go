package userhandler

import (
	"net/http"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/user"
	userservice "github.com/codepnw/go-starter-kit/internal/features/user/service"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	service userservice.UserService
}

func NewUserHandler(service userservice.UserService) *userHandler {
	return &userHandler{service: service}
}

func (h *userHandler) Register(c *gin.Context) {
	req := new(RegisterReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	input := &user.User{
		Email:    req.Email,
		Password: req.Password,
	}
	resp, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		switch err {
		case errs.ErrEmailAlreadyExists:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusCreated, resp)
}

func (h *userHandler) Login(c *gin.Context) {
	req := new(LoginReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case errs.ErrInvalidEmailOrPassword:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}

func (h *userHandler) RefreshToken(c *gin.Context) {
	req := new(RefreshTokenReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case errs.ErrTokenNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		case errs.ErrTokenRevoked:
			response.ResponseError(c, http.StatusBadRequest, err)
		case errs.ErrTokenExpires:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}

func (h *userHandler) Logout(c *gin.Context) {
	req := new(RefreshTokenReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		switch err {
		case errs.ErrTokenNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusNoContent, nil)
}

func (h *userHandler) GetProfile(c *gin.Context) {
	resp, err := h.service.GetProfile(c.Request.Context())
	if err != nil {
		switch err {
		case errs.ErrUserNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}
	response.ResponseSuccess(c, http.StatusOK, resp)
}
