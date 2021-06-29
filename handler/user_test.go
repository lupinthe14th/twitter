package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/lupinthe14th/twitter/model"
	"gopkg.in/mgo.v2"
)

var (
	h = &Handler{}
)

func TestMain(m *testing.M) {
	if err := setUp(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	status := m.Run()
	if err := tearDown(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(status)
}

func setUp() error {
	// setup using test database
	db, err := mgo.Dial(os.Getenv("MONGO_INITDB"))
	if err != nil {
		return err
	}

	cred := mgo.Credential{
		Username: os.Getenv("MONGO_INITDB_ROOT_USERNAME"),
		Password: os.Getenv("MONGO_INITDB_ROOT_PASSWORD"),
	}
	if err := db.Login(&cred); err != nil {
		return err
	}

	// Create indices
	if err := db.Copy().DB("twitter").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}); err != nil {
		return err
	}
	h = &Handler{DB: db}
	return nil
}

func tearDown() error {
	defer h.DB.Close()
	// drop test database
	if err := h.DB.DB("twitter").DropDatabase(); err != nil {
		return err
	}
	return nil
}

func TestSignup(t *testing.T) {
	t.Parallel()

	type want struct {
		user model.User
		code int
	}

	tests := []struct {
		in   string
		want want
	}{
		{in: `{"email": "alice@example.com", "password": "shhh!"}`, want: want{user: model.User{Email: "alice@example.com", Password: "shhh!"}, code: http.StatusCreated}},
		{in: `{"email": "", "password": "shhh!"}`, want: want{code: http.StatusBadRequest}},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(tt.in))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			if err := h.Signup(c); err != nil {
				he, ok := err.(*echo.HTTPError)
				if ok {
					if he.Code != tt.want.code {
						t.Errorf("in: %v got: %v want: %v", tt.in, rec, tt.want)
					}
				}
			} else {
				if rec.Code != tt.want.code {
					t.Errorf("in: %v rec: %v want: %v", tt.in, rec, tt.want)
				}
				got := make(map[string]interface{})
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				if !(got["email"] == tt.want.user.Email && got["password"] == tt.want.user.Password) {
					t.Errorf("in: %v got: %v want: %v", tt.in, got, tt.want)
				}
			}
		})
	}
}
