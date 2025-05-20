package auth

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/DhavalSuthar-24/miow/config"              // For DB and other app config
	"github.com/DhavalSuthar-24/miow/internal/middleware" // Assuming your middleware is here for GetUserIDFromContext
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/pkg/token" // Assuming token utilities are here
	"github.com/DhavalSuthar-24/miow/pkg/utils" // General utilities like hashing, OTP
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxOTPSendAttempts = 5 // Max attempts for sending OTP before cooldown
	otpCooldownMinutes = 1 // Cooldown period in minutes
	otpExpiryMinutes   = 5 // OTP expiry time
	DefaultUserRole    = "player"
)

type AuthController struct {
	repo   AuthRepository
	config *config.Config // If you have a general config struct
	// mailer MailerService // Interface for sending emails
	// sms    SMSService    // Interface for sending SMS
}

func NewAuthController(repo AuthRepository, cfg *config.Config /* mailer MailerService, sms SMSService*/) *AuthController {
	return &AuthController{
		repo:   repo,
		config: cfg,
		// mailer: mailer,
		// sms:    sms,
	}
}

func (ac *AuthController) generateAndSaveTokens(c *gin.Context, userID uint) (string, string, error) {
	accessToken, err := token.GenerateJWT(userID, ac.config.JWT.AccessTokenSecret, ac.config.JWT.AccessTokenExpiryMinutes)
	if err != nil {
		return "", "", fmt.Errorf("access token generation failed: %w", err)
	}

	refreshTokenString, err := token.GenerateRefreshToken(userID, ac.config.JWT.RefreshTokenSecret, ac.config.JWT.RefreshTokenExpiryDays)
	if err != nil {
		return "", "", fmt.Errorf("refresh token generation failed: %w", err)
	}

	refreshToken := &user.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().AddDate(0, 0, ac.config.JWT.RefreshTokenExpiryDays),
	}

	if err := ac.repo.SaveRefreshToken(refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}
	return accessToken, refreshTokenString, nil
}

// sendOTPToPhone simulates sending OTP. Replace with actual SMS service.
func (ac *AuthController) sendOTPToPhone(phone, otpCode string) error {
	fmt.Printf("SIMULATING: Sending OTP %s to %s\n", otpCode, phone)
	// Example: return ac.sms.Send(phone, fmt.Sprintf("Your OTP code is: %s", otpCode))

	// Integrate with your SMS provider here
	return nil
}

// sendEmail simulates sending an email. Replace with actual email service.
func (ac *AuthController) sendEmail(to, subject, body string) error {
	fmt.Printf("SIMULATING: Sending Email\nTo: %s\nSubject: %s\nBody: %s\n", to, subject, body)

	// Integrate with your Email provider here
	return nil
}

