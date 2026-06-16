package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"undangan-digital/internal/models"
)

type BroadcastService struct {
	apiURL    string
	apiKey    string
	appBaseURL string
	client    *http.Client
}

func NewBroadcastService(apiURL, apiKey, appBaseURL string) *BroadcastService {
	return &BroadcastService{
		apiURL:     apiURL,
		apiKey:     apiKey,
		appBaseURL: appBaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (bs *BroadcastService) UpdateConfig(apiURL, apiKey, appBaseURL string) {
	bs.apiURL = apiURL
	bs.apiKey = apiKey
	bs.appBaseURL = appBaseURL
}

type BroadcastResult struct {
	Total   int
	Success int
	Failed  int
	Errors  []string
}

func (bs *BroadcastService) Send(guests []models.Guest, message, imageURL string) *BroadcastResult {
	result := &BroadcastResult{Total: len(guests)}

	if bs.apiURL == "" || bs.apiKey == "" {
		result.Errors = append(result.Errors, "OneSender API belum dikonfigurasi")
		return result
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i, guest := range guests {
		wg.Add(1)
		go func(idx int, g models.Guest) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ok, errMsg := bs.sendSingle(g, message, imageURL)
			mu.Lock()
			if ok {
				result.Success++
			} else {
				result.Failed++
				if errMsg != "" {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", g.Name, errMsg))
				}
			}
			mu.Unlock()
		}(i, guest)
	}

	wg.Wait()
	log.Printf("[broadcast] Done: total=%d success=%d failed=%d", result.Total, result.Success, result.Failed)
	return result
}

func (bs *BroadcastService) sendSingle(guest models.Guest, messageTemplate, imageURL string) (bool, string) {
	if guest.PhoneNumber == "" {
		return false, "nomor telepon kosong"
	}

	phone := guest.PhoneNumber
	if !strings.HasPrefix(phone, "62") {
		phone = "62" + strings.TrimPrefix(phone, "0")
		phone = strings.TrimPrefix(phone, "62")
		phone = "62" + phone
	}

	inviteLink := fmt.Sprintf("%s/undangan/%s", bs.appBaseURL, guest.Slug)

	var payload map[string]interface{}

	if imageURL != "" {
		personalCaption := bs.personalize(messageTemplate, guest.Name, inviteLink)
		payload = map[string]interface{}{
			"recipient_type": "individual",
			"to":             phone,
			"type":           "image",
			"image": map[string]string{
				"link":    imageURL,
				"caption": personalCaption,
			},
		}
	} else {
		personalMessage := bs.personalize(messageTemplate, guest.Name, inviteLink)
		payload = map[string]interface{}{
			"recipient_type": "individual",
			"to":             phone,
			"type":           "text",
			"text": map[string]string{
				"body": personalMessage,
			},
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Sprintf("marshal error: %v", err)
	}

	req, err := http.NewRequest("POST", bs.apiURL+"/api/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Sprintf("request error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bs.apiKey)

	resp, err := bs.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("send error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return false, fmt.Sprintf("API error %d: %v", resp.StatusCode, errResp)
	}

	return true, ""
}

func (bs *BroadcastService) personalize(template, name, link string) string {
	if template == "" {
		template = "Halo {nama},\n\nKamu diundang ke acara perpisahan kami!\n\nDetail undangan: {link}\n\nTerima kasih!"
	}
	result := strings.ReplaceAll(template, "{nama}", name)
	result = strings.ReplaceAll(result, "{link}", link)
	return result
}

func (bs *BroadcastService) SendTest(phone, message, imageURL string) (bool, string) {
	if bs.apiURL == "" || bs.apiKey == "" {
		return false, "OneSender API belum dikonfigurasi"
	}

	if !strings.HasPrefix(phone, "62") {
		phone = "62" + strings.TrimPrefix(phone, "0")
		phone = strings.TrimPrefix(phone, "62")
		phone = "62" + phone
	}

	var payload map[string]interface{}

	if imageURL != "" {
		payload = map[string]interface{}{
			"recipient_type": "individual",
			"to":             phone,
			"type":           "image",
			"image": map[string]string{
				"link":    imageURL,
				"caption": message,
			},
		}
	} else {
		payload = map[string]interface{}{
			"recipient_type": "individual",
			"to":             phone,
			"type":           "text",
			"text": map[string]string{
				"body": message,
			},
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Sprintf("marshal error: %v", err)
	}

	req, err := http.NewRequest("POST", bs.apiURL+"/api/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Sprintf("request error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bs.apiKey)

	resp, err := bs.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("send error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return false, fmt.Sprintf("API error %d: %v", resp.StatusCode, errResp)
	}

	return true, ""
}