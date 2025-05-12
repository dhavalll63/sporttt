package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DhavalSuthar-24/miow/config"
	"github.com/gin-gonic/gin"

	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/utils"
)

func sendOTPToPhone(phone, code string) error {
	// Replace this with real SMS integration
	fmt.Printf("Sending OTP %s to %s\n", code, phone)
	return nil
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user with username, email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body  RegisterRequest  true  "Register"
// @Success      201   {object} map[string]string
// @Failure      400   {object} map[string]string
// @Failure      500   {object} map[string]string
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	var roles []user.Role
	if len(req.Roles) > 0 {
		if err := config.DB.Where("name IN ?", req.Roles).Find(&roles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching roles"})
			return
		}
	}
	u := user.User{Username: req.Username, Email: req.Email, Password: hashed, Roles: roles}
	if err := CreateUser(&u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body  LoginRequest  true  "Login"
// @Success      200   {object} map[string]string
// @Failure      400   {object} map[string]string
// @Failure      401   {object} map[string]string
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	u, err := GetUserByEmail(req.Email)
	if err != nil || !utils.CheckPassword(u.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := utils.GenerateJWT(u.ID, 15) // expires in 15 minutes
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken handles the refresh token process.
//
// @Summary Refresh Access Token
// @Description Refreshes the access token using a valid refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body AccessReq true "Refresh Token Request"
// @Success 200 {object} map[string]string "Returns a new access token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 500 {object} map[string]string "Token generation failed"
// @Router /auth/refresh-token [post]
func RefreshToken(c *gin.Context) {
	var req AccessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, err := utils.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	newAccessToken, err := utils.GenerateJWT(userID, 15)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}

func RequestOTP(c *gin.Context) {
	var req OTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
		return
	}
	code := utils.GenerateOTP()
	otp := OTP{
		Phone:     req.Phone,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OTP"})
		return

	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent", "otp": code})
}

func VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var otp OTP
	if err := config.DB.Where("phone = ? AND code = ? AND verified = false AND expires_at > ?", req.Phone, req.Code, time.Now()).First(&otp).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	otp.Verified = true
	config.DB.Save(&otp)

	var newUser user.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&newUser).Error; err != nil {
		// Auto-register user if not exists

		newUser = user.User{
			Username: "user_" + req.Phone,
			Phone:    req.Phone,
			Roles:    []user.Role{{Name: "player"}}, // Ensure Role type is defined in the user package
		}
		config.DB.Create(&newUser)
	}

	accessToken, _ := utils.GenerateJWT(newUser.ID, 15)
	refreshToken, _ := utils.GenerateRefreshToken(newUser.ID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          newUser,
	})

}

func ResendOTP(c *gin.Context) {
	var req OTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
		return
	}

	// Check existing unexpired OTP
	var existing OTP
	err := config.DB.
		Where("phone = ? AND verified = false AND expires_at > ?", req.Phone, time.Now()).
		Order("created_at desc").
		First(&existing).Error

	if err == nil {
		// Return early if recently sent (e.g. less than 1 minute ago)
		if time.Since(existing.CreatedAt) < 1*time.Minute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "OTP recently sent. Please wait."})
			return
		}

		// Reuse existing OTP
		_ = sendOTPToPhone(req.Phone, existing.Code)
		c.JSON(http.StatusOK, gin.H{"message": "OTP resent"})
		return
	}

	// Generate new OTP
	code := utils.GenerateOTP()
	newOTP := OTP{
		Phone:     req.Phone,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := config.DB.Create(&newOTP).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate OTP"})
		return
	}

	_ = sendOTPToPhone(req.Phone, code)
	c.JSON(http.StatusOK, gin.H{"message": "New OTP sent"})
}
