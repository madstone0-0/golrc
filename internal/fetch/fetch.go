package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	// "golrc/internal/errs"
	"golrc/internal/logger"
	"io"
	"strings"
	"time"

	"net/http"
)

var (
	log = logger.NewTaggedLogger("INTERNAL(FETCH)")
)

var DefaultTimeout = 30 * time.Second

type Fetcher interface {
	Get(path string, obj any) error
}

var client *http.Client = &http.Client{
	Timeout: DefaultTimeout,
}

type Fetch struct {
	BaseURL string
}

func cleanURL(url string) string {
	url = strings.TrimSpace(url)
	return url
}

func ParseBody[T any](resp *http.Response, obj *T) (err error) {

	if (399 - resp.StatusCode) < 0 {
		log.D("Error fetching data", "status", resp.Status)
		return fmt.Errorf("error fetching data -> %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}

	return nil
}

func checkFetcher(f *Fetch) error {
	if f == nil {
		return errors.New("fetcher is nil")
	}
	return nil
}

func (f *Fetch) Get(path string, obj any) error {
	err := checkFetcher(f)
	if err != nil {
		return err
	}

	urlPath := fmt.Sprintf("%s%s", f.BaseURL, path)
	urlPath = cleanURL(urlPath)

	resp, err := client.Get(urlPath)
	if err != nil {
		return err
	}

	err = ParseBody(resp, &obj)
	return err
}

func (f *Fetch) Do(req *http.Request, obj any) error {
	err := checkFetcher(f)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = ParseBody(resp, &obj)
	return err
}

func (f *Fetch) Post(path, contentType string, body io.Reader, obj any) error {
	err := checkFetcher(f)
	if err != nil {
		return err
	}

	urlPath := fmt.Sprintf("%s/%s", f.BaseURL, path)
	urlPath = cleanURL(urlPath)

	resp, err := client.Post(urlPath, contentType, body)
	if err != nil {
		return err
	}

	err = ParseBody(resp, &obj)
	return err
}

func (f *Fetch) GetReturn(path string) (string, error) {
	err := checkFetcher(f)
	if err != nil {
		return "", err
	}

	urlPath := fmt.Sprintf("%s/%s", f.BaseURL, path)
	urlPath = cleanURL(urlPath)

	resp, err := client.Get(urlPath)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		// log.Printf("Error fetching data -> %v\n", body)
		return "", fmt.Errorf("error fetching data -> %d", resp.StatusCode)
	}

	return string(body), nil
}

func NewFetcher(baseURL string) Fetch {
	return Fetch{
		BaseURL: baseURL,
	}
}
