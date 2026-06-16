package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"undangan-digital/internal/middleware"
	"undangan-digital/internal/models"
	"undangan-digital/internal/repository"
)

type Handler struct {
	repo        *repository.Repository
	session     *middleware.SessionManager
	authUser    string
	authPass    string
	broadcast   *BroadcastService
}

func NewHandler(repo *repository.Repository, session *middleware.SessionManager, authUser, authPass string, broadcastService *BroadcastService) *Handler {
	return &Handler{
		repo:     repo,
		session:  session,
		authUser: authUser,
		authPass: authPass,
		broadcast: broadcastService,
	}
}

func (h *Handler) HomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title": "Undangan Digital",
	})
}

func (h *Handler) GetInvitation(c *gin.Context) {
	slug := c.Param("slug")

	guest, err := h.repo.GetGuestBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "Not Found"})
			return
		}
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"title": "Error"})
		return
	}

	settings, err := h.repo.GetEventSettings()
	if err != nil {
		settings = &models.EventSettings{}
	}

	rundowns, _ := h.repo.GetRundowns()
	galleries, _ := h.repo.GetGalleries()
	guestbooks, _ := h.repo.GetGuestbookByGuestID(guest.ID)

	initial := "N"
	if len(guest.Name) > 0 {
		initial = guest.Name[:1]
	}

	c.HTML(http.StatusOK, "invitation.html", gin.H{
		"title":      "Undangan untuk " + guest.Name,
		"guest":      guest,
		"settings":   settings,
		"rundowns":   rundowns,
		"galleries":  galleries,
		"guestbooks": guestbooks,
		"initial":    initial,
	})
}

func (h *Handler) PostRSVP(c *gin.Context) {
	var req models.RSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	status := models.NormalizeRSVPStatus(req.Status)
	if err := h.repo.UpdateRSVP(req.GuestID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update RSVP"})
		return
	}

	if req.Message != "" {
		if err := h.repo.InsertGuestbook(req.GuestID, req.Message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) GetPublicGuestbooks(c *gin.Context) {
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	guestbooks, total, err := h.repo.GetGuestbooksPaginated(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load guestbooks"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"comments":    guestbooks,
		"page":        page,
		"total_pages": totalPages,
		"total":       total,
	})
}

func (h *Handler) GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/login.html", gin.H{
		"title": "Login Admin",
	})
}

func (h *Handler) PostLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		if c.GetHeader("Accept") == "application/json" || c.GetHeader("Content-Type") == "application/json" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		c.HTML(http.StatusUnauthorized, "admin/login.html", gin.H{
			"title": "Login Admin",
			"error": "Username dan password harus diisi",
		})
		return
	}

	if req.Username != h.authUser || req.Password != h.authPass {
		if c.GetHeader("Accept") == "application/json" || c.GetHeader("Content-Type") == "application/json" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.HTML(http.StatusUnauthorized, "admin/login.html", gin.H{
			"title": "Login Admin",
			"error": "Username atau password salah",
		})
		return
	}

	h.session.SetSession(c, req.Username)

	if c.GetHeader("Accept") == "application/json" || c.GetHeader("Content-Type") == "application/json" {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (h *Handler) PostLogout(c *gin.Context) {
	h.session.ClearSession(c)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (h *Handler) PostScan(c *gin.Context) {
	var req models.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	guest, err := h.repo.GetGuestByQRToken(req.QRToken)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "QR Code tidak valid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Server error"})
		return
	}

	if guest.IsAttended {
		c.JSON(http.StatusOK, gin.H{
			"status":     "already",
			"guest_name": guest.Name,
			"message":    "Tamu sudah melakukan check-in",
		})
		return
	}

	if err := h.repo.MarkAttended(req.QRToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to mark attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"guest_name": guest.Name,
		"message":    "Check-in berhasil",
	})
}

func (h *Handler) GetAdminDashboard(c *gin.Context) {
	stats, err := h.repo.GetStats()
	if err != nil {
		stats = &models.Stats{}
	}

	guests, err := h.repo.GetAllGuests()
	if err != nil {
		guests = []models.Guest{}
	}

	settings, err := h.repo.GetEventSettings()
	if err != nil {
		settings = &models.EventSettings{}
	}

	guestbooks, _ := h.repo.GetGuestbooks()

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"title":      "Dashboard Admin",
		"stats":      stats,
		"guests":     guests,
		"settings":   settings,
		"guestbooks": guestbooks,
	})
}

