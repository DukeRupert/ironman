package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dukerupert/ironman/dto"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func addRoutes(mux *http.ServeMux, t *Template) {
	// Create a FileServer handler for the "static" directory
	fs := http.FileServer(http.Dir("./public/static"))

	// Handle all requests at the root path using the FileServer
	http.Handle("/static", fs)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "landing", nil)
	})
	
	// e.GET("/", handleGetLandingPage)
	// e.GET("/login", handleGetLoginPage)
	// e.GET("/signup", handleGetSignupPage)
	// e.GET("/forgot-password", handleForgotPasswordPage)
	// e.POST("/forgot-password", handleForgotPassword)
	// e.GET("/reset-password", handleResetPasswordPage)
	// e.POST("/reset-password", handleResetPassword)
	// e.GET("/app/dashboard", handleDashboardPage)
	// e.GET("/app/projects", handleDashboardPage)
	// e.GET("/app/projects/:id", handleProjectDetails)
	// e.GET("/hello", Hello)
	// e.GET("/upload", upload)
	// e.POST("/upload", handleUpload)
}

func handleGetLandingPage(c echo.Context) error {
	return c.Render(http.StatusOK, "landing", nil)
}

func handleGetLoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func handleGetSignupPage(c echo.Context) error {
	return c.Render(http.StatusOK, "signup", nil)
}

func handleForgotPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "forgot-password", nil)
}

func handleForgotPassword(c echo.Context) error {
	return c.String(http.StatusOK, "FIXME: Send email with reset link")
}

func handleResetPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "reset-password", nil)
}

func handleResetPassword(c echo.Context) error {
	return c.String(http.StatusOK, "FIXME: Update password")
}

func handleDashboardPage(c echo.Context) error {
	user := getCurrentUser(c)

    data := dto.DashboardData{
        AppData: dto.AppData{
            PageTitle: "Dashboard",
            CurrentPage: "dashboard",
            User: user,
            RecentProjects: getRecentProjects(user.ID),
        },
        Stats: getDashboardStats(user.ID),
        RecentProjects: getDetailedProjects(user.ID, 5),
        CriticalViolations: getCriticalViolations(user.ID),
    }
    return c.Render(http.StatusOK, "dashboard", data)
}

// getCurrentUser returns the current authenticated user
func getCurrentUser(c echo.Context) dto.User {
	return dto.User{
		ID:       "user_123",
		Name:     "John Doe",
		Email:    "john.doe@abcconstruction.com",
		Initials: "JD",
		Role:     "inspector",
		Company:  "ABC Construction",
		Avatar:   "", // No avatar for now
	}
}

// getRecentProjects returns recent projects for sidebar navigation
func getRecentProjects(userID string) []dto.RecentProject {
	return []dto.RecentProject{
		{
			ID:            "proj_001",
			Name:          "Downtown Office Complex",
			InitialLetter: "D",
			Status:        "active",
		},
		{
			ID:            "proj_002",
			Name:          "Riverside Apartments",
			InitialLetter: "R",
			Status:        "active",
		},
		{
			ID:            "proj_003",
			Name:          "Metro Shopping Center",
			InitialLetter: "M",
			Status:        "completed",
		},
		{
			ID:            "proj_004",
			Name:          "Industrial Warehouse",
			InitialLetter: "I",
			Status:        "active",
		},
	}
}

// getDashboardStats returns dashboard statistics
func getDashboardStats(userID string) dto.DashboardStats {
	return dto.DashboardStats{
		TotalInspections: 147,
		ViolationsFound:  23,
		ComplianceRate:   94.3,
		ActiveProjects:   8,
	}
}

