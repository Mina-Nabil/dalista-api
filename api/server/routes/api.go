package routes

import (
	"dalista/dalista-be/api/server/handlers/activityHandlers"
	"dalista/dalista-be/api/server/handlers/appleIAPHandlers"
	"dalista/dalista-be/api/server/handlers/authHandlers"
	"dalista/dalista-be/api/server/handlers/billingHandlers"
	"dalista/dalista-be/api/server/handlers/healthHandlers"
	"dalista/dalista-be/api/server/handlers/listHandlers"
	"dalista/dalista-be/api/server/handlers/periodicHandlers"
	"dalista/dalista-be/api/server/handlers/playBillingHandlers"
	"dalista/dalista-be/api/server/handlers/profileHandlers"
	"dalista/dalista-be/api/server/handlers/sharingHandlers"
	"dalista/dalista-be/api/server/middlewares"

	"github.com/gin-gonic/gin"
)

// LoadRoutes registers every feature's route group under /api/v1. Feature
// phases add their initializers here, each grouping endpoints and attaching the
// middleware they need.
func LoadRoutes(app *gin.Engine) {
	v1 := app.Group("/api/v1")

	initializeHealthRoutes(v1)
	initializeAuthRoutes(v1)
	initializeProfileRoutes(v1)
	initializeListDomainRoutes(v1)
	initializeBillingRoutes(v1)
	initializeSharingRoutes(v1)
	initializeActivityRoutes(v1)
	initializePeriodicRoutes(v1)
}

func initializeHealthRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", healthHandlers.Health)
}

func initializeAuthRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")

	// Public auth endpoints.
	g.POST("/register", authHandlers.RegisterEmail)
	g.POST("/login", authHandlers.LoginEmail)
	g.POST("/refresh", authHandlers.Refresh)
	g.POST("/otp/request", authHandlers.RequestSigninOTP)
	g.POST("/otp/verify", authHandlers.VerifySigninOTP)
	g.POST("/google", authHandlers.GoogleAuth)
	g.POST("/facebook", authHandlers.FacebookAuth)

	// Authenticated auth endpoints (post-signup phone verification, logout).
	authed := g.Group("")
	authed.Use(middlewares.AuthMiddleware())
	authed.POST("/logout", authHandlers.Logout)
	authed.POST("/phone/request", authHandlers.RequestPhoneVerification)
	authed.POST("/phone/verify", authHandlers.VerifyPhone)
}

func initializeProfileRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/profile")
	g.Use(middlewares.AuthMiddleware())
	g.GET("", profileHandlers.GetProfile)
	g.PUT("", profileHandlers.UpdateProfile)
}

// initializeListDomainRoutes wires categories, grouped lists, lists, items, and
// the category-interest export. All require authentication; per-record
// ownership is enforced in the service layer.
func initializeListDomainRoutes(rg *gin.RouterGroup) {
	auth := middlewares.AuthMiddleware()

	cat := rg.Group("/categories")
	cat.Use(auth)
	cat.GET("", listHandlers.ListCategories)
	cat.POST("", listHandlers.CreateCategory)
	cat.PUT("/:id", listHandlers.UpdateCategory)
	cat.DELETE("/:id", listHandlers.DeleteCategory)

	gl := rg.Group("/grouped-lists")
	gl.Use(auth)
	gl.GET("", listHandlers.ListGroupedLists)
	gl.POST("", listHandlers.CreateGroupedList)
	gl.PUT("/:id", listHandlers.UpdateGroupedList)
	gl.DELETE("/:id", listHandlers.DeleteGroupedList)

	lists := rg.Group("/lists")
	lists.Use(auth)
	lists.GET("", listHandlers.ListLists)
	lists.POST("", listHandlers.CreateList)
	lists.GET("/:id", listHandlers.GetList)
	lists.PUT("/:id", listHandlers.UpdateList)
	lists.DELETE("/:id", listHandlers.DeleteList)
	lists.GET("/:id/items", listHandlers.ListItems)
	lists.POST("/:id/items", listHandlers.CreateItem)

	items := rg.Group("/items")
	items.Use(auth)
	items.GET("/:id", listHandlers.GetItem)
	items.PUT("/:id", listHandlers.UpdateItem)
	items.DELETE("/:id", listHandlers.DeleteItem)

	exports := rg.Group("/exports")
	exports.Use(auth)
	exports.GET("/category-interest", listHandlers.CategoryInterestExport)
}

