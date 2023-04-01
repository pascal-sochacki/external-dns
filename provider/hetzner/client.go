/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type hetznerAPI interface {
	GetAllZones() (*AllZones, error)
	GetAllRecords(params GetAllRecordsParams) (*AllRecords, error)
	DeleteRecord(id string) error
	CreateRecord(request CreateRecordParams) error
}

type AllZones struct {
	Zones []Zone `json:"zones"`
}

type Zone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AllRecords struct {
	Records []Record `json:"records"`
}

type Record struct {
	Type     string    `json:"type"`
	Id       string    `json:"id"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	ZoneId   string    `json:"zone_id"`
	Name     string    `json:"buildRequest"`
	Value    string    `json:"value"`
	Ttl      int       `json:"ttl"`
}

type hetznerClient struct {
	httpClient *http.Client
	token      string
}

type GetAllRecordsParams struct {
	page *int
	zone string
}

type CreateRecordParams struct {
	ZoneId string `json:"zone_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Ttl    *int   `json:"ttl"`
}

func (h hetznerClient) buildRequest(url string, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Auth-API-Token", h.token)

	return req, err
}

// GetAllRecords TODO: pages
func (h hetznerClient) GetAllZones() (*AllZones, error) {
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/zones")
	req, err := h.buildRequest(url, "GET", nil)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var zones AllZones
	err = json.Unmarshal(body, zones)
	if err != nil {
		return nil, err
	}

	return &zones, nil
}

// GetAllRecords TODO: pages
func (h hetznerClient) GetAllRecords(params GetAllRecordsParams) (*AllRecords, error) {
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records?zone_id=%s", params.zone)
	req, err := h.buildRequest(url, "GET", nil)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var records AllRecords
	err = json.Unmarshal(body, records)
	if err != nil {
		return nil, err
	}

	return &records, nil
}

func (h hetznerClient) DeleteRecord(id string) error {
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records?zone_id=%s", id)
	req, err := h.buildRequest(url, "GET", nil)
	if err != nil {
		return err
	}
	_, err = h.httpClient.Do(req)
	return err
}

func (h hetznerClient) CreateRecord(request CreateRecordParams) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records")
	req, err := h.buildRequest(url, "POST", bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func newHetznerClient(token string) *hetznerClient {
	return &hetznerClient{
		httpClient: &http.Client{},
		token:      token,
	}

}