// @Summary      Register a new user
// @Description  Create a new user with username, email, phone and password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body  RegisterRequest  true  "User registration details"
// @Success      201   {object} AuthResponse "User registered successfully, returns tokens and user info"
// @Failure      400   {object} map[string]string "Validation error or invalid input"
// @Failure      409   {object} map[string]string "User with this email or phone or username already exists"
// @Failure      500   {object} map[string]string "Internal server error"
// @Router       /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check for existing users
	if _, err := ac.repo.GetUserByEmail(req.Email); !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	if _, err := ac.repo.GetUserByPhone(req.Phone); !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this phone number already exists"})
		return
	}
	if _, err := ac.repo.GetUserByUsername(req.Username); !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this username already exists"})
		return
	}

	for _, rn := range req.Roles {
		_, err := ac.repo.GetRoleByName(rn)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("role %q does not exist", rn)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "role lookup failed"})
			return
		}

	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	emailVerifyToken := utils.GenerateRandomToken(32)
	emailVerifyExpires := time.Now().Add(24 * time.Hour)

	newUser := &user.User{
		Name:          req.Name,
		Username:      req.Username,
		Email:         strings.ToLower(req.Email),
		Password:      hashedPassword,
		Phone:         req.Phone,
		PhoneVerified: false,
		EmailVerified: false,
		Verified:      false,
		LastActive:    time.Now(),
		VerifyToken:   emailVerifyToken,
		VerifyExpires: &emailVerifyExpires,
	}

	// Set optional fields if provided
	if req.Address != "" {
		newUser.Address = req.Address
	}
	if req.City != "" {
		newUser.City = req.City
	}
	if req.District != "" {
		newUser.District = req.District
	}
	if req.State != "" {
		newUser.State = req.State
	}
	if req.Country != "" {
		newUser.Country = req.Country
	}
	if req.PostalCode != "" {
		newUser.PostalCode = req.PostalCode
	}
	if req.Bio != "" {
		newUser.Bio = req.Bio
	}
	if len(req.PreferredSports) > 0 {
		newUser.PreferredSports = req.PreferredSports
	}
	if req.SocialMedia != nil {
		newUser.SocialMedia = *req.SocialMedia
	}

	// Create user
	if err := ac.repo.CreateUser(newUser); err != nil {
		// Print the real error
		log.Printf("❌ CreateUser failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed: " + err.Error()})
		return
	}
	DefaultUserRoleID := 1

	if len(req.Roles) == 0 {

		if err := ac.repo.AssignRoleToUser(newUser.ID, "player"); err != nil {
			log.Printf("Assign role %d failed: %v", DefaultUserRoleID, err)
		}
	}
	for _, role := range req.Roles {
		if err := ac.repo.AssignRoleToUser(newUser.ID, role); err != nil {
			log.Printf("Assign role %d failed: %v", DefaultUserRoleID, err)
		}
	}

	// Send verification email
	verificationLink := fmt.Sprintf("%s/api/auth/verify-email?token=%s", ac.config.App.FrontendURL, emailVerifyToken)
	emailBody := fmt.Sprintf("Hello %s, please verify your email by clicking on this link: %s", newUser.Name, verificationLink)
	if err := ac.sendEmail(newUser.Email, "Verify Your Email Address", emailBody); err != nil {
		fmt.Printf("Failed to send verification email to %s: %v\n", newUser.Email, err)
	}

	accessToken, refreshToken, err := ac.generateAndSaveTokens(c, newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         FilterUserRecord(newUser),
	})
}

// @Summary      Login user
// @Description  Authenticate user with email/username and password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  LoginRequest  true  "Login credentials"
// @Success      200   {object} AuthResponse "Login successful, returns tokens and user info"
// @Failure      400   {object} map[string]string "Invalid input"
// @Failure      401   {object} map[string]string "Invalid credentials or user not verified"
// @Failure      404   {object} map[string]string "User not found"
// @Failure      500   {object} map[string]string "Internal server error"
// @Router       /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var foundUser *user.User
	var err error

	// Try finding by email first, then by username (if you allow username login)
	foundUser, err = ac.repo.GetUserByEmail(strings.ToLower(req.LoginIdentifier))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// foundUser, err = ac.repo.GetUserByUsername(req.LoginIdentifier) // Uncomment if username login is supported
		// if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
		// }
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	if !utils.CheckPassword(foundUser.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Optional: Check if user is verified (email or phone or both)
	// if !foundUser.Verified {
	//  c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is not verified."})
	//  return
	// }

	accessToken, refreshToken, err := ac.generateAndSaveTokens(c, foundUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	foundUser.LastActive = time.Now()
	if err := ac.repo.UpdateUser(foundUser); err != nil {
		fmt.Printf("Error updating last active for user %d: %v\n", foundUser.ID, err)
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         FilterUserRecord(foundUser),
	})
}

// @Summary      Refresh Access Token
// @Description  Refreshes the access token using a valid refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh Token Request"
// @Success      200 {object} map[string]string "Returns a new access token"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Invalid or expired refresh token"
// @Failure      500 {object} map[string]string "Token generation failed"
// @Router       /auth/refresh-token [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	rt, err := ac.repo.GetRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	newAccessToken, err := token.GenerateJWT(rt.UserID, ac.config.JWT.AccessTokenSecret, ac.config.JWT.AccessTokenExpiryMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "New access token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}

// @Summary      Get User Profile
// @Description  Retrieves the profile of the currently authenticated user.
// @Tags         Profile
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} UserResponse "User profile data"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "User not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/me [get]
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c) // Assumes your middleware sets this
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	currentUser, err := ac.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, FilterUserRecord(currentUser))
}

