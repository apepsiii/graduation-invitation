package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"undangan-digital/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS event_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_title TEXT NOT NULL DEFAULT '',
		event_date TEXT NOT NULL DEFAULT '',
		event_time TEXT NOT NULL DEFAULT '',
		venue_name TEXT NOT NULL DEFAULT '',
		venue_address TEXT NOT NULL DEFAULT '',
		maps_link TEXT NOT NULL DEFAULT '',
		dresscode TEXT NOT NULL DEFAULT ''
	);	CREATE TABLE IF NOT EXISTS guests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		phone_number TEXT NOT NULL DEFAULT '',
		qr_token TEXT NOT NULL UNIQUE,
		rsvp_status TEXT NOT NULL DEFAULT 'pending',
		is_attended INTEGER NOT NULL DEFAULT 0,
		attended_at TEXT,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS guestbooks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guest_id INTEGER NOT NULL,
		message TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (guest_id) REFERENCES guests(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS rundowns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		activity_name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS galleries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_url TEXT NOT NULL,
		caption TEXT NOT NULL DEFAULT '',
		sort_order INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_guests_slug ON guests(slug);
	CREATE INDEX IF NOT EXISTS idx_guests_qr_token ON guests(qr_token);
	CREATE INDEX IF NOT EXISTS idx_guestbooks_guest_id ON guestbooks(guest_id);
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	r.migrateSchema()

	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM event_settings").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check event_settings: %w", err)
	}
	if count == 0 {
		_, err = r.db.Exec("INSERT INTO event_settings (event_title, event_date, event_time, venue_name, venue_address, maps_link, dresscode) VALUES ('', '', '', '', '', '', '')")
		if err != nil {
			return fmt.Errorf("failed to insert default event_settings: %w", err)
		}
	}

	_, err = r.db.Exec(`
		DELETE FROM guestbooks WHERE id NOT IN (
			SELECT MAX(id) FROM guestbooks GROUP BY guest_id
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to deduplicate guestbooks: %w", err)
	}

	return nil
}

func (r *Repository) migrateSchema() {
	migrations := []string{
		"ALTER TABLE event_settings ADD COLUMN onesender_url TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE event_settings ADD COLUMN onesender_api_key TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE event_settings ADD COLUMN app_base_url TEXT NOT NULL DEFAULT ''",
	}
	for _, m := range migrations {
		r.db.Exec(m)
	}
}

func (r *Repository) GetEventSettings() (*models.EventSettings, error) {
	settings := &models.EventSettings{}
	err := r.db.QueryRow(`
		SELECT id, event_title, event_date, event_time, venue_name, venue_address, maps_link, dresscode, onesender_url, onesender_api_key, app_base_url
		FROM event_settings LIMIT 1
	`	).Scan(
		&settings.ID, &settings.EventTitle, &settings.EventDate, &settings.EventTime,
		&settings.VenueName, &settings.VenueAddress, &settings.MapsLink, &settings.Dresscode,
		&settings.OneSenderURL, &settings.OneSenderAPIKey, &settings.AppBaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get event settings: %w", err)
	}
	return settings, nil
}

func (r *Repository) UpdateEventSettings(settings *models.EventSettings) error {
	_, err := r.db.Exec(`
		UPDATE event_settings
		SET event_title = ?, event_date = ?, event_time = ?, venue_name = ?, venue_address = ?, maps_link = ?, dresscode = ?, onesender_url = ?, onesender_api_key = ?, app_base_url = ?
		WHERE id = ?
	`, settings.EventTitle, settings.EventDate, settings.EventTime, settings.VenueName,
		settings.VenueAddress, settings.MapsLink, settings.Dresscode,
		settings.OneSenderURL, settings.OneSenderAPIKey, settings.AppBaseURL, settings.ID)
	if err != nil {
		return fmt.Errorf("failed to update event settings: %w", err)
	}
	return nil
}

func (r *Repository) GetGuestBySlug(slug string) (*models.Guest, error) {
	guest := &models.Guest{}
	var attendedAt, createdAt sql.NullString
	err := r.db.QueryRow(`
		SELECT id, slug, name, phone_number, qr_token, rsvp_status, is_attended, attended_at, created_at
		FROM guests WHERE slug = ?
	`, slug).Scan(
		&guest.ID, &guest.Slug, &guest.Name, &guest.PhoneNumber, &guest.QRToken,
		&guest.RSVPStatus, &guest.IsAttended, &attendedAt, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get guest by slug: %w", err)
	}
	if attendedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", attendedAt.String)
		guest.AttendedAt = sql.NullTime{Time: t, Valid: true}
	}
	if createdAt.Valid {
		guest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	return guest, nil
}

func (r *Repository) GetGuestByID(id int64) (*models.Guest, error) {
	guest := &models.Guest{}
	var attendedAt, createdAt sql.NullString
	err := r.db.QueryRow(`
		SELECT id, slug, name, phone_number, qr_token, rsvp_status, is_attended, attended_at, created_at
		FROM guests WHERE id = ?
	`, id).Scan(
		&guest.ID, &guest.Slug, &guest.Name, &guest.PhoneNumber, &guest.QRToken,
		&guest.RSVPStatus, &guest.IsAttended, &attendedAt, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get guest by id: %w", err)
	}
	if attendedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", attendedAt.String)
		guest.AttendedAt = sql.NullTime{Time: t, Valid: true}
	}
	if createdAt.Valid {
		guest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	return guest, nil
}

func (r *Repository) GetGuestByQRToken(qrToken string) (*models.Guest, error) {
	guest := &models.Guest{}
	var attendedAt, createdAt sql.NullString
	err := r.db.QueryRow(`
		SELECT id, slug, name, phone_number, qr_token, rsvp_status, is_attended, attended_at, created_at
		FROM guests WHERE qr_token = ?
	`, qrToken).Scan(
		&guest.ID, &guest.Slug, &guest.Name, &guest.PhoneNumber, &guest.QRToken,
		&guest.RSVPStatus, &guest.IsAttended, &attendedAt, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get guest by qr token: %w", err)
	}
	if attendedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", attendedAt.String)
		guest.AttendedAt = sql.NullTime{Time: t, Valid: true}
	}
	if createdAt.Valid {
		guest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	return guest, nil
}

func (r *Repository) GetAllGuests() ([]models.Guest, error) {
	rows, err := r.db.Query(`
		SELECT id, slug, name, phone_number, qr_token, rsvp_status, is_attended, attended_at, created_at
		FROM guests ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all guests: %w", err)
	}
	defer rows.Close()

	var guests []models.Guest
	for rows.Next() {
		guest := models.Guest{}
		var attendedAt, createdAt sql.NullString
		err := rows.Scan(
			&guest.ID, &guest.Slug, &guest.Name, &guest.PhoneNumber, &guest.QRToken,
			&guest.RSVPStatus, &guest.IsAttended, &attendedAt, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan guest: %w", err)
		}
		if attendedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", attendedAt.String)
			guest.AttendedAt = sql.NullTime{Time: t, Valid: true}
		}
		if createdAt.Valid {
			guest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		guests = append(guests, guest)
	}
	return guests, nil
}

