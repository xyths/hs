package jsonapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
)

func GetWithHeader(ctx context.Context, url string, header map[string]string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range header {
		req.Header.Add(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, response); err != nil {
		return err
	}
	return nil
}

func PostWithForm(ctx context.Context, url string, header, data map[string]string, response interface{}) error {
	data_ := urlpkg.Values{}
	for k, v := range data {
		data_.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data_.Encode()))
	if err != nil {
		return err
	}
	for k, v := range header {
		req.Header.Add(k, v)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, response); err != nil {
		return err
	}
	return nil
}
