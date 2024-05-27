package api

import (
	"net/http"
	"os"
	"payment-system-one/internal/middleware"
	"payment-system-one/internal/models"
	"payment-system-one/internal/util"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Readiness is to check if server is up
func (u *HTTPHandler) Readiness(c *gin.Context) {
	data := "server is up and running"

	// healthcheck
	util.Response(c, "Ready to go", 200, data, nil)
}

// Create a user
func (u *HTTPHandler) CreateUser(c *gin.Context) {
	var user *models.User
	if err := c.ShouldBind(&user); err != nil {
		util.Response(c, "invalid request", 400, "bad request body", nil)
		return
	}

	//validate user email

	//validate user password

	//persist information in the data base
	err := u.Repository.CreateUser(user)
	if err != nil {
		util.Response(c, "user not created", 400, err.Error(), nil)
		return
	}
	util.Response(c, "user created", 200, "success", nil)
}

//New added code by me
//Create admin function for handler

func (u *HTTPHandler) CreateAdmin(c *gin.Context) {
	var admin *models.Admin
	if err := c.ShouldBind(&admin); err != nil {
		util.Response(c, "invalid request", 400, "bad request body", nil)
		return
	}
	err := u.Repository.CreateAdmin(admin)
	if err != nil {
		util.Response(c, "admin not created", 400, err.Error(), nil)
		return
	}
	util.Response(c, "admin created", 200, "success", nil)

}

// admin login
func (u *HTTPHandler) LoginAdmin(c *gin.Context) {
	var adminLoginRequest *models.AdminRequest
	if err := c.ShouldBind(&adminLoginRequest); err != nil {
		util.Response(c, "invalid request", 400, "bad request body", nil)
		return
	}
	if adminLoginRequest.Email == "" || adminLoginRequest.Password == "" {
		util.Response(c, "Please enter your email or password", 400, "bad request body", nil)
		return
	}
	admin, err := u.Repository.FindAdminByEmail(adminLoginRequest.Email)
	if err != nil {
		util.Response(c, "user does not exist", 404, "user not found", nil)
		return
	}
	if admin.LoginCounter >= 3 {
		admin.IsLocked = true
		admin.UpdatedAt = time.Now()
		err = u.Repository.UpdateAdmin(admin)
		if err != nil {
			return
		}
		util.Response(c, "Your account has been lock after 3 failed attempt, contact IT Support for assistance", 200, "success", nil)
		return
	}
	/*hashedPassword,err:=util.HashPassword(admin.Password)
	if err!=nill{
		util.Response(c, "internal server error", 500, "user not found", nil)
		return
	}
	*/

	if admin.Password != adminLoginRequest.Password {
		admin.LoginCounter++
		err := u.Repository.UpdateAdmin(admin)
		if err != nil {
			util.Response(c, "internal server error", 500, "user not found", nil)
			return
		}
		util.Response(c, "password mismatch", 404, "user not found", nil)
		return
	}

	// Generate token for access and refresh
	accessClaims, refreshClaims := middleware.GenerateClaims(admin.Email)

	secret := os.Getenv("JWT_SECRET")

	accessToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, accessClaims, &secret)
	if err != nil {
		util.Response(c, "error generating access token", 500, "error generating access token", nil)
		return
	}
	refreshToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, refreshClaims, &secret)
	if err != nil {
		util.Response(c, "error generating refresh token", 500, "error generating refresh token", nil)
		return
	}
	c.Header("access_token", *accessToken)
	c.Header("refresh_token", *refreshToken)

	util.Response(c, "login successful", http.StatusOK, gin.H{
		"admin":         admin,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil)
}

// login user
func (u *HTTPHandler) LoginUser(c *gin.Context) {
	var loginRequest *models.LoginRequest
	if err := c.ShouldBind(&loginRequest); err != nil {
		util.Response(c, "invalid request", 400, "bad request body", nil)
		return
	}
	if loginRequest.Email == "" || loginRequest.Password == "" {
		util.Response(c, "Please enter your email or password", 400, "bad request body", nil)
		return
	}

	user, err := u.Repository.FindUserByEmail(loginRequest.Email)
	if err != nil {
		util.Response(c, "user does not exist", 404, "user not found", nil)
		return
	}
	if user.LoginCounter >= 3 {
		user.IsLocked = true
		user.UpdatedAt = time.Now()
		err = u.Repository.UpdateUser(user)
		if err != nil {
			return
		}
		util.Response(c, "Your account has been lock after 3 failed attempt, contact customer care for assistance", 200, "success", nil)
		return
	}

	if user.Password != loginRequest.Password {
		user.LoginCounter++
		err := u.Repository.UpdateUser(user)
		if err != nil {
			util.Response(c, "internal server error", 500, "user not found", nil)
			return
		}
		util.Response(c, "password mismatch", 404, "user not found", nil)
		return
	}

	//Generate token
	accessClaims, refreshClaims := middleware.GenerateClaims(user.Email)

	secret := os.Getenv("JWT_SECRET")

	accessToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, accessClaims, &secret)
	if err != nil {
		util.Response(c, "error generating access token", 500, "error generating access token", nil)
		return
	}
	refreshToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, refreshClaims, &secret)
	if err != nil {
		util.Response(c, "error generating refresh token", 500, "error generating refresh token", nil)
		return
	}
	c.Header("access_token", *accessToken)
	c.Header("refresh_token", *refreshToken)

	util.Response(c, "login successful", http.StatusOK, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil)
}

// call a protected route
func (u *HTTPHandler) GetUserByEmail(c *gin.Context) {
	_, err := u.GetUserFromContext(c)
	if err != nil {
		util.Response(c, "User not logged in", 500, "user not found", nil)
		return
	}

	email := c.Query("email")

	if email == "" {
		util.Response(c, "email is required", 400, "email is required", nil)
		return
	}

	user, err := u.Repository.FindUserByEmail(email)
	if err != nil {
		util.Response(c, "user not fount", 500, "user not found", nil)
		return
	}

	util.Response(c, "user found", 200, user, nil)
}

//query parameter
//path parameter

//100 ---- informtional
//200 ---- success 200, 201, 202
//300 ---- redirect
//400 ---- client error
//500 ----- server error

//syntax error