// getDetailedProjects returns full project details for dashboard
func getDetailedProjects(userID string, limit int) []dto.Project {
	projects := []dto.Project{
		{
			ID:                   "proj_001",
			Name:                 "Downtown Office Complex",
			Description:          "15-story office building construction",
			Status:               "in-progress",
			Location:             "425 Market St, San Francisco, CA",
			CreatedAt:            time.Now().AddDate(0, -2, -15),
			LastUpdated:          time.Now().AddDate(0, 0, -2),
			LastUpdatedFormatted: "2 days ago",
			ViolationCount:       3,
			ComplianceScore:      91.2,
			Inspector:            "John Doe",
			InspectorID:          "user_123",
			PhotoCount:           24,
			ReportGenerated:      false,
		},
		{
			ID:                   "proj_002",
			Name:                 "Riverside Apartments",
			Description:          "120-unit residential complex",
			Status:               "needs-review",
			Location:             "1200 River Rd, Portland, OR",
			CreatedAt:            time.Now().AddDate(0, -1, -20),
			LastUpdated:          time.Now().AddDate(0, 0, -5),
			LastUpdatedFormatted: "5 days ago",
			ViolationCount:       7,
			ComplianceScore:      85.6,
			Inspector:            "Sarah Wilson",
			InspectorID:          "user_456",
			PhotoCount:           18,
			ReportGenerated:      true,
		},
		{
			ID:                   "proj_003",
			Name:                 "Metro Shopping Center",
			Description:          "250,000 sq ft retail complex",
			Status:               "completed",
			Location:             "3400 Metro Blvd, Seattle, WA",
			CreatedAt:            time.Now().AddDate(0, -3, -10),
			LastUpdated:          time.Now().AddDate(0, 0, -1),
			LastUpdatedFormatted: "1 day ago",
			ViolationCount:       2,
			ComplianceScore:      96.8,
			Inspector:            "Mike Johnson",
			InspectorID:          "user_789",
			PhotoCount:           32,
			ReportGenerated:      true,
		},
		{
			ID:                   "proj_004",
			Name:                 "Industrial Warehouse",
			Description:          "500,000 sq ft distribution center",
			Status:               "in-progress",
			Location:             "5500 Industrial Way, Phoenix, AZ",
			CreatedAt:            time.Now().AddDate(0, -1, -5),
			LastUpdated:          time.Now().AddDate(0, 0, -3),
			LastUpdatedFormatted: "3 days ago",
			ViolationCount:       5,
			ComplianceScore:      88.4,
			Inspector:            "Lisa Chen",
			InspectorID:          "user_101",
			PhotoCount:           15,
			ReportGenerated:      false,
		},
		{
			ID:                   "proj_005",
			Name:                 "Hospital Expansion",
			Description:          "New emergency wing construction",
			Status:               "in-progress",
			Location:             "1000 Medical Center Dr, Denver, CO",
			CreatedAt:            time.Now().AddDate(0, 0, -30),
			LastUpdated:          time.Now().AddDate(0, 0, -7),
			LastUpdatedFormatted: "1 week ago",
			ViolationCount:       1,
			ComplianceScore:      98.2,
			Inspector:            "John Doe",
			InspectorID:          "user_123",
			PhotoCount:           8,
			ReportGenerated:      false,
		},
	}

	// Return only the requested number of projects
	if limit > 0 && limit < len(projects) {
		return projects[:limit]
	}
	return projects
}

func handleProjectDetails(c echo.Context) error {
	fmt.Println("handleProjectDetails()")
    projectID := c.Param("id")
	fmt.Println(projectID)
    project, _ := getProjectById(projectID)
    
    data := dto.ProjectDetailData{
        AppData: dto.AppData{
            PageTitle: project.Name,
            CurrentPage: "projects",
            User: getCurrentUser(c),
            RecentProjects: getRecentProjects(getCurrentUser(c).ID),
        },
        Project: *project,
        Violations: getViolationsByProject(projectID),
        Timeline: getProjectTimeline(projectID),
        CanEdit: canUserEditProject(getCurrentUser(c).ID, projectID),
        CanDelete: canUserDeleteProject(getCurrentUser(c).ID, projectID),
    }
    
    return c.Render(http.StatusOK, "project-detail", data)
}

// getProjectTimeline returns activity timeline for a project
func getProjectTimeline(projectID string) []dto.TimelineEvent {
	return []dto.TimelineEvent{
		{
			ID:          "timeline_001",
			ProjectID:   projectID,
			Type:        "created",
			Description: "Project created by",
			UserID:      "user_123",
			UserName:    "John Doe",
			Timestamp:   time.Now().AddDate(0, -2, -15),
			Metadata:    map[string]interface{}{},
		},
		{
			ID:          "timeline_002",
			ProjectID:   projectID,
			Type:        "photo_uploaded",
			Description: "Uploaded 8 photos by",
			UserID:      "user_123",
			UserName:    "John Doe",
			Timestamp:   time.Now().AddDate(0, -2, -14),
			Metadata:    map[string]interface{}{"photo_count": 8},
		},
		{
			ID:          "timeline_003",
			ProjectID:   projectID,
			Type:        "violation_found",
			Description: "Safety violation detected by AI -",
			UserID:      "system",
			UserName:    "SafeSite AI",
			Timestamp:   time.Now().AddDate(0, -2, -13),
			Metadata:    map[string]interface{}{"violation_id": "viol_001"},
		},
		{
			ID:          "timeline_004",
			ProjectID:   projectID,
			Type:        "photo_uploaded",
			Description: "Uploaded 12 additional photos by",
			UserID:      "user_456",
			UserName:    "Sarah Wilson",
			Timestamp:   time.Now().AddDate(0, -1, -20),
			Metadata:    map[string]interface{}{"photo_count": 12},
		},
		{
			ID:          "timeline_005",
			ProjectID:   projectID,
			Type:        "violation_found",
			Description: "Critical safety violation identified by",
			UserID:      "user_456",
			UserName:    "Sarah Wilson",
			Timestamp:   time.Now().AddDate(0, 0, -5),
			Metadata:    map[string]interface{}{"violation_id": "viol_002"},
		},
		{
			ID:          "timeline_006",
			ProjectID:   projectID,
			Type:        "violation_resolved",
			Description: "Safety violation marked as resolved by",
			UserID:      "user_789",
			UserName:    "Mike Johnson",
			Timestamp:   time.Now().AddDate(0, 0, -2),
			Metadata:    map[string]interface{}{"violation_id": "viol_003"},
		},
		{
			ID:          "timeline_007",
			ProjectID:   projectID,
			Type:        "report_generated",
			Description: "Safety inspection report generated by",
			UserID:      "user_123",
			UserName:    "John Doe",
			Timestamp:   time.Now().AddDate(0, 0, -1),
			Metadata:    map[string]interface{}{"report_id": "report_001"},
		},
	}
}

