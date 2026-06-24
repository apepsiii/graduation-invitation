package models

import (
	"database/sql"
	"time"
)

const (
	RSVPBelumKonfirmasi = "Belum Konfirmasi"
	RSVPHadir           = "Hadir"
	RSVPTidakHadir      = "Tidak Hadir"
	RSVPRagu            = "Ragu"
	RSVPPending         = "pending"
)

var ValidRSVPStatuses = []string{RSVPHadir, RSVPTidakHadir, RSVPRagu, RSVPBelumKonfirmasi, RSVPPending}

type Guest struct {
	ID          int64        `db:"id" form:"id"`
	Slug        string       `db:"slug" form:"slug"`
	Name        string       `db:"name" form:"name" binding:"required"`
	PhoneNumber string       `db:"phone_number" form:"phone_number"`
	Kelas       string       `db:"kelas" form:"kelas"`
	QRToken     string       `db:"qr_token" form:"qr_token"`
	RSVPStatus  string       `db:"rsvp_status" form:"rsvp_status"`
	IsAttended  bool         `db:"is_attended" form:"is_attended"`
	AttendedAt  sql.NullTime `db:"attended_at" form:"attended_at"`
	MealTakenAt sql.NullTime `db:"meal_taken_at" form:"meal_taken_at"`
	CreatedAt   time.Time    `db:"created_at" form:"created_at"`
}

type MealStats struct {
	TotalGuests int64 `json:"total_guests"`
	MealTaken   int64 `json:"meal_taken"`
	MealRemain  int64 `json:"meal_remain"`
}

type MealCheckin struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Kelas       string    `json:"kelas"`
	MealTakenAt time.Time `json:"meal_taken_at"`
}

type AttendanceCheckin struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Kelas       string    `json:"kelas"`
	AttendedAt  time.Time `json:"attended_at"`
}

type EventSettings struct {
	ID                int64  `db:"id" form:"id"`
	EventTitle        string `db:"event_title" form:"event_title"`
	EventDate         string `db:"event_date" form:"event_date"`
	EventTime         string `db:"event_time" form:"event_time"`
	VenueName         string `db:"venue_name" form:"venue_name"`
	VenueAddress      string `db:"venue_address" form:"venue_address"`
	MapsLink          string `db:"maps_link" form:"maps_link"`
	Dresscode         string `db:"dresscode" form:"dresscode"`
	OneSenderURL      string `db:"onesender_url" form:"onesender_url"`
	OneSenderAPIKey   string `db:"onesender_api_key" form:"onesender_api_key"`
	AppBaseURL        string `db:"app_base_url" form:"app_base_url"`
	BroadcastTemplate string `db:"broadcast_template" form:"broadcast_template"`
	BroadcastImageURL string `db:"broadcast_image_url" form:"broadcast_image_url"`
}

type Guestbook struct {
	ID        int64     `db:"id" form:"id"`
	GuestID   int64     `db:"guest_id" form:"guest_id"`
	Message   string    `db:"message" form:"message"`
	CreatedAt time.Time `db:"created_at" form:"created_at"`
}

type GuestbookWithGuest struct {
	ID        int64     `json:"id"`
	GuestID   int64     `json:"guest_id"`
	GuestName string    `json:"guest_name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Rundown struct {
	ID           int64  `db:"id" form:"id"`
	StartTime    string `db:"start_time" form:"start_time" binding:"required"`
	EndTime      string `db:"end_time" form:"end_time" binding:"required"`
	ActivityName string `db:"activity_name" form:"activity_name" binding:"required"`
	Description  string `db:"description" form:"description"`
}

type Gallery struct {
	ID        int64  `db:"id" form:"id"`
	ImageURL  string `db:"image_url" form:"image_url" binding:"required"`
	Caption   string `db:"caption" form:"caption"`
	SortOrder int    `db:"sort_order" form:"sort_order"`
}

type RSVPRequest struct {
	GuestID int64  `json:"guest_id" form:"guest_id" binding:"required"`
	Status  string `json:"status" form:"status" binding:"required"`
	Message string `json:"message" form:"message"`
}

type ScanRequest struct {
	QRToken string `json:"qr_token" form:"qr_token" binding:"required"`
}

type BroadcastRequest struct {
	Message  string `json:"message" form:"message"`
	ImageURL string `json:"image_url" form:"image_url"`
	GuestIDs []int64 `json:"guest_ids" form:"guest_ids"`
	Type     string `json:"type" form:"type"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type Stats struct {
	TotalGuests   int64 `json:"total_guests"`
	TotalRSVP     int64 `json:"total_rsvp"`
	TotalAttended int64 `json:"total_attended"`
}

func IsValidRSVPStatus(status string) bool {
	for _, s := range ValidRSVPStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func NormalizeRSVPStatus(status string) string {
	switch status {
	case "Hadir":
		return RSVPHadir
	case "Tidak Hadir":
		return RSVPTidakHadir
	case "Ragu":
		return RSVPRagu
	case RSVPBelumKonfirmasi, RSVPPending:
		return RSVPBelumKonfirmasi
	default:
		return RSVPBelumKonfirmasi
	}
}
