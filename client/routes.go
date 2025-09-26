
package main

import (
	"embed"
	"fmt"
	"io/fs"

	"net/http"
	"time"

	"github.com/dukerupert/ironman/dto"

)

//go:embed public/static/*
var staticFS embed.FS

func addRoutes(mux *http.ServeMux, t *Template) {
	// Create a FileServer handler for the embedded "static" directory
	staticSubFS, err := fs.Sub(staticFS, "public/static")
	if err != nil {
		panic(fmt.Sprintf("failed to create static sub-filesystem: %v", err))
	}
	
	// Handle all requests to /static/ using the embedded FileServer
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))
	
	// Landing page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "landing", nil)
	})
	
	// Auth routes
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "login", nil)
	})
	
	mux.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "signup", nil)
	})
	
	mux.HandleFunc("GET /forgot-password", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "forgot-password", nil)
	})
	
	mux.HandleFunc("POST /forgot-password", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FIXME: Send email with reset link"))
	})
	
	mux.HandleFunc("GET /reset-password", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Token string
		}{
			Token: r.URL.Query().Get("token"),
		}
		t.Render(w, "reset-password", data)
	})
	
	mux.HandleFunc("POST /reset-password", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FIXME: Update password"))
	})
	
	// App routes (require authentication in production)
	mux.HandleFunc("GET /app/dashboard", func(w http.ResponseWriter, r *http.Request) {
		handleDashboard(w, r, t)
	})
	
	mux.HandleFunc("GET /app/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleProjectDetail(w, r, t)
	})
	
	mux.HandleFunc("GET /app/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "upload", nil)
	})
	
	mux.HandleFunc("POST /app/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FIXME: Handle upload"))
	})
	
	// Hello world example
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "hello", "World")
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request, t *Template) {
	user := getCurrentUser()

	data := dto.DashboardData{
		AppData: dto.AppData{
			PageTitle:      "Dashboard",
			CurrentPage:    "dashboard",
			User:           user,
			RecentProjects: getRecentProjects(user.ID),
		},
		Stats:              getDashboardStats(user.ID),
		RecentProjects:     getDetailedProjects(user.ID, 5),
		CriticalViolations: getCriticalViolations(user.ID),
	}
	t.Render(w, "dashboard", data)
}

func handleProjectDetail(w http.ResponseWriter, r *http.Request, t *Template) {
	projectID := r.PathValue("id")
	project, err := getProjectById(projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	data := dto.ProjectDetailData{
		AppData: dto.AppData{
			PageTitle:      project.Name,
			CurrentPage:    "projects",
			User:           getCurrentUser(),
			RecentProjects: getRecentProjects(getCurrentUser().ID),
		},
		Project:     *project,
		Violations:  getViolationsByProject(projectID),
		Timeline:    getProjectTimeline(projectID),
		CanEdit:     canUserEditProject(getCurrentUser().ID, projectID),
		CanDelete:   canUserDeleteProject(getCurrentUser().ID, projectID),
	}

	t.Render(w, "project-detail", data)
}

// getCurrentUser returns the current authenticated user
func getCurrentUser() dto.User {
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
	user := getCurrentUser() // Simplified for mock
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