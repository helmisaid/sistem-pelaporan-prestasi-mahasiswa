package model

import "time"

type DashboardStatistics struct {
	TotalStudents        int            `json:"total_students"`
	TotalLecturers       int            `json:"total_lecturers"`
	TotalAchievements    int            `json:"total_achievements"`
	AchievementsByStatus map[string]int `json:"achievements_by_status"`
	TopStudents          []TopStudent   `json:"top_students"`
}

type TopStudent struct {
	StudentID   string `json:"student_id"`
	Name        string `json:"name"`
	TotalPoints int    `json:"total_points"`
}

type StudentReportDTO struct {
	StudentProfile     StudentListDTO       `json:"student_profile"`
	TotalPoints        int                  `json:"total_points"`
	TotalAchievements  int                  `json:"total_achievements"`
	PointsByType       map[string]int       `json:"points_by_type"`
	RecentAchievements []AchievementListDTO `json:"recent_achievements"`
	GeneratedAt        time.Time            `json:"generated_at"`
}