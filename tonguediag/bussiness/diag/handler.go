package diag

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"tonguediag/utils"

	_ "github.com/mattn/go-sqlite3" //for sqlite3

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var config *utils.Config
var dbx *sqlx.DB
var logger *zap.Logger

//Init diag module init
func Init(c *utils.Config, r *gin.Engine) {
	config = c
	logger = utils.Logger(c)
	r.POST("/diag/checkuser", checkUserHandler)
	r.POST("/diag/upload", uploadHandler)
	r.GET("/diag/images", getImagesHandler)
	r.POST("/diag/settag", setImageTagHandler)
	r.GET("/diag/tags", getTagsHandler)

	r.LoadHTMLGlob("assets/templates/*")
	r.GET("/tag", func(c *gin.Context) {
		c.HTML(http.StatusOK, "tongue.html", gin.H{
			"title": "Main website",
		})
	})

	r.GET("/apk", func(c *gin.Context) {
		c.HTML(http.StatusOK, "apk.html", gin.H{})
	})

	r.Static("/diag/img", "./upload")

	var err error
	dbx, err = sqlx.Connect("sqlite3", fmt.Sprintf("file:%s", c.SqliteDB))
	if err != nil {
		logger.Fatal("open database error", zap.Error(err))
	}

	schema := `
	create table if not exists user (
		id INTEGER primary key,
		name text not null unique
	);

	create table if not exists uploads (
		id INTEGER primary key,
		user_id INTEGER,
		path text,
		tags text
	)
	`

	_, err = dbx.Exec(schema)
	if err != nil {
		logger.Fatal("init db schema error", zap.Error(err))
	}
}

//curl -X POST -F name=kingwang -v http://localhost:8080/diag/checkuser
func checkUserHandler(c *gin.Context) {
	var args = struct {
		Name string `form:"name" json:"name"`
	}{}
	if err := c.ShouldBind(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "'name' required",
		})
		return
	}
	var userID int64
	err := dbx.QueryRowx(`select id from user where name=?`, args.Name).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			result, err := dbx.Exec(`insert into user(name) values(?)`, args.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				logger.Error("insert into user error:", zap.Error(err))
				return
			}
			userID, _ = result.LastInsertId()
		} else {
			c.JSON(http.StatusInternalServerError, nil)
			logger.Error("get user error:", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"userID": userID,
	})
}

//image upload handler
//curl -v -F userID=1 -F image=@/Users/yu/Downloads/Snow.png http://localhost:8080/diag/upload
func uploadHandler(c *gin.Context) {
	file, _ := c.FormFile("image")
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "open upload image error",
		})
		return
	}
	userID, err := strconv.Atoi(c.PostForm("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "parameter name required",
		})
		return
	}
	uploadDir := filepath.Join(config.UploadDir, fmt.Sprintf("%d", userID))
	if err = os.MkdirAll(uploadDir, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "create directory for uploaded image failed",
		})
		return
	}
	ext := filepath.Ext(file.Filename)
	now := time.Now().UnixNano()
	path := fmt.Sprintf("/diag/img/%d/%d%s", userID, now, ext)
	contents, _ := ioutil.ReadAll(f)
	savePath := filepath.Join(uploadDir, fmt.Sprintf("%d%s", now, ext))
	if err = ioutil.WriteFile(savePath, contents, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "write upload image error",
		})
		return
	}

	result, err := dbx.Exec(`insert into uploads(user_id, path, tags) values(?,?, '')`, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "add record to database error",
		})
		logger.Error("insert uploads error", zap.Error(err))
		return
	}

	uploadID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"uploadID": uploadID,
		"path":     path,
	})
}

//get image list handler
//curl -v http://localhost:8080/diag/images
func getImagesHandler(c *gin.Context) {
	var images = []struct {
		ID     int    `db:"id" json:"id"`
		UserID int    `db:"user_id" json:"user_id"`
		Path   string `db:"path" json:"path"`
		Tags   string `db:"tags" json:"tags"`
	}{}
	err := dbx.Select(&images, `select * from uploads`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		logger.Error("getImagesHandler error", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})
}

//set image tag
func setImageTagHandler(c *gin.Context) {
	var args = struct {
		ID   int    `form:"id" json:"id"`
		Tags string `form:"tags" json:"tags"`
	}{}
	if err := c.ShouldBind(&args); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	_, err := dbx.Exec(`update uploads set tags=? where id=?`, args.Tags, args.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		logger.Error("update tags error", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func getTagsHandler(c *gin.Context) {
	tags, err := ioutil.ReadFile("./tags.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.Writer.Write(tags)
}
