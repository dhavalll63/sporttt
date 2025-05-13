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

type AuthController struct {
	repo AuthRepository
}

func NewAuthController(repo AuthRepository) *AuthController {
	return &AuthController{
		repo: repo,
	}
}

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
func (ac *AuthController) Register(c *gin.Context) {
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

	u := user.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     req.Roles,
		Phone:    req.Phone,
	}

	if err := ac.repo.CreateUser(&u); err != nil {
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
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	u, err := ac.repo.GetUserByEmail(req.Email)
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

	// Save refresh token to database
	refreshTokenObj := user.RefreshToken{
		UserID:    u.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, 7), // 7 days
		Revoked:   false,
	}
	if err := ac.repo.SaveRefreshToken(&refreshTokenObj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
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
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req AccessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Verify refresh token from database
	token, err := ac.repo.GetRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateJWT(token.UserID, 15)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}

// RequestOTP handles sending an OTP to the user's phone.
//
// @Summary      Request OTP
// @Description  Generate and send an OTP to the user's phone number
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  auth.OTPReq  true  "Phone Number Request"
// @Success      200  {object}  map[string]string  "OTP sent successfully"
// @Failure      400  {object}  map[string]string  "Invalid phone number"
// @Failure      500  {object}  map[string]string  "Failed to create OTP"
// @Router       /auth/request-otp [post]
func (ac *AuthController) RequestOTP(c *gin.Context) {
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

	if err := ac.repo.SaveOTP(&otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OTP"})
		return
	}

	// Send OTP via SMS
	if err := sendOTPToPhone(req.Phone, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent"})
}

// VerifyOTP verifies the OTP and logs in or registers the user.
//
// @Summary      Verify OTP and Login/Register
// @Description  Verify the OTP, auto-register if user doesn't exist, and return tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  auth.OTPVerifyRequest  true  "OTP Verification Request"
// @Success      200  {object}  map[string]interface{}  "OTP verified, tokens and user info returned"
// @Failure      400  {object}  map[string]string  "Invalid input"
// @Failure      401  {object}  map[string]string  "Invalid or expired OTP"
// @Router       /auth/verify-otp [post]
func (ac *AuthController) VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Verify OTP
	otp, err := ac.repo.GetOTP(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// Mark OTP as verified
	otp.Verified = true
	ac.repo.UpdateOTP(otp)

	// Find or create user
	var newUser *user.User
	newUser, err = ac.repo.GetUserByPhone(req.Phone)

	if err != nil {
		// Auto-register user if not exists
		role := user.Role{Name: "player"}
		if err := config.DB.FirstOrCreate(&role, user.Role{Name: "player"}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
			return
		}

		newUser = &user.User{
			Username:      "user_" + req.Phone,
			Phone:         req.Phone,
			PhoneVerified: true,
			Role:          "player",
		}

		if err := ac.repo.CreateUser(newUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// Update phone verified status
		newUser.PhoneVerified = true
		if err := ac.repo.UpdateUser(newUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Generate tokens
	accessToken, _ := utils.GenerateJWT(newUser.ID, 15)
	refreshToken, _ := utils.GenerateRefreshToken(newUser.ID)

	// Save refresh token
	refreshTokenObj := user.RefreshToken{
		UserID:    newUser.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, 7), // 7 days
		Revoked:   false,
	}
	ac.repo.SaveRefreshToken(&refreshTokenObj)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          newUser,
	})
}

// ResendOTP handles the request to resend an OTP to a user's phone number.
// @Summary Resend OTP
// @Description Resends an OTP to the provided phone number. If an unexpired OTP exists and was sent recently, it will be reused.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body OTPReq true "Phone number for which OTP needs to be resent"
// @Success 200 {object} map[string]string "OTP resent successfully"
// @Failure 400 {object} map[string]string "Invalid phone number"
// @Failure 429 {object} map[string]string "OTP recently sent. Please wait."
// @Failure 500 {object} map[string]string "Could not generate OTP"
// @Router /auth/resend-otp [post]
func (ac *AuthController) ResendOTP(c *gin.Context) {
	var req OTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
		return
	}

	// Check existing unexpired OTP
	existing, err := ac.repo.GetLatestOTP(req.Phone)

	if err == nil {
		// Return early if recently sent (e.g. less than 1 minute ago)
		if time.Since(existing.CreatedAt) < 1*time.Minute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "OTP recently sent. Please wait."})
			return
		}

		// Reuse existing OTP
		if err := sendOTPToPhone(req.Phone, existing.Code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
			return
		}

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

	if err := ac.repo.SaveOTP(&newOTP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate OTP"})
		return
	}

	if err := sendOTPToPhone(req.Phone, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "New OTP sent"})
}

// ForgotPassword initiates the password reset process
// @Summary Request password reset
// @Description Sends a password reset link to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} map[string]string "Reset link sent"
// @Failure 400 {object} map[string]string "Invalid email"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Failed to send reset email"
// @Router /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	// Get user by email
	u, err := ac.repo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate reset token
	resetToken := utils.GenerateRandomToken(32)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Update user with reset token
	u.ResetToken = resetToken
	u.ResetExpires = &expiresAt

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	// TODO: Send email with reset token
	// For now just log it
	fmt.Printf("Password reset token for user %s: %s\n", u.Email, resetToken)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent to email"})
}

// ResetPassword resets a user's password with a valid token
// @Summary Reset password
// @Description Resets the user's password using a valid reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} map[string]string "Password updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Failure 500 {object} map[string]string "Failed to update password"
// @Router /auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Find user by reset token
	var u user.User
	if err := config.DB.Where("reset_token = ? AND reset_expires > ?", req.Token, time.Now()).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password and clear reset token
	u.Password = hashedPassword
	u.ResetToken = ""
	u.ResetExpires = nil

	if err := ac.repo.UpdateUser(&u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
