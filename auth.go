package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func AuthSales(c *gin.Context) {
	ck, err := c.Request.Cookie("token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	cl, err1 := parseToken(ck.Value)
	if err1 != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if (uint32(cl.(jwt.MapClaims)["acc"].(float64)) & UserSales) != UserSales {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set("usr_id", cl.(jwt.MapClaims)["usr"])
	c.Next()
}

func AuthAdmin(c *gin.Context) {
	ck, err := c.Request.Cookie("token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	cl, err1 := parseToken(ck.Value)
	if err1 != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if (uint32(cl.(jwt.MapClaims)["acc"].(float64)) & UserAdmin) != UserAdmin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set("usr_id", cl.(jwt.MapClaims)["usr"])
	c.Next()
}

func createToken(u *User) string {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"usr": u.Id,
		"des": u.Description,
		"acc": u.Access,
	}
	tokenString, err := token.SignedString([]byte(keySecret))
	if err != nil {
		return ""
	}
	return tokenString
}

func parseToken(t string) (jwt.Claims, error) {
	tk, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		return []byte(keySecret), nil
	})
	if err == nil && tk.Valid {
		return tk.Claims, nil
	}
	return nil, err
}

func login(c *gin.Context) {
	userid, _ := c.GetQuery("userid")
	password, _ := c.GetQuery("password")
	flg, _ := c.GetQuery("flag")
	flag, _ := strconv.ParseBool(flg)

	usr := User{}
	err := DB.Get(&usr, "select * from user where id=? and password=?", userid, password)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	tk := createToken(&usr)
	ck := http.Cookie{
		Name:     "token",
		Value:    tk,
		HttpOnly: true,
	}
	if flag {
		ck.Expires = time.Now().Add(365 * 24 * time.Hour)
	}
	http.SetCookie(c.Writer, &ck)
}

func logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}