// @Summary      Update User Profile
// @Description  Updates the profile of the currently authenticated user.
// @Tags         Profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        profileData body UpdateProfileRequest true "Profile data to update"
// @Success      200 {object} UserResponse "Updated user profile data"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "User not found"
// @Failure      409 {object} map[string]string "Username already taken"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/me [put]
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	u, err := ac.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		return
	}

	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Username != nil {
		existingUser, findErr := ac.repo.GetUserByUsername(*req.Username)
		if findErr == nil && existingUser.ID != u.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken."})
			return
		}
		u.Username = *req.Username
	}
	if req.Bio != nil {
		u.Bio = *req.Bio
	}
	if req.Address != nil {
		u.Address = *req.Address
	}
	if req.City != nil {
		u.City = *req.City
	}
	if req.District != nil {
		u.District = *req.District
	}
	if req.State != nil {
		u.State = *req.State
	}
	if req.Country != nil {
		u.Country = *req.Country
	}
	if req.PostalCode != nil {
		u.PostalCode = *req.PostalCode
	}
	if req.PreferredSports != nil {
		u.PreferredSports = req.PreferredSports
	}
	if req.SocialMedia != nil {
		u.SocialMedia = *req.SocialMedia
	}
	if req.Coordinates != nil {
		u.Coordinates = *req.Coordinates
	}

	u.LastActive = time.Now()

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update profile: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, FilterUserRecord(u))
}

// @Summary      Update Profile Image
// @Description  Updates the profile image for the currently authenticated user.
// @Tags         Profile
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Profile image file"
// @Success      200 {object} map[string]string "Profile image updated successfully"
// @Failure      400 {object} map[string]string "Invalid file or input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to upload or save image path"
// @Router       /auth/me/profile-image [put]
func (ac *AuthController) UpdateProfileImage(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required: " + err.Error()})
		return
	}

	// Validate file type and size if necessary
	// E.g., check file.Header.Get("Content-Type") and file.Size

	u, err := ac.repo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		return
	}

	// Generate a unique filename to prevent collisions
	extension := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("user_%d_profile_%d%s", userID, time.Now().UnixNano(), extension)
	uploadPath := filepath.Join(ac.config.App.UploadDir, "profiles", filename) // e.g., ./uploads/profiles/user_1_profile_timestamp.jpg

	// Ensure directory exists
	if err := utils.EnsureDir(filepath.Dir(uploadPath)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory: " + err.Error()})
		return
	}

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded image: " + err.Error()})
		return
	}

	// Store relative path or full URL depending on your setup
	u.ProfileImage = "/uploads/profiles/" + filename // Path accessible by frontend
	u.LastActive = time.Now()

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile image path to database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile image updated successfully", "profile_image_url": u.ProfileImage})
}

// @Summary      Change Password
// @Description  Allows an authenticated user to change their password.
// @Tags         Profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        passwords body ChangePasswordRequest true "Old and new password details"
// @Success      200 {object} map[string]string "Password changed successfully"
// @Failure      400 {object} map[string]string "Invalid input or password mismatch"
// @Failure      401 {object} map[string]string "Unauthorized or incorrect old password"
// @Failure      500 {object} map[string]string "Failed to change password"
// @Router       /auth/change-password [post]
func (ac *AuthController) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	u, err := ac.repo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		return
	}

	if !utils.CheckPassword(u.Password, req.OldPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password."})
		return
	}

	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password cannot be the same as the old password."})
		return
	}

	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password."})
		return
	}

	u.Password = newHashedPassword
	u.LastActive = time.Now()
	// Optionally: Invalidate all other active sessions/refresh tokens for this user
	// if err := ac.repo.InvalidateAllRefreshTokensForUser(u.ID); err != nil { ... }

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully."})
}

// @Summary      Logout User
// @Description  Invalidates the user's current session and refresh tokens (optionally all sessions)
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body LogoutRequest false "Logout options"
// @Success      200 {object} map[string]string "Logged out successfully"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to logout"
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// Get user ID from context (set by your auth middleware)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Try to get refresh token from different sources
	refreshToken := ""
	if req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	} else {
		// Check cookie if no token in request body
		refreshTokenCookie, _ := c.Cookie("refresh_token")
		if refreshTokenCookie != "" {
			refreshToken = refreshTokenCookie
		}
	}

	// Invalidate the specific refresh token if provided
	if refreshToken != "" {
		if err := ac.repo.InvalidateRefreshToken(refreshToken); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate refresh token: " + err.Error()})
				return
			}
			// Token not found is acceptable (maybe already expired/revoked)
		}
	}

	// If requested, invalidate ALL user's refresh tokens
	if req.InvalidateAllSessions {
		if err := ac.repo.InvalidateAllRefreshTokensForUser(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate all sessions: " + err.Error()})
			return
		}
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true) // secure flag true in production
	c.SetCookie("access_token", "", -1, "/", "", false, true)  // if you use access token cookies

	c.JSON(http.StatusOK, gin.H{
		"message":                  "Logged out successfully",
		"all_sessions_invalidated": req.InvalidateAllSessions,
	})
}