// getCriticalViolations returns high-priority violations needing attention
func getCriticalViolations(userID string) []dto.Violation {
	return []dto.Violation{
		{
			ID:           "viol_001",
			ProjectID:    "proj_002",
			ProjectName:  "Riverside Apartments",
			Description:  "Workers not wearing hard hats in active construction zone",
			Regulation:   "1926.95",
			RiskLevel:    "critical",
			Category:     "PPE",
			Location:     "Building A, 3rd Floor",
			PhotoURL:     "/photos/viol_001.jpg",
			Status:       "open",
			FoundAt:      time.Now().AddDate(0, 0, -5),
			ResolvedAt:   nil,
			Notes:        "Multiple workers observed without proper head protection during concrete pour",
			AIConfidence: 0.94,
		},
		{
			ID:           "viol_002",
			ProjectID:    "proj_001",
			ProjectName:  "Downtown Office Complex",
			Description:  "Unsecured scaffolding exceeding height limits",
			Regulation:   "1926.451",
			RiskLevel:    "high",
			Category:     "Fall Protection",
			Location:     "East Side Exterior",
			PhotoURL:     "/photos/viol_002.jpg",
			Status:       "open",
			FoundAt:      time.Now().AddDate(0, 0, -2),
			ResolvedAt:   nil,
			Notes:        "Scaffolding platform at 15ft height without proper guardrails",
			AIConfidence: 0.87,
		},
		{
			ID:           "viol_003",
			ProjectID:    "proj_004",
			ProjectName:  "Industrial Warehouse",
			Description:  "Electrical panel left open and unlocked",
			Regulation:   "1926.416",
			RiskLevel:    "high",
			Category:     "Electrical",
			Location:     "Main Electrical Room",
			PhotoURL:     "/photos/viol_003.jpg",
			Status:       "open",
			FoundAt:      time.Now().AddDate(0, 0, -3),
			ResolvedAt:   nil,
			Notes:        "480V panel accessible to unauthorized personnel",
			AIConfidence: 0.91,
		},
		{
			ID:           "viol_004",
			ProjectID:    "proj_002",
			ProjectName:  "Riverside Apartments",
			Description:  "Improper ladder placement and angle",
			Regulation:   "1926.1053",
			RiskLevel:    "critical",
			Category:     "Fall Protection",
			Location:     "Building B, Stairwell",
			PhotoURL:     "/photos/viol_004.jpg",
			Status:       "open",
			FoundAt:      time.Now().AddDate(0, 0, -1),
			ResolvedAt:   nil,
			Notes:        "Extension ladder at unsafe angle (>75 degrees) with no spotter",
			AIConfidence: 0.89,
		},
		{
			ID:           "viol_005",
			ProjectID:    "proj_001",
			ProjectName:  "Downtown Office Complex",
			Description:  "Missing safety signage in excavation area",
			Regulation:   "1926.651",
			RiskLevel:    "high",
			Category:     "Excavation",
			Location:     "North Parking Area",
			PhotoURL:     "/photos/viol_005.jpg",
			Status:       "open",
			FoundAt:      time.Now().AddDate(0, 0, -4),
			ResolvedAt:   nil,
			Notes:        "8ft deep excavation without proper warning signs or barriers",
			AIConfidence: 0.93,
		},
	}
}

// getProjectById returns a single project by ID
func getProjectById(projectID string) (*dto.Project, error) {
	projects := getDetailedProjects("", 0) // Get all projects
	for _, project := range projects {
		if project.ID == projectID {
			return &project, nil
		}
	}
	return nil, fmt.Errorf("project not found")
}

// getViolationsByProject returns all violations for a specific project
func getViolationsByProject(projectID string) []dto.Violation {
	allViolations := getCriticalViolations("") // Get all violations
	var projectViolations []dto.Violation
	
	for _, violation := range allViolations {
		if violation.ProjectID == projectID {
			projectViolations = append(projectViolations, violation)
		}
	}
	return projectViolations
}

// getUserRole returns the role for permission checks
func getUserRole(userID string) string {
	user := getCurrentUser(nil) // Simplified for mock
	return user.Role
}

// canUserEditProject checks if user can edit a project
func canUserEditProject(userID, projectID string) bool {
	role := getUserRole(userID)
	return role == "admin" || role == "inspector"
}

// canUserDeleteProject checks if user can delete a project
func canUserDeleteProject(userID, projectID string) bool {
	role := getUserRole(userID)
	return role == "admin"
}

func GlobalMiddleware(e *echo.Echo, logger *slog.Logger) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
}
