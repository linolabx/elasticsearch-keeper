package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/linolabx/lino_redis"
	"github.com/rs/zerolog"
)

type ServerContext struct {
	Logger        *zerolog.Logger
	Redis         *lino_redis.LinoRedis
	Es            *elasticsearch.Client
	EsIndexPrefix string

	SynonymsConfigDir string
	ApiKey            string

	GinAddr string
	GinEng  *gin.Engine

	DevMode bool
}

func RunServer(ctx *ServerContext) {
	// Start Gin Server
	ginServer := gin.New()
	ginServer.Use(gin.Recovery())
	ginServer.Use(gin.Logger())
	ginServer.MaxMultipartMemory = 128 << 20 // 128 MiB

	ctx.GinEng = ginServer
	defer ginServer.Run(ctx.GinAddr)

	// Update Synonym
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	ctx.GinEng.POST("/synonyms/:name", func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer "+ctx.ApiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		name := c.Param("name")
		if !nameRegex.MatchString(name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid synonym name, only allow [a-zA-Z0-9_-]"})
			return
		}

		concurrencyGuard := ctx.Redis.Fork("synonym_lock").NewHeartBeatLock(name, 2*time.Second)
		defer concurrencyGuard.Del(context.TODO())
		if err := concurrencyGuard.TryLock(context.TODO()); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("another update is in progress: %s", name)})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid multipart form"})
			return
		}

		file := form.File["file"][0]
		if file == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, file is required"})
			return
		}

		mFile, err := file.Open()
		if err != nil {
			ctx.Logger.Error().Err(err).Msg("load uploaded file failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load uploaded file failed"})
			return
		}

		body, err := io.ReadAll(mFile)
		if err != nil {
			ctx.Logger.Error().Err(err).Msg("read uploaded file failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read uploaded file failed"})
			return
		}

		// TODO: validate synonym syntax

		filePath := filepath.Join(ctx.SynonymsConfigDir, name)
		tempFilePath := filePath + ".tmp"

		if err := os.WriteFile(tempFilePath, body, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write synonym file failed"})
			return
		}

		if err := os.Rename(tempFilePath, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rename synonym file failed"})
			return
		}

		ctx.Logger.Info().Msgf("%s updated, new hash: %s", filePath, fmt.Sprintf("%x", md5.Sum(body)))

		indexes := form.Value["indexes[]"]
		if len(indexes) == 0 {
			ctx.Logger.Info().Msg("no indexes specified, skip reloading")
			c.JSON(http.StatusOK, gin.H{"message": "synonym set"})
			return
		}

		reloadResp, err := ctx.Es.Indices.ReloadSearchAnalyzers(indexes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "reload search analyzers failed"})
			return
		}

		if reloadResp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("reload search analyzers failed, status code: %s", reloadResp.Status())})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "synonym set, analyzer reloaded"})
	})
}