func (r *Repository) CreateGuest(guest *models.Guest) error {
	if guest.QRToken == "" {
		guest.QRToken = uuid.New().String()
	}
	if guest.Slug == "" {
		guest.Slug = sanitizeSlug(guest.Name)
	}

	result, err := r.db.Exec(`
		INSERT INTO guests (slug, name, phone_number, qr_token, rsvp_status, is_attended, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, guest.Slug, guest.Name, guest.PhoneNumber, guest.QRToken, guest.RSVPStatus, guest.IsAttended, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create guest: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	guest.ID = id
	return nil
}

func (r *Repository) UpdateGuest(guest *models.Guest) error {
	_, err := r.db.Exec(`
		UPDATE guests
		SET slug = ?, name = ?, phone_number = ?, rsvp_status = ?, is_attended = ?, attended_at = ?
		WHERE id = ?
	`, guest.Slug, guest.Name, guest.PhoneNumber, guest.RSVPStatus, guest.IsAttended, guest.AttendedAt, guest.ID)
	if err != nil {
		return fmt.Errorf("failed to update guest: %w", err)
	}
	return nil
}

func (r *Repository) DeleteGuest(id int64) error {
	_, err := r.db.Exec("DELETE FROM guests WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete guest: %w", err)
	}
	return nil
}

func (r *Repository) UpdateRSVP(guestID int64, status string) error {
	_, err := r.db.Exec("UPDATE guests SET rsvp_status = ? WHERE id = ?", status, guestID)
	if err != nil {
		return fmt.Errorf("failed to update rsvp: %w", err)
	}
	return nil
}

func (r *Repository) MarkAttended(qrToken string) error {
	_, err := r.db.Exec(`
		UPDATE guests SET is_attended = 1, attended_at = ? WHERE qr_token = ?
	`, time.Now(), qrToken)
	if err != nil {
		return fmt.Errorf("failed to mark attended: %w", err)
	}
	return nil
}

func (r *Repository) InsertGuestbook(guestID int64, message string) error {
	var existingID int64
	err := r.db.QueryRow(`SELECT id FROM guestbooks WHERE guest_id = ? ORDER BY id DESC LIMIT 1`, guestID).Scan(&existingID)
	if err == nil {
		_, err = r.db.Exec(`UPDATE guestbooks SET message = ?, created_at = ? WHERE id = ?`, message, time.Now(), existingID)
		if err != nil {
			return fmt.Errorf("failed to update guestbook: %w", err)
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing guestbook: %w", err)
	}
	_, err = r.db.Exec(`INSERT INTO guestbooks (guest_id, message, created_at) VALUES (?, ?, ?)`, guestID, message, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert guestbook: %w", err)
	}
	return nil
}

func (r *Repository) GetGuestbookByGuestIDCount(guestID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM guestbooks WHERE guest_id = ?`, guestID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count guestbook: %w", err)
	}
	return count, nil
}