func (h *Handler) GetAdminGuests(c *gin.Context) {
	guests, err := h.repo.GetAllGuests()
	if err != nil {
		guests = []models.Guest{}
	}

	c.HTML(http.StatusOK, "admin/guests.html", gin.H{
		"title":  "Manajemen Tamu",
		"guests": guests,
	})
}

func (h *Handler) GetAdminSettings(c *gin.Context) {
	settings, err := h.repo.GetEventSettings()
	if err != nil {
		settings = &models.EventSettings{}
	}

	c.HTML(http.StatusOK, "admin/settings.html", gin.H{
		"title":    "Pengaturan Acara",
		"settings": settings,
	})
}

func (h *Handler) PostAdminSettings(c *gin.Context) {
	existing, err := h.repo.GetEventSettings()
	if err != nil {
		existing = &models.EventSettings{}
	}

	var input models.EventSettings
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if input.EventTitle != "" {
		existing.EventTitle = input.EventTitle
	}
	if input.EventDate != "" {
		existing.EventDate = input.EventDate
	}
	if input.EventTime != "" {
		existing.EventTime = input.EventTime
	}
	if input.VenueName != "" {
		existing.VenueName = input.VenueName
	}
	if input.VenueAddress != "" {
		existing.VenueAddress = input.VenueAddress
	}
	if input.MapsLink != "" {
		existing.MapsLink = input.MapsLink
	}
	if input.Dresscode != "" {
		existing.Dresscode = input.Dresscode
	}
	if input.OneSenderURL != "" {
		existing.OneSenderURL = input.OneSenderURL
	}
	if input.OneSenderAPIKey != "" {
		existing.OneSenderAPIKey = input.OneSenderAPIKey
	}
	if input.AppBaseURL != "" {
		existing.AppBaseURL = input.AppBaseURL
	}

	existing.ID = 1

	if err := h.repo.UpdateEventSettings(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) GetAdminGuestbooksPage(c *gin.Context) {
	guestbooks, _, err := h.repo.GetGuestbooksPaginated(1, 100)
	if err != nil {
		guestbooks = []models.GuestbookWithGuest{}
	}

	c.HTML(http.StatusOK, "admin/guestbooks.html", gin.H{
		"title":      "Pesan Tamu",
		"guestbooks": guestbooks,
	})
}

func (h *Handler) DeleteAdminGuestbook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repo.DeleteGuestbook(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete guestbook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) PostAdminGuests(c *gin.Context) {
	var guest models.Guest
	if err := c.ShouldBind(&guest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	guest.PhoneNumber = normalizePhone(guest.PhoneNumber)

	if guest.RSVPStatus == "" {
		guest.RSVPStatus = models.RSVPBelumKonfirmasi
	}

	if err := h.repo.CreateGuest(&guest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "guest": guest})
}

func (h *Handler) PutAdminGuest(c *gin.Context) {
	var guest models.Guest
	if err := c.ShouldBind(&guest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	guest.PhoneNumber = normalizePhone(guest.PhoneNumber)

	if err := h.repo.UpdateGuest(&guest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update guest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) DeleteAdminGuest(c *gin.Context) {
	id := c.Param("id")
	var guestID int64
	if _, err := fmt.Sscanf(id, "%d", &guestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repo.DeleteGuest(guestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete guest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) PostAdminImportGuests(c *gin.Context) {
	file, _, err := c.Request.FormFile("csv_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header and at least one data row"})
		return
	}

	var guests []models.Guest
	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}
		guest := models.Guest{
			Name:        strings.TrimSpace(record[0]),
			PhoneNumber: strings.TrimSpace(record[1]),
			RSVPStatus:  models.RSVPBelumKonfirmasi,
		}
		if len(record) > 2 {
			guest.Slug = strings.TrimSpace(record[2])
		}
		guests = append(guests, guest)
	}

	if len(guests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid guests found in CSV"})
		return
	}

	if err := h.repo.ImportGuests(guests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import guests: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Successfully imported %d guests", len(guests)),
		"count":   len(guests),
	})
}

func (h *Handler) GetAdminScanner(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/scanner.html", gin.H{
		"title": "QR Scanner",
	})
}

