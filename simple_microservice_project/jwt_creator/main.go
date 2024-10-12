package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var MySigningKey = []byte(os.Getenv("SECRET_KEY"))

type MyCustomClaims struct {
    Authorized bool `json:"authorized"`
    Client string `json:"client"`
    jwt.RegisteredClaims
}

func GetJWT() (string, error){
    claims := MyCustomClaims{
        Authorized: true,
        Client:"jack",
        RegisteredClaims: jwt.RegisteredClaims{
            Audience: []string{"billing.jwtgo.io"},
            Issuer:   "jwtgo.io",
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute*1)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString(MySigningKey)
    if err != nil {
        fmt.Printf("Something went wrong: %v\n", err.Error())
        return "", err
    }
    return tokenString, nil
}

func Index(w http.ResponseWriter, r *http.Request){
    validToken, err := GetJWT()
    fmt.Println(validToken)
    if err != nil {
        fmt.Printf("failed to genrate token: %v\n", err)
    }
    fmt.Fprintf(w, validToken)
}

func HandleRequests() {
    http.HandleFunc("/", Index)
    log.Fatal(http.ListenAndServe(":8888", nil))
}


func main() {
    HandleRequests()
}