// @Summary      Request OTP
// @Description  Generate and send an OTP to the user's phone number for verification or login.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  OTPRequest  true  "Phone Number Request"
// @Success      200  {object}  map[string]string  "OTP sent successfully"
// @Failure      400  {object}  map[string]string  "Invalid phone number format"
// @Failure      429  {object}  map[string]string  "Too many OTP requests. Please try again later."
// @Failure      500  {object}  map[string]string  "Failed to generate or send OTP"
// @Router       /auth/request-otp [post]
func (ac *AuthController) RequestOTP(c *gin.Context) {
	var req OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Optional: Check if user exists if OTP is for a registered user action
	// _, err := ac.repo.GetUserByPhone(req.Phone)
	// if errors.Is(err, gorm.ErrRecordNotFound) {
	//     c.JSON(http.StatusNotFound, gin.H{"error": "User with this phone number not found"})
	//     return
	// }
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
	//     return
	// }

	latestOTP, err := ac.repo.GetLatestOTP(req.Phone)
	if err == nil && latestOTP != nil {
		if latestOTP.Attempt >= maxOTPSendAttempts && time.Since(latestOTP.CreatedAt) < otpCooldownMinutes*time.Minute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("Too many OTP requests. Please try again in %.0f minute(s).", otpCooldownMinutes-time.Since(latestOTP.CreatedAt).Minutes())})
			return
		}
		// If an OTP was sent recently (e.g., within the last 60 seconds), resend it or ask user to wait
		if time.Since(latestOTP.CreatedAt) < 60*time.Second {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "An OTP was recently sent. Please wait a moment before requesting a new one."})
			return
		}
	}

	otpCode := utils.GenerateOTP() // Generate a 6-digit OTP
	otp := &OTP{
		Phone:     req.Phone,
		Code:      otpCode,
		ExpiresAt: time.Now().Add(time.Duration(otpExpiryMinutes) * time.Minute),
		Attempt:   1, // Reset attempt count for new OTP
	}

	if latestOTP != nil && latestOTP.Attempt >= maxOTPSendAttempts {
		otp.Attempt = 1 // Reset if cooldown passed
	} else if latestOTP != nil {
		otp.Attempt = latestOTP.Attempt + 1
	}

	if err := ac.repo.SaveOTP(otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP: " + err.Error()})
		return
	}

	if err := ac.sendOTPToPhone(req.Phone, otpCode); err != nil {
		// Log error, but don't necessarily expose detailed failure to client for security
		fmt.Printf("Failed to send OTP to %s: %v\n", req.Phone, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP. Please try again."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully."})
}

// @Summary      Verify OTP
// @Description  Verify the OTP. If user with phone doesn't exist, create one. Then log in user.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  VerifyOTPRequest  true  "OTP Verification Request"
// @Success      200  {object}  AuthResponse "OTP verified, tokens and user info returned"
// @Failure      400  {object}  map[string]string  "Invalid input or OTP format"
// @Failure      401  {object}  map[string]string  "Invalid, expired, or already used OTP"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /auth/verify-otp [post]
func (ac *AuthController) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	otp, err := ac.repo.GetOTP(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid, expired, or already used OTP."})
		return
	}

	otp.Verified = true
	if err := ac.repo.UpdateOTP(otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update OTP status: " + err.Error()})
		return
	}

	var u *user.User
	u, err = ac.repo.GetUserByPhone(req.Phone)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Auto-register user with minimal information
		newUser := &user.User{
			Name:          "User_" + strings.ReplaceAll(req.Phone, "+", ""),
			Username:      "user_" + strings.ReplaceAll(req.Phone, "+", ""),
			Phone:         req.Phone,
			PhoneVerified: true,
			Verified:      true,
			LastActive:    time.Now(),
		}

		if errCreate := ac.repo.CreateUser(newUser); errCreate != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + errCreate.Error()})
			return
		}

		// Assign default role
		if errRole := ac.repo.AssignRoleToUser(newUser.ID, DefaultUserRole); errRole != nil {
			fmt.Printf("Failed to assign default role to user: %v\n", errRole)
		}

		u = newUser
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	} else {
		// User exists, update verification status
		u.PhoneVerified = true
		u.Verified = u.EmailVerified // Verified becomes true if email was already verified
		u.LastActive = time.Now()
		if errUpdate := ac.repo.UpdateUser(u); errUpdate != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + errUpdate.Error()})
			return
		}
	}

	accessToken, refreshToken, err := ac.generateAndSaveTokens(c, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         FilterUserRecord(u),
	})
}

