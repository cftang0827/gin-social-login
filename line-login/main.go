package main

//localhost:8080/line-login
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/unrolled/secure"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

var clientID = os.Getenv("LineClientID")
var callBackURL = "https://localhost:8080/callback"
var clientSecret = os.Getenv("LineClientSecret")

func tlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:8080",
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}
func main() {
	r := gin.Default()
	r.Use(tlsHandler())

	r.GET("/line-login", func(c *gin.Context) {
		state := "test"
		url := fmt.Sprintf("https://access.line.me/oauth2/v2.1/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s&scope=%s",
			clientID,
			callBackURL,
			state,
			"email%20openid%20profile",
		)
		c.Redirect(http.StatusMovedPermanently, url)
	})

	r.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		log.Println("Fetching data from token api")

		payload, _ := json.Marshal(map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  callBackURL,
			"client_id":     clientID,
			"client_secret": clientSecret,
		})
		request := gorequest.New()
		_, body, _ := request.Post("https://api.line.me/oauth2/v2.1/token").
			Set("Content-Type", "application/x-www-form-urlencoded").
			Send(string(payload)).
			End()
		var bodyjSON map[string]interface{}
		json.Unmarshal([]byte(body), &bodyjSON)
		log.Println("Already get the access token")
		// accessToken := bodyjSON["access_token"]
		idToken := bodyjSON["id_token"].(string)

		token, _ := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return "test", nil
		})

		name := token.Claims.(jwt.MapClaims)["name"]
		img := token.Claims.(jwt.MapClaims)["picture"]
		email := token.Claims.(jwt.MapClaims)["email"]
		c.JSON(http.StatusOK, gin.H{
			"name":  name,
			"img":   img,
			"email": email,
		})

	})

	r.RunTLS(":8080", "localhost.crt", "localhost.key") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
