package routes

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"undangan-digital/internal/handlers"
	"undangan-digital/internal/middleware"
)

func loadTemplates(engine *gin.Engine) error {
	templ := template.New("")

	rootFiles, err := filepath.Glob("templates/*.html")
	if err != nil {
		return err
	}
	for _, f := range rootFiles {
		name := filepath.Base(f)
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		templ, err = templ.New(name).Parse(string(data))
		if err != nil {
			return err
		}
	}

	adminFiles, err := filepath.Glob("templates/admin/*.html")
	if err != nil {
		return err
	}
	for _, f := range adminFiles {
		name := "admin/" + filepath.Base(f)
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		templ, err = templ.New(name).Parse(string(data))
		if err != nil {
			return err
		}
	}

	engine.SetHTMLTemplate(templ)
	return nil
}

func SetupRouter(h *handlers.Handler, session *middleware.SessionManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	router.Static("/assets", "./assets")

	if err := loadTemplates(router); err != nil {
		panic("Failed to load templates: " + err.Error())
	}

	router.GET("/", h.HomePage)
	router.GET("/undangan/:slug", h.GetInvitation)
	router.POST("/api/rsvp", h.PostRSVP)
	router.GET("/api/guestbooks", h.GetPublicGuestbooks)

	router.GET("/admin/login", h.GetLogin)
	router.POST("/admin/login", h.PostLogin)
	router.GET("/admin/logout", h.PostLogout)
	router.POST("/admin/logout", h.PostLogout)

	protected := router.Group("/admin")
	protected.Use(session.SessionAuthMiddleware())
	{
		protected.GET("/dashboard", h.GetAdminDashboard)
		protected.GET("/guests", h.GetAdminGuests)
		protected.GET("/settings", h.GetAdminSettings)
		protected.GET("/scanner", h.GetAdminScanner)
		protected.GET("/rundowns", h.GetAdminRundownsPage)
		protected.GET("/galleries", h.GetAdminGalleriesPage)
		protected.GET("/guestbooks", h.GetAdminGuestbooksPage)

		protected.POST("/api/settings", h.PostAdminSettings)

		protected.POST("/api/guests", h.PostAdminGuests)
		protected.PUT("/api/guests", h.PutAdminGuest)
		protected.DELETE("/api/guests/:id", h.DeleteAdminGuest)
		protected.POST("/api/guests/import", h.PostAdminImportGuests)

		protected.POST("/api/scan", h.PostScan)
		protected.GET("/api/stats", h.GetStatsAPI)

		protected.GET("/api/rundowns", h.GetAdminRundowns)
		protected.POST("/api/rundowns", h.PostAdminRundown)
		protected.DELETE("/api/rundowns/:id", h.DeleteAdminRundown)

		protected.GET("/api/galleries", h.GetAdminGalleries)
		protected.POST("/api/galleries", h.PostAdminGallery)
		protected.DELETE("/api/galleries/:id", h.DeleteAdminGallery)

		protected.DELETE("/api/guestbooks/:id", h.DeleteAdminGuestbook)

		protected.POST("/api/broadcast", h.PostAdminBroadcast)
		protected.POST("/api/broadcast/test", h.PostAdminBroadcastTest)
	}

	api := router.Group("/api/admin")
	api.Use(session.APIAuthMiddleware())
	{
		api.POST("/login", h.PostLogin)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.NoRoute(h.NotFound)

	return router
}
