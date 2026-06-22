package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"undangan-digital/internal/models"
)

type BroadcastService struct {
	apiURL     string
	apiKey     string
	appBaseURL string
	client     *http.Client
	lastResult *BroadcastResult
	mu         sync.RWMutex
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
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.apiURL = apiURL
	bs.apiKey = apiKey
	bs.appBaseURL = appBaseURL
}

type BroadcastResult struct {
	Total     int       `json:"total"`
	Success   int       `json:"success"`
	Failed    int       `json:"failed"`
	Errors    []string  `json:"errors,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Running   bool      `json:"running"`
}

func (bs *BroadcastService) GetLastResult() *BroadcastResult {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	if bs.lastResult == nil {
		return &BroadcastResult{}
	}
	result := *bs.lastResult
	return &result
}

func (bs *BroadcastService) Send(guests []models.Guest, message, imageURL string) *BroadcastResult {
	result := &BroadcastResult{
		Total:     len(guests),
		StartTime: time.Now(),
		Running:   true,
	}

	bs.mu.Lock()
	bs.lastResult = result
	bs.mu.Unlock()

	defer func() {
		bs.mu.Lock()
		result.Running = false
		result.EndTime = time.Now()
		bs.lastResult = result
		bs.mu.Unlock()
	}()

	bs.mu.RLock()
	apiURL := bs.apiURL
	apiKey := bs.apiKey
	bs.mu.RUnlock()

	if apiURL == "" || apiKey == "" {
		result.Errors = append(result.Errors, "OneSender API belum dikonfigurasi")
		log.Printf("[broadcast] ERROR: OneSender API belum dikonfigurasi (url=%s, key=%s)", apiURL, apiKey)
		return result
	}

	log.Printf("[broadcast] Mulai mengirim ke %d tamu...", len(guests))

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
				log.Printf("[broadcast] [%d/%d] ✓ %s (%s)", result.Success+result.Failed, result.Total, g.Name, g.PhoneNumber)
			} else {
				result.Failed++
				errStr := fmt.Sprintf("%s (%s): %s", g.Name, g.PhoneNumber, errMsg)
				result.Errors = append(result.Errors, errStr)
				log.Printf("[broadcast] [%d/%d] ✗ %s", result.Success+result.Failed, result.Total, errStr)
			}
			mu.Unlock()
		}(i, guest)
	}

	wg.Wait()
	log.Printf("[broadcast] Selesai: total=%d success=%d failed=%d errors=%d", result.Total, result.Success, result.Failed, len(result.Errors))
	return result
}

func (bs *BroadcastService) getAPIURL() string {
	url := strings.TrimSuffix(bs.apiURL, "/")
	if strings.HasSuffix(url, "/api/v1/messages") {
		return url
	}
	return url + "/api/v1/messages"
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "+", "")

	if phone == "" {
		return phone
	}

	if strings.HasPrefix(phone, "62") {
		return phone
	}
	if strings.HasPrefix(phone, "0") {
		return "62" + phone[1:]
	}
	return "62" + phone
}

func (bs *BroadcastService) sendSingle(guest models.Guest, messageTemplate, imageURL string) (bool, string) {
	if guest.PhoneNumber == "" {
		return false, "nomor telepon kosong"
	}

	phone := normalizePhone(guest.PhoneNumber)
	inviteLink := fmt.Sprintf("%s/undangan/%s", strings.TrimSuffix(bs.appBaseURL, "/"), guest.Slug)

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

	req, err := http.NewRequest("POST", bs.getAPIURL(), bytes.NewBuffer(body))
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

	bodyBytes, _ := io.ReadAll(resp.Body)

	log.Printf("[broadcast] Response from OneSender (status=%d): %s", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode >= 400 {
		return false, fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var respData struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	json.Unmarshal(bodyBytes, &respData)

	if respData.Code != 0 && respData.Code != 200 {
		return false, fmt.Sprintf("OneSender error: %s", respData.Message + respData.Error)
	}

	return true, string(bodyBytes)
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

	phone = normalizePhone(phone)

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

	req, err := http.NewRequest("POST", bs.getAPIURL(), bytes.NewBuffer(body))
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

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return false, fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return true, string(bodyBytes)
}