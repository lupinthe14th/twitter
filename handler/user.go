package handler

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/lupinthe14th/twitter/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const TokenExpHour = 72

func (h *Handler) Signup(c echo.Context) error {
	// Bind
	u := &model.User{ID: bson.NewObjectId()}
	if err := c.Bind(u); err != nil {
		return err
	}

	// Validate
	if u.Email == "" || u.Password == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid email or password"}
	}

	// Save user
	db := h.DB.Clone()
	defer db.Close()
	if err := db.DB("twitter").C("users").Insert(u); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) error {
	// Bind
	u := new(model.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	// Find user
	db := h.DB.Clone()
	defer db.Close()
	if err := db.DB("twitter").C("users").Find(bson.M{"email": u.Email, "password": u.Password}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
		}
		return err
	}

	// ---
	// JWT
	// ---

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["exp"] = time.Now().Add(time.Hour * TokenExpHour).Unix()

	// Genarate encoded token and send it as response
	var err error
	u.Token, err = token.SignedString([]byte(Key))
	if err != nil {
		return err
	}

	u.Password = "" // Don't send password
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) Follow(c echo.Context) error {
	userID := userIDFromToken(c)
	id := c.Param("id")

	// Add a follow to user
	db := h.DB.Clone()
	defer db.Close()
	if err := db.DB("twitter").C("users").UpdateId(bson.ObjectIdHex(id), bson.M{"$addToSet": bson.M{"followers": userID}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}
	return nil
}

func userIDFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
