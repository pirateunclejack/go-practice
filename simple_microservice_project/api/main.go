package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var MySigningKey = []byte(os.Getenv("SECRET_KEY"))

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Super Secret Information")
}

func isAuthorized(endpoint func(w http.ResponseWriter,r *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        if r.Header["Token"] != nil {
            token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error){
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("Invalid signning method")
                }

                aud, err := token.Claims.GetAudience()
                if err != nil {
                    return nil, err
                }
                if len(aud) != 1 {
                    return nil, errors.New("wrong token audience length")
                }
                if aud[0] != "billing.jwtgo.io" {
                    return nil, errors.New("wrong token audience")
                }

                iss, err := token.Claims.GetIssuer()
                if err != nil {
                    return nil, err
                }
                if iss != "jwtgo.io" {
                    return nil, errors.New("wrong token issuer")
                }

                issued_at , err := token.Claims.GetExpirationTime()
                if err != nil {
                    return nil, err
                }
                if issued_at.Time.Before(time.Now()) {
                    return nil, errors.New("token expired")
                }

                return MySigningKey, nil
            })

            if err != nil {
                fmt.Fprintf(w, err.Error())
            }

            if token.Valid {
                endpoint(w, r)
            } 
        } else {
            fmt.Fprintf(w, "no authorization token provided")
        }
    })
}

func handleRequests()  {
    http.Handle("/", isAuthorized(homePage))
    log.Fatal(http.ListenAndServe(":9001", nil))
}

func main() {
    fmt.Printf("server started")
    handleRequests()
}