// initializeBillingRoutes wires subscription/storage (authed), purchase
// verification (authed), the admin oversight endpoints (authed + admin
// permission), and the provider webhooks (public — authenticity is the
// provider signature / purchase re-verification, so no auth middleware).
func initializeBillingRoutes(rg *gin.RouterGroup) {
	auth := middlewares.AuthMiddleware()

	billing := rg.Group("/billing")
	billing.Use(auth)
	billing.GET("/subscription", billingHandlers.GetSubscription)
	billing.GET("/storage", billingHandlers.GetStorage)
	billing.POST("/apple/verify", billingHandlers.VerifyApple)
	billing.POST("/google/verify", billingHandlers.VerifyGoogle)

	// Provider webhooks — public, verified by signature / purchase re-check.
	webhooks := rg.Group("/billing")
	webhooks.POST("/apple/notifications", appleIAPHandlers.Notifications)
	webhooks.POST("/google/rtdn", playBillingHandlers.RTDN)

	admin := rg.Group("/admin/subscriptions")
	admin.Use(auth)
	admin.GET("", billingHandlers.ListSubscriptions)
	admin.GET("/summary", billingHandlers.SubscriptionSummary)
}

// initializeSharingRoutes wires list sharing: share/manage collaborators on
// lists you own, view collaborators as a member, and list lists shared with you.
func initializeSharingRoutes(rg *gin.RouterGroup) {
	auth := middlewares.AuthMiddleware()

	shares := rg.Group("/lists/:id")
	shares.Use(auth)
	shares.POST("/shares", sharingHandlers.ShareList)
	shares.PUT("/shares/:userId", sharingHandlers.ChangeRole)
	shares.DELETE("/shares/:userId", sharingHandlers.RevokeShare)
	shares.DELETE("/invites/:inviteId", sharingHandlers.CancelInvite)
	shares.GET("/collaborators", sharingHandlers.ListCollaborators)

	// Separate top-level path (not /lists/...) to avoid a static-vs-:id route
	// conflict with the list CRUD group.
	sharedLists := rg.Group("/shared-lists")
	sharedLists.Use(auth)
	sharedLists.GET("", sharingHandlers.SharedWithMe)
}

// initializeActivityRoutes wires per-item comments, comment management, and the
// access-scoped activity timeline.
func initializeActivityRoutes(rg *gin.RouterGroup) {
	auth := middlewares.AuthMiddleware()

	items := rg.Group("/items/:id")
	items.Use(auth)
	items.POST("/comments", activityHandlers.CreateComment)
	items.GET("/comments", activityHandlers.ListComments)

	comments := rg.Group("/comments")
	comments.Use(auth)
	comments.PUT("/:id", activityHandlers.UpdateComment)
	comments.DELETE("/:id", activityHandlers.DeleteComment)

	rg.GET("/timeline", auth, activityHandlers.Timeline)
}

// initializePeriodicRoutes wires the need-to-buy toggle, the "today's list"
// review + generation, and variant history.
func initializePeriodicRoutes(rg *gin.RouterGroup) {
	auth := middlewares.AuthMiddleware()

	rg.PUT("/items/:id/need-to-buy", auth, periodicHandlers.SetNeedToBuy)

	lists := rg.Group("/lists/:id")
	lists.Use(auth)
	lists.GET("/today", periodicHandlers.TodayReview)
	lists.POST("/today", periodicHandlers.GenerateToday)
	lists.GET("/variants", periodicHandlers.ListVariants)

	variants := rg.Group("/variants")
	variants.Use(auth)
	variants.GET("/:id", periodicHandlers.GetVariant)
}
