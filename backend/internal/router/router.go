package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/handler"
	"unistep-platform/backend/internal/middleware"
	"unistep-platform/backend/internal/response"
)

// New 构建路由；uploader 可为 nil（此时附件上传接口将返回 STORAGE_DISABLED）
func New(cfg config.Config, db *gorm.DB, uploader handler.Uploader) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger 文档与 UI
	r.GET("/swagger.json", handler.SwaggerJSON)
	r.GET("/swagger", handler.SwaggerUI)

	api := r.Group("/api/v1")
	api.GET("/ping", func(c *gin.Context) {
		response.OK(c, gin.H{"message": "pong"})
	})

	// 用户认证模块：注册、登录
	authHandler := handler.NewAuthHandler(db, cfg.JWTSecret)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// 需要登录的接口统一挂载 JWT 中间件
	userHandler := handler.NewUserHandler(db)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.GET("/me", func(c *gin.Context) {
		response.OK(c, gin.H{"message": "authenticated"})
	})
	protected.GET("/users/me", userHandler.GetProfile)

	// 团员发展模块：登录后可访问；后续可叠加更细的角色控制
	memberHandler := handler.NewMemberHandler(db, uploader)
	members := protected.Group("/members")
	members.GET("", memberHandler.ListProfiles)
	members.POST("", memberHandler.CreateProfile)
	members.GET("/:id", memberHandler.GetProfile)
	members.PUT("/:id", memberHandler.UpdateProfile)
	members.DELETE("/:id", memberHandler.DeleteProfile)
	members.GET("/:id/archive", memberHandler.GenerateArchive)
	members.POST("/:id/applications", memberHandler.CreateApplication)
	members.POST("/:id/activists", memberHandler.CreateActivistRecord)
	members.POST("/:id/develop-targets", memberHandler.CreateDevelopRecord)
	members.POST("/:id/political-reviews", memberHandler.CreatePoliticalReview)
	members.POST("/:id/attachments", memberHandler.UploadAttachment)

	// 社团活动模块
	activityHandler := handler.NewActivityHandler(db, uploader)
	activities := protected.Group("/activities")
	activities.GET("", activityHandler.ListActivities)
	activities.POST("", activityHandler.CreateActivity)
	activities.GET("/statistics", activityHandler.ActivityStatistics)
	activities.GET("/:id", activityHandler.GetActivity)
	activities.PUT("/:id", activityHandler.UpdateActivity)
	activities.DELETE("/:id", activityHandler.DeleteActivity)
	activities.POST("/:id/submit", activityHandler.SubmitForApproval)
	activities.POST("/:id/approve", activityHandler.ApproveActivity)
	activities.POST("/:id/register", activityHandler.RegisterActivity)
	activities.POST("/:id/cancel-registration", activityHandler.CancelRegistration)
	activities.POST("/:id/checkin", activityHandler.CheckinActivity)
	activities.POST("/:id/files", activityHandler.UploadActivityImage)
	activities.POST("/:id/summary", activityHandler.SubmitSummary)
	activities.PUT("/:id/status", activityHandler.UpdateStatus)

	// 学生社区与自治队伍模块
	communityHandler := handler.NewCommunityHandler(db)
	community := protected.Group("/community")
	community.GET("/teams/statistics", communityHandler.TeamStatistics)
	community.GET("/service-profile", communityHandler.ServiceProfile)
	community.GET("/teams", communityHandler.ListTeams)
	community.POST("/teams", communityHandler.CreateTeam)
	community.GET("/teams/:id", communityHandler.GetTeam)
	community.PUT("/teams/:id", communityHandler.UpdateTeam)
	community.DELETE("/teams/:id", communityHandler.DeleteTeam)
	community.GET("/teams/:id/members", communityHandler.ListTeamMembers)
	community.POST("/teams/:id/members", communityHandler.AddTeamMember)
	community.PUT("/teams/:id/members/:memberId", communityHandler.UpdateTeamMember)
	community.DELETE("/teams/:id/members/:memberId", communityHandler.RemoveTeamMember)
	community.GET("/teams/:id/duties", communityHandler.ListDutySchedules)
	community.POST("/teams/:id/duties", communityHandler.CreateDutySchedule)
	community.POST("/teams/:id/duties/:scheduleId/checkin", communityHandler.DutyCheckin)
	community.POST("/teams/:id/duties/:scheduleId/checkout", communityHandler.DutyCheckout)
	community.GET("/teams/:id/services", communityHandler.ListVolunteerServices)
	community.POST("/teams/:id/services", communityHandler.CreateVolunteerService)
	community.PUT("/teams/:id/services/:serviceId/verify", communityHandler.VerifyVolunteerService)

	// 勤工助学模块
	workstudyHandler := handler.NewWorkStudyHandler(db, uploader)
	workstudy := protected.Group("/workstudy")
	workstudy.GET("/statistics", workstudyHandler.WorkStudyStatistics)
	workstudy.GET("/jobs", workstudyHandler.ListJobs)
	workstudy.POST("/jobs", workstudyHandler.CreateJob)
	workstudy.GET("/jobs/:id", workstudyHandler.GetJob)
	workstudy.PUT("/jobs/:id", workstudyHandler.UpdateJob)
	workstudy.DELETE("/jobs/:id", workstudyHandler.DeleteJob)
	workstudy.POST("/jobs/:id/publish", workstudyHandler.PublishJob)
	workstudy.POST("/jobs/:id/close", workstudyHandler.CloseJob)
	workstudy.POST("/jobs/:id/apply", workstudyHandler.ApplyJob)
	workstudy.POST("/jobs/:id/cancel-application", workstudyHandler.CancelApplication)
	workstudy.GET("/jobs/:id/applications", workstudyHandler.ListApplications)
	workstudy.POST("/jobs/:id/applications/:appId/accept", workstudyHandler.AcceptApplication)
	workstudy.POST("/jobs/:id/applications/:appId/reject", workstudyHandler.RejectApplication)
	workstudy.POST("/jobs/:id/attendances", workstudyHandler.CreateAttendance)
	workstudy.GET("/jobs/:id/attendances", workstudyHandler.ListAttendances)
	workstudy.PUT("/attendances/:attId/checkout", workstudyHandler.CheckoutAttendance)
	workstudy.POST("/jobs/:id/salary/calculate", workstudyHandler.CalculateSalary)
	workstudy.GET("/jobs/:id/salary", workstudyHandler.ListSalaries)
	workstudy.PUT("/salary/:salaryId/pay", workstudyHandler.PaySalary)
	workstudy.POST("/jobs/:id/files", workstudyHandler.UploadJobFile)

	// 统计分析仪表盘（登录即可访问）
	dashboardHandler := handler.NewDashboardHandler(db)
	dashboard := protected.Group("/dashboard")
	dashboard.GET("/overview", dashboardHandler.Overview)
	dashboard.GET("/member-trend", dashboardHandler.MemberTrend)
	dashboard.GET("/activity-trend", dashboardHandler.ActivityTrend)
	dashboard.GET("/service-trend", dashboardHandler.ServiceTrend)

	// 管理员接口
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	admin.GET("/dashboard", dashboardHandler.Overview)
	// RBAC 管理：用户、角色、权限列表
	adminHandler := handler.NewAdminHandler(db)
	admin.GET("/users", adminHandler.ListUsers)
	admin.GET("/roles", adminHandler.ListRoles)
	admin.GET("/permissions", adminHandler.ListPermissions)

	return r
}
