package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/db"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/server/middleware/myJwt"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/server/templates"
)

func recoverHandler(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        defer func ()  {
            if err := recover(); err != nil {
                log.Panic("recovered! panic: ", err)
                http.Error(w, http.StatusText(500), 500)
            }
        }()
        next.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

func authHandler(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/restricted", "/logout", "/deleteUser":
            log.Println("in auth restricted section")

            // read cookies
            AuthCookie, authErr := r.Cookie("AuthToken")
            if authErr == http.ErrNoCookie {
                log.Println("unauthorized attempt! no auth cookie")
                nullifyTokenCookies(&w, r)
                // http.Redirect(w, r, "/login", 302)
                http.Error(
                    w,
                    http.StatusText(http.StatusUnauthorized),
                    http.StatusUnauthorized)
                return
            } else if authErr != nil {
                log.Panic("panic: ", authErr)
                nullifyTokenCookies(&w, r)
                http.Error(w, http.StatusText(500), 500)
                return
            }

            RefreshCookie, refreshErr := r.Cookie("RefreshToken")
            if refreshErr == http.ErrNoCookie {
                log.Println("unauthorized attempt! no refresh cookie")
                nullifyTokenCookies(&w, r)
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            } else if refreshErr != nil {
                log.Panic("panic: ", refreshErr)
                nullifyTokenCookies(&w, r)
                http.Error(w, http.StatusText(500), 500)
                return
            }

            // grab the csrf token
            requestCsrfToken := grabCsrfFromReq(r)
            log.Println("requestCsrfToken: ", requestCsrfToken)

            // check the jwt's for validity
            authTokenString, 
            refreshTokenString, 
            csrfSecret, 
            err := myJwt.CheckAndRefreshTokens(
                AuthCookie.Value,
                RefreshCookie.Value,
                requestCsrfToken,
            )
            if err != nil {
                if err.Error() == "unauthorized" {
                    log.Println("unauthorized attempt! JWT's not valid!")
                    // nullifyTokenCookies(&w, r)
                    // http.Redirect(w, r, "/login", 302)
                    http.Error(
                        w,
                        http.StatusText(http.StatusUnauthorized),
                        http.StatusUnauthorized)
                    return
                } else {
                    // @adam-hanna: do we 401 or 500, here?
					// it could be 401 bc the token they provided was messed up
					// or it could be 500 bc there was some error on our end
					log.Println("err not nil")
					log.Panic("panic: ", err)
					// nullifyTokenCookies(&w, r)
					http.Error(w, http.StatusText(500), 500)
					return
                }
            }
            log.Println("successfully recreated jwts")

            // @adam-hanna: Change this. Only allow whitelisted origins! Also check referer header
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// if we've made it this far, everything is valid!
			// And tokens have been refreshed if need-be
			setAuthAndRefreshCookies(&w, authTokenString, refreshTokenString)
			w.Header().Set("X-CSRF-Token", csrfSecret)

        default:
            // no jwt check necessary
        }

        next.ServeHTTP(w, r)
    }

    return http.HandlerFunc(fn)
}

func nullifyTokenCookies(w *http.ResponseWriter, r *http.Request)  {
    authCookie := http.Cookie {
        Name: "AuthToken",
        Value: "",
        Expires: time.Now().Add(-1000 * time.Hour),
        HttpOnly: true,
    }
    http.SetCookie(*w, &authCookie)

    refreshCookie := http.Cookie{
        Name: "RefreshToken",
        Value: "",
        Expires: time.Now().Add(-1000*time.Hour),
        HttpOnly: true,
    }
    http.SetCookie(*w, &refreshCookie)

    RefreshCookie, refreshErr := r.Cookie("RefreshToken")
    if refreshErr == http.ErrNoCookie{
        return
    } else if refreshErr != nil {
        log.Panic("panic: ", refreshErr)
        http.Error(*w, http.StatusText(500), 500)
    }

    myJwt.RevokeRefreshToken(RefreshCookie.Value)
}

func setAuthAndRefreshCookies(
    w *http.ResponseWriter,
    authTokenString string,
    refreshTokenString string,
) {
    authCookie := http.Cookie{
        Name: "AuthToken",
        Value: authTokenString,
        HttpOnly: true,
    }
    http.SetCookie(*w, &authCookie)

    refreshCookie := http.Cookie{
        Name: "RefreshToken",
        Value: refreshTokenString,
        HttpOnly: true,
    }

    http.SetCookie(*w, &refreshCookie)
}

