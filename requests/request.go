package requests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type requestConfig struct {
	method  string
	url     string
	body    io.Reader
	headers map[string]string
	host    string
	cookies []*http.Cookie
}

type RequestOpt = func(*requestConfig) error

type part struct {
	fieldName string
	fileName  string
	data      interface{}
	diskPath  string
}

func Method(method string) RequestOpt {
	return func(config *requestConfig) error {
		config.method = method
		return nil
	}
}

var GET = Method("GET")
var POST = Method("POST")
var PUT = Method("PUT")
var DELETE = Method("DELETE")
var HEAD = Method("HEAD")
var TRACE = Method("TRACE")
var OPTIONS = Method("OPTIONS")
var CONNECT = Method("CONNECT")

func URL(url string) RequestOpt {
	return func(config *requestConfig) error {
		config.url = url
		return nil
	}
}

func Body(body io.Reader) RequestOpt {
	return func(config *requestConfig) error {
		config.body = body
		return nil
	}
}

func JSONBody(body interface{}) RequestOpt {
	return func(config *requestConfig) error {
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)

		err := encoder.Encode(body)
		if err != nil {
			return err
		}

		config.body = buffer
		config.headers["Content-Type"] = "application/json"
		return nil
	}
}

func FormBody(form *url.Values) RequestOpt {
	return func(config *requestConfig) error {
		config.body = strings.NewReader(form.Encode())
		config.headers["Content-Type"] = "application/x-www-form-urlencoded"
		return nil
	}
}

func MultipartForm(parts ...*part) RequestOpt {
	return func(config *requestConfig) error {
		var buffer bytes.Buffer
		w := multipart.NewWriter(&buffer)

		for _, part := range parts {
			var data io.Reader

			if part.data != nil {
				if reader, ok := part.data.(io.Reader); ok {
					data = reader
				} else if b, ok := part.data.([]byte); ok {
					data = bytes.NewReader(b)
				} else if s, ok := part.data.(string); ok {
					data = strings.NewReader(s)
				} else {
					return errors.New("invalid type of data field in multipart form")
				}
			} else if len(part.diskPath) != 0 {
				file, err := os.Open(part.diskPath)
				if err != nil {
					return err
				}

				data = file
				defer file.Close()
			} else {
				return errors.New("no data/diskPath specified for mutlipart form")
			}

			fileWriter, err := w.CreateFormFile(part.fieldName, part.fileName)
			if err != nil {
				return err
			}

			_, err = io.Copy(fileWriter, data)
			if err != nil {
				return err
			}
		}

		if err := w.Close(); err != nil {
			return err
		}

		config.body = &buffer
		config.headers["Content-Type"] = w.FormDataContentType()
		return nil
	}
}

func Header(key, value string) RequestOpt {
	return func(config *requestConfig) error {
		config.headers[key] = value
		return nil
	}
}

func BearerToken(token string) RequestOpt {
	return Header("Authorization", fmt.Sprintf("Bearer %s", token))
}

func BasicAuth(username, password string) RequestOpt {
	authString := fmt.Sprintf("%s:%s", username, password)
	return Header("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString))))
}

func UserAgent(userAgent string) RequestOpt {
	return Header("User-Agent", userAgent)
}

func ContentType(contentType string) RequestOpt {
	return Header("Content-Type", contentType)
}

func RandomRequestID() RequestOpt {
	return Header("X-Request-Id", uuid.NewV4().String())
}

func Host(host string) RequestOpt {
	return func(config *requestConfig) error {
		config.host = host
		return nil
	}
}

func Cookie(cookie *http.Cookie) RequestOpt {
	return func(config *requestConfig) error {
		config.cookies = append(config.cookies, cookie)
		return nil
	}
}

func NewRequest(opts ...RequestOpt) (*http.Request, error) {
	config := &requestConfig{
		method:  "GET",
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(config.method, config.url, config.body)
	if err != nil {
		return nil, err
	}

	for header, value := range config.headers {
		request.Header.Set(header, value)
	}

	if len(config.host) != 0 {
		request.Host = config.host
	}

	for _, cookie := range config.cookies {
		request.AddCookie(cookie)
	}

	return request, nil
}

func GetRequestID(request *http.Request) string {
	return request.Header.Get("X-Request-Id")
}

func PartFromData(fieldName, fileName string, data interface{}) *part {
	return &part{
		fieldName: fieldName,
		fileName:  fileName,
		data:      data,
	}
}

func PartFromDiskFile(fieldName, fileName, diskPath string) *part {
	return &part{
		fieldName: fieldName,
		fileName:  fileName,
		diskPath:  diskPath,
	}
}
