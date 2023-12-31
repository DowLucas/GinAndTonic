package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const YOUR_CALLBACK_URL = "http://localhost:8080/login-complete"

var db *gorm.DB
var jwtKey []byte

func init() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	jwtKey = []byte(os.Getenv("JWT_KEY"))

	db, err = database.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	db.AutoMigrate(&models.User{})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// Logout logic
	// Remove the cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		// Secure: true, // Uncomment this line if you are using HTTPS
		// Domain: "yourfrontenddomain.com", // Set your domain here
	})

	c.Redirect(http.StatusSeeOther, "/")
}

// Login redirects to the external login page
func LoginPostman(c *gin.Context) {
	loginURL := os.Getenv("LOGIN_BASE_URL") + "/login?callback=" + "http://localhost:8080/postman-login-complete/"

	c.Redirect(http.StatusSeeOther, loginURL)
}

func Login(c *gin.Context) {
	loginURL := os.Getenv("LOGIN_BASE_URL") + "/login?callback=" + "http://localhost:8080/login-complete/"

	c.JSON(http.StatusOK, gin.H{
		"login_url": loginURL,
	})
}

func CurrentUser(c *gin.Context) {
	// Get the user from the context
	UGKthID := c.MustGet("ugkthid").(string)

	// Get the user from the database
	user, err := models.GetUserByUGKthIDIfExist(db, UGKthID)

	if err != nil {
		// Remove the cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1,
			// Secure: true, // Uncomment this line if you are using HTTPS
			// Domain: "yourfrontenddomain.com", // Set your domain here
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// LoginComplete handles the callback from the external login system
func LoginComplete(c *gin.Context) {
	token := c.Param("token")
	client := &http.Client{}

	verificationURL := os.Getenv("LOGIN_BASE_URL") + "/verify/" + token + ".json?api_key=" + os.Getenv("LOGIN_API_KEY")

	req, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating request",
		})
		return
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api_key", os.Getenv("LOGIN_API_KEY"))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error sending request",
		})
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var body types.Body
		decoder := json.NewDecoder(res.Body)
		err := decoder.Decode(&body)
		if err != nil {
			println("Error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error decoding response",
			})
			return
		}

		// Check if user exists in database
		var user models.User
		user, err = models.GetUserByUGKthIDIfExist(db, body.UGKthID)
		if err == nil {
			// User exists
			tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				HttpOnly: true,
				Path:     "/",
				Secure:   false, // Uncomment this line if you are using HTTPS
				// Domain: "yourfrontenddomain.com", // Set your domain here
			})

			c.Redirect(http.StatusSeeOther, os.Getenv("FRONTEND_BASE_URL")+"?auth=success")

			return
		}

		role, err := models.GetRole(db, "user")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching role",
			})
			return
		}

		user = models.User{
			Username:  body.User,
			FirstName: body.FirstName,
			LastName:  body.LastName,
			Email:     body.Emails,
			UGKthID:   body.UGKthID,
			RoleID:    role.ID,
			Role:      role,
		}

		tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = models.CreateUserIfNotExist(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user food preference",
			})
			return
		}

		// Set the JWT token in an HTTP-only cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			HttpOnly: true,
			Path:     "/",
			// Secure: true, // Uncomment this line if you are using HTTPS
			// Domain: "yourfrontenddomain.com", // Set your domain here
		})

		c.Redirect(http.StatusSeeOther, os.Getenv("FRONTEND_BASE_URL")+"?auth=success")
	} else {
		println("Error: " + res.Status)
		http.Redirect(c.Writer, c.Request, os.Getenv("FRONTEND_BASE_URL")+"?auth=failed", http.StatusSeeOther)
	}
}

func LoginCompletePostman(c *gin.Context) {
	token := c.Param("token")
	client := &http.Client{}

	verificationURL := os.Getenv("LOGIN_BASE_URL") + "/verify/" + token + ".json?api_key=" + os.Getenv("LOGIN_API_KEY")

	req, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating request",
		})
		return
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api_key", os.Getenv("LOGIN_API_KEY"))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error sending request",
		})
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var body types.Body
		decoder := json.NewDecoder(res.Body)
		err := decoder.Decode(&body)
		if err != nil {
			println("Error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error decoding response",
			})
			return
		}

		// Check if user exists in database
		var user models.User
		user, err = models.GetUserByUGKthIDIfExist(db, body.UGKthID)
		if err == nil {
			// User exists
			tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				HttpOnly: true,
				Path:     "/",
				// Secure: true, // Uncomment this line if you are using HTTPS
				// Domain: "yourfrontenddomain.com", // Set your domain here
			})

			c.JSON(http.StatusOK, gin.H{
				"token": tokenString,
				"user":  user,
			})

			return
		}

		role, err := models.GetRole(db, "user")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching role",
			})
			return
		}

		user = models.User{
			Username:  body.User,
			FirstName: body.FirstName,
			LastName:  body.LastName,
			Email:     body.Emails,
			UGKthID:   body.UGKthID,
			RoleID:    role.ID,
			Role:      role,
		}

		tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = models.CreateUserIfNotExist(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user food preference",
			})
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			HttpOnly: true,
			Path:     "/",
			// Secure: true, // Uncomment this line if you are using HTTPS
			// Domain: "yourfrontenddomain.com", // Set your domain here
		})

		// Set the JWT token in an HTTP-only cookie
		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user":  user,
		})

	} else {
		println("Error: " + res.Status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error verifying token",
		})
	}
}