func grabCsrfFromReq(r *http.Request) string {
    csrfFromFrom := r.FormValue("X-CSRF-Token")

    if csrfFromFrom != "" {
        return csrfFromFrom
    } else {
        return r.Header.Get("X-CSRF-Token")
    }
}

func logicHandler(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path{
    case "/restricted":
        csrfSecret := grabCsrfFromReq(r)
        templates.RenderTemplate(
            w,
            "restricted",
            &templates.RestrictedPage{
                CsrfSecret: csrfSecret,
                SecretMessage: "Hello Jack"},
        )
    case "/login":
        switch r.Method {
        case "GET":
            templates.RenderTemplate(
                w,
                "login",
                &templates.LoginPage{ 
                    BAlertUser: false,
                    AlertMsg: "" ,
                })
        case "POST":
            r.ParseForm()
			log.Println(r.Form)

			user, uuid, loginErr := db.LogUserIn(
                strings.Join(r.Form["username"], ""),
                strings.Join(r.Form["password"], ""))
			log.Println(user, uuid, loginErr)
			if loginErr != nil {
				// login err
                // templates.RenderTemplate(
                //     w,
                //     "login",
                //     &templates.LoginPage{ 
                //         BAlertUser: true,
                //         AlertMsg: "Login failed\n\nIncorrect username or password" })
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				// no login err
				// now generate cookies for this user
				authTokenString,
                refreshTokenString,
                csrfSecret, err := myJwt.CreateNewTokens(uuid, user.Role)
				if err != nil {
					http.Error(w, http.StatusText(500), 500)
				}

				// set the cookies to these newly created jwt's
				setAuthAndRefreshCookies(&w, authTokenString, refreshTokenString)
				w.Header().Set("X-CSRF-Token", csrfSecret)

				w.WriteHeader(http.StatusOK)
			}
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    case "/register":
        switch r.Method{
        case "GET":
            templates.RenderTemplate(
                w,
                "register",
                &templates.RegisterPage{
                    BAlertUser: false,
                    AlertMsg: "",
                },
            )
        case "POST":
            r.ParseForm()
            log.Println("r.Form: ", r.Form)

            _, _, err := db.FetchUserByUsername(strings.Join(r.Form["username"], ""))
            if err == nil {
                w.WriteHeader(http.StatusUnauthorized)
            } else {
                role := "user"
                uuid, err := db.StoreUser(
                    strings.Join(r.Form["username"], ""),
                    strings.Join(r.Form["password"], ""),
                    role,
                )
                if err != nil {
                    http.Error(w, http.StatusText(500), 500)
                }
                log.Println("uuid: " + uuid)

                authTokenString, refreshTokenString, csrfSecret, err := 
                    myJwt.CreateNewTokens(uuid, role)
                
                if err != nil {
                    http.Error(w, http.StatusText(500), 500)
                }

                setAuthAndRefreshCookies(
                    &w,
                    authTokenString,
                    refreshTokenString,
                )

                w.Header().Set("X-CSRF-Token", csrfSecret)
                w.WriteHeader(http.StatusOK)


            }
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    case "/logout":
        // remove this user's ability to make requests
		nullifyTokenCookies(&w, r)
		// use 302 to force browser to do GET request
		http.Redirect(w, r, "/login", http.StatusFound)
    case "/deleteUser":
		log.Println("Deleting user")

		// grab auth cookie
		AuthCookie, authErr := r.Cookie("AuthToken")
		if authErr == http.ErrNoCookie {
			log.Println("Unauthorized attempt! No auth cookie")
			nullifyTokenCookies(&w, r)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		} else if authErr != nil {
			log.Panic("panic: ", authErr)
			nullifyTokenCookies(&w, r)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		uuid, uuidErr := myJwt.GrabUUID(AuthCookie.Value)
		if uuidErr != nil {
			log.Panic("panic: ", uuidErr)
			nullifyTokenCookies(&w, r)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		db.DeleteUser(uuid)
		// remove this user's ability to make requests
		nullifyTokenCookies(&w, r)
		// use 302 to force browser to do GET request
		http.Redirect(w, r, "/register", http.StatusFound)
    default:
        w.WriteHeader(http.StatusOK)
    }
}

func NewHandler() http.Handler {
    return alice.New(recoverHandler, authHandler).ThenFunc(logicHandler)    
}
