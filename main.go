package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

// DB
var (
	// uid: name
	users = map[string]string{}

	comments []comment
)

type comment struct {
	uid      string
	comment  string
	postedAt time.Time
}

func main() {
	jsonKeyFilePath := flag.String("j", "", "Required. JSON key file of the Firebase Admin SDK.")
	flag.Parse()

	ctx := context.Background()
	opt := option.WithCredentialsFile(*jsonKeyFilePath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatal(err)
	}

	b := bbs{authClient: client}

	r := gin.Default()
	r.LoadHTMLGlob("views/*")

	store := cookie.NewStore([]byte("cookie_secret_key"))
	r.Use(sessions.Sessions("session", store))

	authGroup := r.Group("/", authentication)
	authGroup.POST("/comments", postComment)
	authGroup.GET("/comments", getComments)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", nil)
	})
	r.POST("/signup", b.signup)
	r.POST("/signin", b.signin)

	if err := r.Run(":80"); err != nil {
		log.Fatal(err)
	}
}

func authentication(c *gin.Context) {
	sess := sessions.Default(c)
	uid := sess.Get("uid")
	if uid == nil {
		log.Print("uid is nil")
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	name := sess.Get("name")
	if name == nil {
		log.Print("name is nil")
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	c.Set("uid", uid)
	c.Set("name", name)
}

func getComments(c *gin.Context) {
	fmt.Println(c.GetString("uid"))
}

func postComment(c *gin.Context) {
	fmt.Println(c.GetString("uid"))
}

type bbs struct {
	authClient *auth.Client
}

func (b *bbs) signup(c *gin.Context) {
	bs, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Abort()
		log.Print(err)
	}

	token, err := b.authClient.VerifyIDToken(c, string(bs))
	if err != nil {
		c.Abort()
		log.Print(err)
	}

	sess := sessions.Default(c)
	sess.Set("uid", token.UID)
	//sess.Set("name", token.UID)
	sess.Save()

	c.JSON(http.StatusCreated, nil)
}

func (b *bbs) signin(c *gin.Context) {
	bs, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	token, err := b.authClient.VerifyIDToken(c, string(bs))
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	sess := sessions.Default(c)
	sess.Set("uid", token.UID)
	//sess.Set("name", name)
	sess.Save()

	c.JSON(http.StatusCreated, nil)
}
