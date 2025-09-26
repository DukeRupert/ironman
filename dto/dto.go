package dto

import "time"

// Base struct for all app pages
type AppData struct {
    PageTitle      string          // "Dashboard", "Projects", etc.
    CurrentPage    string          // "dashboard", "projects", etc. for nav highlighting
    User           User            // Current authenticated user
    RecentProjects []RecentProject // Recent projects for sidebar
}

// User represents the authenticated user
type User struct {
    ID       string `json:"id"`
    Name     string `json:"name"`         // "John Doe"
    Email    string `json:"email"`        // "john@company.com"
    Initials string `json:"initials"`     // "JD" - for avatar display
    Role     string `json:"role"`         // "admin", "inspector", "viewer"
    Company  string `json:"company"`      // "ABC Construction"
    Avatar   string `json:"avatar"`       // URL to profile image (optional)
}

// RecentProject for sidebar navigation
type RecentProject struct {
    ID            string `json:"id"`             // "proj_123"
    Name          string `json:"name"`           // "Downtown Office Building"
    InitialLetter string `json:"initial_letter"` // "D" - first letter for avatar
    Status        string `json:"status"`         // "active", "completed", "archived"
}

// Dashboard-specific data
type DashboardData struct {
    AppData                         // Embedded base data
    Stats            DashboardStats // Dashboard metrics
    RecentProjects   []Project      // Full project details (different from AppData.RecentProjects)
    CriticalViolations []Violation  // High-priority violations needing attention
}

// Dashboard statistics
type DashboardStats struct {
    TotalInspections int     `json:"total_inspections"`  // Total number of inspections performed
    ViolationsFound  int     `json:"violations_found"`   // Total violations identified
    ComplianceRate   float64 `json:"compliance_rate"`    // Percentage (e.g., 94.5)
    ActiveProjects   int     `json:"active_projects"`    // Currently active project count
}

// Full project details
type Project struct {
    ID                   string    `json:"id"`                     // "proj_123"
    Name                 string    `json:"name"`                   // "Downtown Office Building"
    Description          string    `json:"description"`            // Project description
    Status               string    `json:"status"`                 // "completed", "in-progress", "needs-review", "archived"
    Location             string    `json:"location"`               // "123 Main St, City, State"
    CreatedAt            time.Time `json:"created_at"`             
    LastUpdated          time.Time `json:"last_updated"`
    LastUpdatedFormatted string    `json:"last_updated_formatted"` // "2 days ago"
    ViolationCount       int       `json:"violation_count"`        // Number of violations found
    ComplianceScore      float64   `json:"compliance_score"`       // Percentage
    Inspector            string    `json:"inspector"`              // Inspector name
    InspectorID          string    `json:"inspector_id"`           // Inspector user ID
    PhotoCount           int       `json:"photo_count"`            // Number of photos uploaded
    ReportGenerated      bool      `json:"report_generated"`       // Whether final report exists
}

// Safety violation details
type Violation struct {
    ID           string    `json:"id"`            // "viol_456"
    ProjectID    string    `json:"project_id"`    // "proj_123"
    ProjectName  string    `json:"project_name"`  // "Downtown Office Building"
    Description  string    `json:"description"`   // "Missing hard hat in work area"
    Regulation   string    `json:"regulation"`    // "1926.95" (OSHA regulation number)
    RiskLevel    string    `json:"risk_level"`    // "high", "medium", "low", "critical"
    Category     string    `json:"category"`      // "PPE", "Fall Protection", "Electrical", etc.
    Location     string    `json:"location"`      // Specific location within project
    PhotoURL     string    `json:"photo_url"`     // URL to violation photo
    Status       string    `json:"status"`        // "open", "resolved", "dismissed"
    FoundAt      time.Time `json:"found_at"`      
    ResolvedAt   *time.Time `json:"resolved_at"`   // Null if not resolved
    Notes        string    `json:"notes"`         // Additional inspector notes
    AIConfidence float64   `json:"ai_confidence"` // AI detection confidence (0-1)
}

// Projects page data
type ProjectsData struct {
    AppData
    Projects     []Project          // All projects
    Filter       ProjectFilter      // Current filter settings
    Pagination   ProjectPagination  // Pagination info
    StatusCounts map[string]int     // Count by status for filter tabs
}