func (h *Handler) GetStatsAPI(c *gin.Context) {
	stats, err := h.repo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetAdminRundowns(c *gin.Context) {
	rundowns, err := h.repo.GetRundowns()
	if err != nil {
		rundowns = []models.Rundown{}
	}
	c.JSON(http.StatusOK, rundowns)
}

func (h *Handler) PostAdminRundown(c *gin.Context) {
	var rundown models.Rundown
	if err := c.ShouldBind(&rundown); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.repo.CreateRundown(&rundown); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rundown"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "rundown": rundown})
}

func (h *Handler) DeleteAdminRundown(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repo.DeleteRundown(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rundown"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) GetAdminGalleries(c *gin.Context) {
	galleries, err := h.repo.GetGalleries()
	if err != nil {
		galleries = []models.Gallery{}
	}
	c.JSON(http.StatusOK, galleries)
}

func (h *Handler) PostAdminGallery(c *gin.Context) {
	caption := c.PostForm("caption")
	sortStr := c.PostForm("sort_order")
	sortOrder := 0
	if sortStr != "" {
		sortOrder, _ = strconv.Atoi(sortStr)
	}

	file, err := c.FormFile("image")
	if err != nil {
		imageURL := c.PostForm("image_url")
		if imageURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image file or URL required"})
			return
		}
		gallery := models.Gallery{
			ImageURL:  imageURL,
			Caption:   caption,
			SortOrder: sortOrder,
		}
		if err := h.repo.CreateGallery(&gallery); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gallery"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "gallery": gallery})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Use JPG, PNG, WebP, or GIF"})
		return
	}

	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Max 10MB"})
		return
	}

	uploadDir := "assets/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	filename := hex.EncodeToString(randBytes) + ext
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepath.ToSlash(filePath)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	gallery := models.Gallery{
		ImageURL:  "/assets/uploads/" + filename,
		Caption:   caption,
		SortOrder: sortOrder,
	}
	if err := h.repo.CreateGallery(&gallery); err != nil {
		os.Remove(filepath.ToSlash(filePath))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "gallery": gallery})
}

func (h *Handler) DeleteAdminGallery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repo.DeleteGallery(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) PostAdminBroadcast(c *gin.Context) {
	var req models.BroadcastRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	settings, err := h.repo.GetEventSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	h.broadcast.UpdateConfig(settings.OneSenderURL, settings.OneSenderAPIKey, settings.AppBaseURL)

	var guests []models.Guest
	if len(req.GuestIDs) > 0 {
		for _, id := range req.GuestIDs {
			g, err := h.repo.GetGuestByID(id)
			if err == nil {
				guests = append(guests, *g)
			}
		}
	} else {
		allGuests, err := h.repo.GetAllGuests()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get guests"})
			return
		}
		guests = allGuests
	}

	if len(guests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No guests to broadcast"})
		return
	}

	guestCopy := make([]models.Guest, len(guests))
	copy(guestCopy, guests)

	broadcast := h.broadcast
	go func() {
		broadcast.Send(guestCopy, req.Message, req.ImageURL)
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Broadcast dimulai untuk %d tamu", len(guests)),
		"count":   len(guests),
	})
}

func (h *Handler) PostAdminBroadcastTest(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" form:"phone"`
		Message  string `json:"message" form:"message"`
		ImageURL string `json:"image_url" form:"image_url"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	settings, err := h.repo.GetEventSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	h.broadcast.UpdateConfig(settings.OneSenderURL, settings.OneSenderAPIKey, settings.AppBaseURL)

	ok, errMsg := h.broadcast.SendTest(req.Phone, req.Message, req.ImageURL)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Test broadcast berhasil"})
}

func (h *Handler) NotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"title": "Halaman Tidak Ditemukan",
	})
}

func (h *Handler) GetAdminRundownsPage(c *gin.Context) {
	rundowns, err := h.repo.GetRundowns()
	if err != nil {
		rundowns = []models.Rundown{}
	}
	c.HTML(http.StatusOK, "admin/rundowns.html", gin.H{
		"title":    "Manajemen Rundown",
		"rundowns": rundowns,
	})
}

func (h *Handler) GetAdminGalleriesPage(c *gin.Context) {
	galleries, err := h.repo.GetGalleries()
	if err != nil {
		galleries = []models.Gallery{}
	}
	c.HTML(http.StatusOK, "admin/galleries.html", gin.H{
		"title":     "Manajemen Galeri",
		"galleries": galleries,
	})
}
