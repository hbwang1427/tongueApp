package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

//GetWorkDirectory get work directory
func GetWorkDirectory() string {
	//设置当前目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

//SearchPath search given file or dir name `fname` upward the directory tree until found
func SearchPath(start string, fname string) (string, bool) {
	var fp string
	for len(start) >= 2 {
		fp = filepath.Join(start, fname)
		if _, err := os.Stat(fp); err == nil {
			return fp, true
		}
		start = filepath.Dir(start)
	}
	return "", false
}

//GetToken get request token
func GetToken(c *gin.Context) string {
	authTokenName := AppConfig().Token.AuthName
	if token := c.GetHeader(authTokenName); len(token) > 0 {
		return token
	}
	for _, c := range c.Request.Cookies() {
		if c.Name == authTokenName {
			return c.Value
		}
	}
	if token := c.PostForm(authTokenName); len(token) > 0 {
		return token
	}
	return c.Query(authTokenName)
}

//GetParamInt returns true if name exists and can be convert to int else false
func GetParamInt(c *gin.Context, name string) (int64, bool) {
	v, found := GetParam(c, name)
	if !found {
		return 0, false
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

//GetParam returns true if name exists else false
func GetParam(c *gin.Context, name string) (string, bool) {
	v, found := c.GetPostForm(name)
	if found {
		return v, true
	}
	return c.GetQuery(name)
}

//GetClientIP get client ip
func GetClientIP(req *http.Request) string {
	//http header is case insensitive
	ip := req.Header.Get("X-Real-IP")
	if len(ip) > 0 {
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			return ip
		}
		return host
	}

	forward := req.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		ips := strings.Split(forward, ",")
		if len(ips) > 0 {
			host, _, err := net.SplitHostPort(ips[0])
			if err != nil {
				return ips[0]
			}
			return host
		}
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

//CreateRequest ...
func CreateRequest(method string, url string, data url.Values) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(data.Encode()))
	if strings.ToUpper(method) == "POST" || strings.ToUpper(method) == "PUT" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	}
	return req
}

type errResponse struct {
	Code int
	Msg  string
}

//ExtractResponse extract response body
func ExtractResponse(t *testing.T, resp *http.Response, out interface{}) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	err = json.Unmarshal(body, &out)
	require.Nil(t, err, string(body))
}

//ResponseExpectErr ...
func ResponseExpectErr(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var errResp errResponse
	ExtractResponse(t, resp, &errResp)
	require.NotEqual(t, errResp.Code, 0)
}

//ResponseExpect expect code
func ResponseExpect(t *testing.T, resp *http.Response, code int) {
	ResponseExpectIn(t, resp, []int{code})
}

//ResponseExpectIn expect code in
func ResponseExpectIn(t *testing.T, resp *http.Response, codes []int) {
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var errResp errResponse
	ExtractResponse(t, resp, &errResp)
	require.Contains(t, codes, errResp.Code)
	//assert.Equal(t, code, int(c.(float64)), bodyMap)
}