func (r *Repository) DeleteGuestbook(id int64) error {
	_, err := r.db.Exec(`DELETE FROM guestbooks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete guestbook: %w", err)
	}
	return nil
}

func (r *Repository) GetGuestbooks() ([]models.Guestbook, error) {
	rows, err := r.db.Query(`
		SELECT id, guest_id, message, created_at FROM guestbooks ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get guestbooks: %w", err)
	}
	defer rows.Close()

	var guestbooks []models.Guestbook
	for rows.Next() {
		gb := models.Guestbook{}
		err := rows.Scan(&gb.ID, &gb.GuestID, &gb.Message, &gb.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan guestbook: %w", err)
		}
		guestbooks = append(guestbooks, gb)
	}
	return guestbooks, nil
}

func (r *Repository) GetGuestbookByGuestID(guestID int64) ([]models.Guestbook, error) {
	rows, err := r.db.Query(`
		SELECT id, guest_id, message, created_at FROM guestbooks WHERE guest_id = ? ORDER BY created_at DESC
	`, guestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guestbooks by guest id: %w", err)
	}
	defer rows.Close()

	var guestbooks []models.Guestbook
	for rows.Next() {
		gb := models.Guestbook{}
		err := rows.Scan(&gb.ID, &gb.GuestID, &gb.Message, &gb.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan guestbook: %w", err)
		}
		guestbooks = append(guestbooks, gb)
	}
	return guestbooks, nil
}

