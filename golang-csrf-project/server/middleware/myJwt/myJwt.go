package myJwt

import (
	"crypto/rsa"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/db"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/db/models"
)

const (
    privKeyPath = "keys/app.rsa"
    pubKeyPath = "keys/app.rsa.pub"
)

var signKey *rsa.PrivateKey
var verifyKey *rsa.PublicKey

func InitJWT() error {
    signBytes, err := os.ReadFile(privKeyPath)
    if err != nil {
        return err
    }

    signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
    if err != nil {
        return err
    }

    verifyBytes, err := os.ReadFile(pubKeyPath)
    if err != nil {
        return err
    }

    verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
    if err != nil {
        return err
    }

    return nil
}

func CreateNewTokens(
    uuid string, role string,
) (authTokenString, refreshTokenString, csrfSecret string, err error) {
    // generate the csrf secret
    csrfSecret, err = models.GenerateCSRFSecret()
    if err != nil {
        return
    }

    // generating the refresh token
    refreshTokenString, err = createRefreshTokenString(uuid, role, csrfSecret)
    if err != nil {
        return
    }

    // generate the auth token
    authTokenString, err = createAuthTokenString(uuid, role, csrfSecret)
    if err != nil {
        return
    }

    return
}

func CheckAndRefreshTokens(
    oldAuthTokenString string,
    oldRefreshTokenString string,
    oldCsrfSecret string,
) (newAuthTokenString, newRefreshTokenString, newCsrfSecret string, err error) {
    if oldCsrfSecret == "" {
        log.Println("no csrf token!")
        err = errors.New("unauthorized")
        return
    }
    authToken, err := jwt.ParseWithClaims(
        oldAuthTokenString,
        &models.TokenClaims{},
        func(token *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        })
    authTokenClaims, ok := authToken.Claims.(*models.TokenClaims)
    if !ok {
        return
    }
    if oldCsrfSecret != authTokenClaims.Csrf {
        log.Println("CSRF token doesn't match jwt!")
        err = errors.New("unauthorized")
    }

    if authToken.Valid {
        log.Println("auth token is valid")
        newCsrfSecret = authTokenClaims.Csrf
        newRefreshTokenString, err = updateRefreshTokenExp(oldRefreshTokenString)
        newAuthTokenString = oldAuthTokenString

        return
    } else if ve, ok := err.(*jwt.ValidationError); ok {
        log.Println("auth token is not valid")
        if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
            log.Println("auth token is expired")
            newAuthTokenString, newCsrfSecret, err = updateAuthTokenString(
                oldRefreshTokenString,
                oldAuthTokenString)
            if err != nil {
                return
            }

            newRefreshTokenString, err = updateRefreshTokenExp(oldRefreshTokenString)
            if err != nil {
                return
            }

            newRefreshTokenString, err = updateRefreshTokenCsrf(
                newRefreshTokenString,
                newCsrfSecret,
            )

        } else {
            log.Println("error in auth token")
            err = errors.New("error in auth token")
        }
    } else {
        log.Println("error in auth token")
        err = errors.New("error in auth token")
        return
    }

    err = errors.New("unauthorized")
    return
}

func createAuthTokenString(
    uuid, role, csrfSecret string,
) (authTokenString string, err error) {
    authTokenExp := jwt.NewNumericDate(time.Now().Add(models.AuthTokenValidTime))
    authClaims := models.TokenClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject: uuid,
            ExpiresAt: authTokenExp,
        },
        Role: role,
        Csrf: csrfSecret,
    }

    authJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), authClaims)
    authTokenString, err = authJwt.SignedString(signKey)

    return

}

func createRefreshTokenString(
    uuid, role, csrfString string,
) (refreshTokenString string, err error) {
    refreshTokenExp := jwt.NewNumericDate(time.Now().Add(models.RefreshTokenValidTime))
    refreshJti, err := db.StoreRefreshToken()
    if err != nil {
        return
    }

    refreshClaims := models.TokenClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID: refreshJti,
            Subject: uuid,
            ExpiresAt: refreshTokenExp,
        },
        Role: role,
        Csrf: csrfString,
    }
    
    refreshJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), refreshClaims)
    refreshTokenString, err = refreshJwt.SignedString(signKey)
    return
}

