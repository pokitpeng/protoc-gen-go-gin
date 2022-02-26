package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"
	v1 "github.com/pengpeng/protoc-gen-go-gin/example/api/product/app/v1"
	"github.com/pengpeng/protoc-gen-go-gin/example/api/product/ecode"
)

type service struct {
}

func (s service) CreateArticle(ctx context.Context, article *v1.Article) (*v1.Article, error) {
	if article.AuthorId < 1 {
		// return nil, errors.New("author id must > 0")
		return nil, ecode.Errorf(500, 500, "author id must > 0")
	}
	return article, nil
}

func (s service) GetArticles(ctx context.Context, req *v1.GetArticlesReq) (*v1.GetArticlesResp, error) {
	if req.AuthorId < 0 {
		// return nil, errors.New("author id must >= 0")
		return nil, ecode.Errorf(500, 500, "author id must > 0")
	}
	return &v1.GetArticlesResp{
		Total: 1,
		Articles: []*v1.Article{
			{
				Title:    "test article: " + req.Title,
				Content:  "test",
				AuthorId: 1,
			},
		},
	}, nil
}

func prog(state overseer.State) {
	router, err := NewRouter()
	if err != nil {
		panic(err)
	}
	_ = http.Serve(state.Listener, router)
}

func main() {
	overseer.Run(overseer.Config{
		Program: prog,
		Address: "0.0.0.0:8080",
		Fetcher: &fetcher.HTTP{
			URL:      "http://http.org/head",
			Interval: 60 * time.Second,
		},
	})
}

func RegisterBlogService(s *v1.BlogService) {
	s.Router.GET("/v1/author/:author_id/articles", s.GetArticles)
}

func RegisterBlogServiceHTTPServer(r gin.IRouter, srv v1.BlogServiceHTTPServer) {
	s := v1.BlogService{
		Server: srv,
		Router: r,
		Resp:   v1.DefaultBlogServiceResp{},
	}
	RegisterBlogService(&s)
}

func RegisterBaseHandlers(router *gin.Engine) {
	// version
	router.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"version": "v1.0.0"})
	})

	// ping
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})
}

func NewRouter() (*gin.Engine, error) {
	// gin init
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	// router.Use(ginzap.Ginzap(logger.GetLogger(), time.RFC3339, false))
	// router.Use(ginzap.RecoveryWithZap(logger.GetLogger(), true))
	// no route/method
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no route"})
	})
	router.HandleMethodNotAllowed = true
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "no method"})
	})

	RegisterBaseHandlers(router)
	RegisterBlogServiceHTTPServer(router, &service{})

	return router, nil
}