// Project filtering options
type ProjectFilter struct {
    Status     string `json:"status"`      // Filter by status
    Inspector  string `json:"inspector"`   // Filter by inspector
    DateFrom   string `json:"date_from"`   // Date range start
    DateTo     string `json:"date_to"`     // Date range end
    Search     string `json:"search"`      // Text search
    SortBy     string `json:"sort_by"`     // "name", "date", "violations", "compliance"
    SortOrder  string `json:"sort_order"`  // "asc", "desc"
}

// Pagination for projects
type ProjectPagination struct {
    CurrentPage int `json:"current_page"`
    TotalPages  int `json:"total_pages"`
    TotalItems  int `json:"total_items"`
    ItemsPerPage int `json:"items_per_page"`
    HasPrev     bool `json:"has_prev"`
    HasNext     bool `json:"has_next"`
}

// Individual project detail page data
type ProjectDetailData struct {
    AppData
    Project     Project     // Full project details
    Violations  []Violation // All violations for this project
    Photos      []Photo     // All photos for this project
    Timeline    []TimelineEvent // Project activity timeline
    CanEdit     bool        // Whether current user can edit
    CanDelete   bool        // Whether current user can delete
}

// Project photo
type Photo struct {
    ID          string    `json:"id"`           // "photo_789"
    ProjectID   string    `json:"project_id"`   // "proj_123"
    URL         string    `json:"url"`          // Photo URL
    ThumbnailURL string   `json:"thumbnail_url"` // Thumbnail URL
    Filename    string    `json:"filename"`     // Original filename
    UploadedAt  time.Time `json:"uploaded_at"`
    UploadedBy  string    `json:"uploaded_by"`  // User ID
    Caption     string    `json:"caption"`      // Optional caption
    ViolationIDs []string `json:"violation_ids"` // Violations found in this photo
}

// Timeline event for project activity
type TimelineEvent struct {
    ID          string    `json:"id"`
    ProjectID   string    `json:"project_id"`
    Type        string    `json:"type"`        // "created", "photo_uploaded", "violation_found", "violation_resolved", "report_generated"
    Description string    `json:"description"` // Human-readable description
    UserID      string    `json:"user_id"`     // Who performed the action
    UserName    string    `json:"user_name"`   // User's display name
    Timestamp   time.Time `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata"` // Additional type-specific data
}

// New inspection form data
type NewInspectionData struct {
    AppData
    LocationSuggestions []string `json:"location_suggestions"` // Recent/common locations
    Inspectors         []User   `json:"inspectors"`           // Available inspectors
}

// Reports page data
type ReportsData struct {
    AppData
    Reports    []Report      // Generated reports
    Filter     ReportFilter  // Filter options
    Pagination ReportPagination // Pagination
}

// Generated report
type Report struct {
    ID          string    `json:"id"`           // "report_123"
    ProjectID   string    `json:"project_id"`   // Associated project
    ProjectName string    `json:"project_name"`
    Title       string    `json:"title"`        // Report title
    Type        string    `json:"type"`         // "inspection", "compliance", "summary"
    Status      string    `json:"status"`       // "generating", "completed", "failed"
    GeneratedAt time.Time `json:"generated_at"`
    GeneratedBy string    `json:"generated_by"` // User ID
    FileURL     string    `json:"file_url"`     // URL to download PDF
    FileSize    int64     `json:"file_size"`    // File size in bytes
}

// Report filtering
type ReportFilter struct {
    ProjectID string `json:"project_id"`
    Type      string `json:"type"`
    DateFrom  string `json:"date_from"`
    DateTo    string `json:"date_to"`
}

// Report pagination
type ReportPagination struct {
    CurrentPage  int `json:"current_page"`
    TotalPages   int `json:"total_pages"`
    TotalItems   int `json:"total_items"`
    ItemsPerPage int `json:"items_per_page"`
    HasPrev      bool `json:"has_prev"`
    HasNext      bool `json:"has_next"`
}

// Team management data
type TeamData struct {
    AppData
    TeamMembers []User        `json:"team_members"`
    Invitations []Invitation  `json:"invitations"` // Pending invitations
    Roles       []Role        `json:"roles"`       // Available roles
}

// Team invitation
type Invitation struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    InvitedBy string    `json:"invited_by"`    // User ID
    InvitedAt time.Time `json:"invited_at"`
    ExpiresAt time.Time `json:"expires_at"`
    Status    string    `json:"status"`        // "pending", "accepted", "expired"
}

// User role definition
type Role struct {
    ID          string   `json:"id"`          // "admin", "inspector", "viewer"
    Name        string   `json:"name"`        // "Administrator", "Inspector", "Viewer"
    Description string   `json:"description"` // Role description
    Permissions []string `json:"permissions"` // List of permissions
}