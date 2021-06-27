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
	"github.com/stretchr/testify/assert"
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

	tests := []struct {
		in   string
		want model.User
	}{
		{in: `{"email": "alice@example.com", "password": "shhh!"}`, want: model.User{Email: "alice@example.com", Password: "shhh!"}},
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
			if assert.NoError(t, h.Signup(c)) {
				assert.Equal(t, http.StatusCreated, rec.Code)

				got := make(map[string]interface{})
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, got["email"], tt.want.Email)
				assert.Equal(t, got["password"], tt.want.Password)
			}
		})
	}
}