func updateRefreshTokenExp(
    oldRefreshTokenString string,
) (newRefreshTokenString string, err error) {
    refreshToken, err := jwt.ParseWithClaims(
        oldRefreshTokenString,
        &models.TokenClaims{},
        func(token *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        },
    )
    oldRefreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
    if !ok {
        return
    }

    refreTokenExp := jwt.NewNumericDate(time.Now().Add(models.RefreshTokenValidTime))
    refreshClaims := models.TokenClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID: oldRefreshTokenClaims.RegisteredClaims.ID,
            Subject: oldRefreshTokenClaims.RegisteredClaims.Subject,
            ExpiresAt: refreTokenExp,
        },
        Role: oldRefreshTokenClaims.Role,
        Csrf: oldRefreshTokenClaims.Csrf,
    }

    refreshJwt := jwt.NewWithClaims(
        jwt.GetSigningMethod("RS256"),
        refreshClaims,
    )
    newRefreshTokenString, err = refreshJwt.SignedString(signKey)
    return
}

func updateAuthTokenString(
    refreshTokenString string,
    oldAuthTokenString string,
) (
    newAuthTokenString string,
    csrfSecret string,
    err error,
) {
    refreshToken, err := jwt.ParseWithClaims(
        refreshTokenString,
        &models.TokenClaims{},
        func(token *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        },
    )

    refreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
    if !ok {
        err = errors.New("error reading jwt claims")
        return
    }

    if db.CheckRefreshToken(refreshTokenClaims.RegisteredClaims.ID) {
        if refreshToken.Valid {
            authToken, _ := jwt.ParseWithClaims(
                oldAuthTokenString,
                &models.TokenClaims{},
                func(token *jwt.Token) (interface{}, error) {
                    return verifyKey, nil
                },
            )

            oldAuthTokenClaims, ok :=  authToken.Claims.(*models.TokenClaims)
            if !ok {
                err = errors.New("error reading jwt claims")
                return
            }
            csrfSecret, err = models.GenerateCSRFSecret()
            if err != nil {
                return
            }

            createAuthTokenString(
                oldAuthTokenClaims.RegisteredClaims.Subject,
                oldAuthTokenClaims.Role,
                csrfSecret,
            )

            return
        } else {
            log.Println("refresh token has expired")
            db.DeleteRefreshToken(refreshTokenClaims.RegisteredClaims.ID)

            err = errors.New("unauthorized")
            return
        }
    } else {
        log.Println("refresh token has been revoked")
        err = errors.New("unauthorized")
        return
    }
}

func RevokeRefreshToken(refreshTokenString string) error {
    // use the refresh token string that this function will receive to get your
    // refresh token
    refreshToken, err := jwt.ParseWithClaims(
        refreshTokenString,
        &models.TokenClaims{},
        func(t *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        },
    )
    if err != nil {
        return errors.New("could not parse refresh token with claims")
    }

    // use the refresh token to get the refresh token claims
    refreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
    if !ok {
        return errors.New("could not read refresh token claims")
    }

    // deleting the refresh token using the method in the db package
    db.DeleteRefreshToken(refreshTokenClaims.RegisteredClaims.ID)

    return nil
}

func updateRefreshTokenCsrf(
    oldRefreshTokenString string,
    newCsrfString string,
) (newRefreshTokenString string, err error) {
    // get access to the old refresh token by using the parsewithclaims function
    refreshToken, err := jwt.ParseWithClaims(
        oldRefreshTokenString,
        &models.TokenClaims{},
        func(t *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        },
    )

    // get access to the refresh token claims
    oldRefreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
    if !ok {
        return
    }

    // refreshclaims
    refreshClaims := models.TokenClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID: oldRefreshTokenClaims.RegisteredClaims.ID,
            Subject: oldRefreshTokenClaims.RegisteredClaims.Subject,
            ExpiresAt: oldRefreshTokenClaims.RegisteredClaims.ExpiresAt,
        },
        Role: oldRefreshTokenClaims.Role,
        Csrf: newCsrfString,
    }

    // new refresh jwt
    refreshJwt := jwt.NewWithClaims(
        jwt.GetSigningMethod("RS256"),
        refreshClaims,
    )

    // new token string
    newRefreshTokenString, err = refreshJwt.SignedString(signKey)
    return
}

func GrabUUID(authTokenString string) (string, error) {
    authToken, _ := jwt.ParseWithClaims(
        authTokenString,
        &models.TokenClaims{},
        func(token *jwt.Token) (interface{}, error) {
            return "", errors.New("error fetching claims")
        },
    )
    authTokenClaims, ok := authToken.Claims.(*models.TokenClaims)
    if !ok {
        return "", errors.New("error fetching claims")
    }

    return authTokenClaims.RegisteredClaims.Subject, nil
}