// @Summary      Forgot Password
// @Description  Sends a password reset link/code to the user's email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body ForgotPasswordRequest true "Email for password reset"
// @Success      200 {object} map[string]string "Password reset instructions sent"
// @Failure      400 {object} map[string]string "Invalid email format"
// @Failure      404 {object} map[string]string "User with this email not found"
// @Failure      500 {object} map[string]string "Failed to process request"
// @Router       /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	u, err := ac.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	resetToken := utils.GenerateRandomToken(32)   // Ensure this token is cryptographically secure
	resetExpires := time.Now().Add(1 * time.Hour) // Token valid for 1 hour

	u.ResetToken = resetToken
	u.ResetExpires = &resetExpires
	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token: " + err.Error()})
		return
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", ac.config.App.FrontendURL, resetToken)
	emailBody := fmt.Sprintf("Hello %s,\n\nYou requested a password reset. Click the link below to reset your password:\n%s\n\nIf you didn't request this, please ignore this email.\nThis link is valid for 1 hour.", u.Username, resetLink)

	if err := ac.sendEmail(u.Email, "Password Reset Request", emailBody); err != nil {
		fmt.Printf("Failed to send password reset email to %s: %v\n", u.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send password reset email. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset instructions sent to your email."})
}

// @Summary      Reset Password
// @Description  Resets the user's password using a valid reset token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Password reset token and new password"
// @Success      200 {object} map[string]string "Password reset successfully"
// @Failure      400 {object} map[string]string "Invalid input or password mismatch"
// @Failure      401 {object} map[string]string "Invalid or expired reset token"
// @Failure      500 {object} map[string]string "Failed to update password"
// @Router       /auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	u, err := ac.repo.GetUserByResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired password reset token."})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	u.Password = hashedPassword
	u.ResetToken = ""    // Clear the token
	u.ResetExpires = nil // Clear expiry
	u.LastActive = time.Now()

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}

// @Summary      Verify Email
// @Description  Verifies a user's email address using a token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        token query string true "Email verification token"
// @Success      200 {object} map[string]string "Email verified successfully"
// @Failure      400 {object} map[string]string "Invalid or missing token"
// @Failure      401 {object} map[string]string "Invalid or expired token"
// @Failure      500 {object} map[string]string "Failed to verify email"
// @Router       /auth/verify-email [get]
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required."})
		return
	}

	u, err := ac.repo.GetUserByVerifyToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired email verification token."})
		return
	}

	u.EmailVerified = true
	if u.PhoneVerified { // If phone was also verified, main 'Verified' becomes true
		u.Verified = true
	}
	u.VerifyToken = ""
	u.VerifyExpires = nil
	u.LastActive = time.Now()

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email verification status: " + err.Error()})
		return
	}
	// Instead of No Content, maybe redirect to a success page or return a success message
	// c.Redirect(http.StatusFound, ac.config.App.FrontendURL+"/email-verified")
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully."})
}

// @Summary      Resend Verification Email
// @Description  Resends the email verification link to the user.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body ResendVerificationRequest true "Email to resend verification for"
// @Success      200 {object} map[string]string "Verification email resent"
// @Failure      400 {object} map[string]string "Invalid email format"
// @Failure      404 {object} map[string]string "User not found"
// @Failure      409 {object} map[string]string "Email already verified"
// @Failure      500 {object} map[string]string "Failed to resend verification"
// @Router       /auth/resend-verification [post]
func (ac *AuthController) ResendVerificationEmail(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	u, err := ac.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	if u.EmailVerified {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already verified."})
		return
	}

	newVerifyToken := utils.GenerateRandomToken(32)
	newVerifyExpires := time.Now().Add(24 * time.Hour)
	u.VerifyToken = newVerifyToken
	u.VerifyExpires = &newVerifyExpires

	if err := ac.repo.UpdateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification token: " + err.Error()})
		return
	}

	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s", ac.config.App.FrontendURL, newVerifyToken)
	emailBody := fmt.Sprintf("Hello %s, please verify your email address by clicking on this link: %s", u.Username, verificationLink)

	if err := ac.sendEmail(u.Email, "Resend: Verify Your Email Address", emailBody); err != nil {
		fmt.Printf("Failed to resend verification email to %s: %v\n", u.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email has been resent."})
}