func (r *Repository) GetGuestbooksPaginated(page, limit int) ([]models.GuestbookWithGuest, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM guestbooks`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count guestbooks: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT g.id, g.guest_id, gs.name, g.message, g.created_at 
		FROM guestbooks g 
		JOIN guests gs ON g.guest_id = gs.id 
		ORDER BY g.created_at DESC 
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get guestbooks: %w", err)
	}
	defer rows.Close()

	var guestbooks []models.GuestbookWithGuest
	for rows.Next() {
		gb := models.GuestbookWithGuest{}
		var createdAt sql.NullString
		err := rows.Scan(&gb.ID, &gb.GuestID, &gb.GuestName, &gb.Message, &createdAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan guestbook: %w", err)
		}
		if createdAt.Valid {
			gb.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		guestbooks = append(guestbooks, gb)
	}
	return guestbooks, total, nil
}

func (r *Repository) GetRundowns() ([]models.Rundown, error) {
	rows, err := r.db.Query(`
		SELECT id, start_time, end_time, activity_name, description FROM rundowns ORDER BY start_time ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get rundowns: %w", err)
	}
	defer rows.Close()

	var rundowns []models.Rundown
	for rows.Next() {
		rd := models.Rundown{}
		err := rows.Scan(&rd.ID, &rd.StartTime, &rd.EndTime, &rd.ActivityName, &rd.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rundown: %w", err)
		}
		rundowns = append(rundowns, rd)
	}
	return rundowns, nil
}

func (r *Repository) CreateRundown(rd *models.Rundown) error {
	result, err := r.db.Exec(`
		INSERT INTO rundowns (start_time, end_time, activity_name, description) VALUES (?, ?, ?, ?)
	`, rd.StartTime, rd.EndTime, rd.ActivityName, rd.Description)
	if err != nil {
		return fmt.Errorf("failed to create rundown: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	rd.ID = id
	return nil
}

func (r *Repository) DeleteRundown(id int64) error {
	_, err := r.db.Exec("DELETE FROM rundowns WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete rundown: %w", err)
	}
	return nil
}

func (r *Repository) GetGalleries() ([]models.Gallery, error) {
	rows, err := r.db.Query(`
		SELECT id, image_url, caption, sort_order FROM galleries ORDER BY sort_order ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get galleries: %w", err)
	}
	defer rows.Close()

	var galleries []models.Gallery
	for rows.Next() {
		g := models.Gallery{}
		err := rows.Scan(&g.ID, &g.ImageURL, &g.Caption, &g.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gallery: %w", err)
		}
		galleries = append(galleries, g)
	}
	return galleries, nil
}

func (r *Repository) CreateGallery(g *models.Gallery) error {
	result, err := r.db.Exec(`
		INSERT INTO galleries (image_url, caption, sort_order) VALUES (?, ?, ?)
	`, g.ImageURL, g.Caption, g.SortOrder)
	if err != nil {
		return fmt.Errorf("failed to create gallery: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	g.ID = id
	return nil
}

func (r *Repository) DeleteGallery(id int64) error {
	_, err := r.db.Exec("DELETE FROM galleries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete gallery: %w", err)
	}
	return nil
}

func (r *Repository) GetStats() (*models.Stats, error) {
	stats := &models.Stats{}

	err := r.db.QueryRow("SELECT COUNT(*) FROM guests").Scan(&stats.TotalGuests)
	if err != nil {
		return nil, fmt.Errorf("failed to count guests: %w", err)
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM guests WHERE rsvp_status != 'pending'").Scan(&stats.TotalRSVP)
	if err != nil {
		return nil, fmt.Errorf("failed to count rsvp: %w", err)
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM guests WHERE is_attended = 1").Scan(&stats.TotalAttended)
	if err != nil {
		return nil, fmt.Errorf("failed to count attended: %w", err)
	}

	return stats, nil
}

func (r *Repository) ImportGuests(guests []models.Guest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO guests (slug, name, phone_number, qr_token, rsvp_status, is_attended, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, guest := range guests {
		if guest.QRToken == "" {
			guest.QRToken = uuid.New().String()
		}
		if guest.Slug == "" {
			guest.Slug = sanitizeSlug(guest.Name)
		}

		_, err := stmt.Exec(guest.Slug, guest.Name, guest.PhoneNumber, guest.QRToken, guest.RSVPStatus, guest.IsAttended, time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert guest during import: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func sanitizeSlug(name string) string {
	name = strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	name = reg.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}
