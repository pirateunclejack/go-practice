package db

import (
	"errors"
	"log"

	"github.com/pirateunclejack/go-practice/golang-csrf-project/db/models"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/randomstrings"
	"golang.org/x/crypto/bcrypt"
)

var users = map[string]models.User{}
var refreshTokens map[string]string

func InitDB() {
    refreshTokens = make(map[string]string)
}


func StoreUser(username, password, role string) (uuid string, err error) {
    uuid, err = randomstrings.GenerateRandomString(32)
    if err != nil {
        return "", err
    }

    u := models.User{};
    for u != users[uuid]{
        uuid, err = randomstrings.GenerateRandomString(32)
        if err != nil {
            return "", err
        }
    }

    PasswordHash, hashErr := generateBcryptHash(password)
    if hashErr != nil {
        err = hashErr
        return
    }
    users[uuid] = models.User{
        Username: username,
        PasswordHash: PasswordHash,
        Role: role,
    }
    return uuid, err
}


func DeleteUser(uuid string) {
    delete(users, uuid)
}

func FetchUserByID(uuid string) (models.User, error) {
    u := users[uuid]
    blankUser := models.User{}
    if blankUser != u {
        return u, nil
    } else {
        return u, errors.New("user not found thant matches the given uuid")
    }
}

func FetchUserByUsername(username string) (models.User, string, error) {
    for k, v := range users {
        if v.Username == username {
            return v, k, nil
        }
    }
    return models.User{},
        "",
        errors.New("user not found that matches the given username")
}

func StoreRefreshToken() (jti string, err error)  {
    jti, err = randomstrings.GenerateRandomString(32)
    if err != nil {
        return jti, err
    }

    for refreshTokens[jti] != "" {
        jti, err = randomstrings.GenerateRandomString(32)
        if err != nil {
            return jti, err
        }
    }

    refreshTokens[jti] = "valid"
    return jti, err
}

func DeleteRefreshToken(jti string) {
    delete(refreshTokens, jti)
}

func CheckRefreshToken(jti string) bool {
    return refreshTokens[jti] != ""
}

func LogUserIn(username, password string) (models.User, string, error) {
    user, uuid, userErr := FetchUserByUsername(username)
    log.Println("loguserin :", user, uuid, userErr)
    if userErr != nil {
        return models.User{}, "", userErr
    }

    return user, uuid, checkPasswordAgainstHash(user.PasswordHash, password)
}

func generateBcryptHash(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func checkPasswordAgainstHash(hash string, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
