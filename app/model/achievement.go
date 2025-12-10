package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementListDTO struct {
	ID              string    `json:"id"`                   
	MongoID         string    `json:"-"`                    
	StudentID       string    `json:"student_id"`
	StudentName     string    `json:"student_name"`         
	AchievementType string    `json:"achievement_type"`     
	Title           string    `json:"title"`                
	Status          string    `json:"status"`               
	Points          int       `json:"points"`               
	CreatedAt       time.Time `json:"created_at"`           
}

type AchievementDetailDTO struct {
	ID              string                 `json:"id"`
	Student         StudentListDTO         `json:"student"`       
	AchievementType string                 `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`      
	Tags            []string               `json:"tags"`
	Attachments     []AchievementAttachment`json:"attachments"`
	Status          string                 `json:"status"`
	RejectionNote   *string                `json:"rejection_note,omitempty"`
	VerifiedBy      *string                `json:"verified_by,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type CreateAchievementRequest struct {
	AchievementType string                 `json:"achievement_type" validate:"required"` 
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"` 
	Tags            []string               `json:"tags"`
}

type UpdateAchievementRequest struct {
	Title       *string                 `json:"title"`
	Description *string                 `json:"description"`
	Details     map[string]interface{}  `json:"details"`
	Tags        []string                `json:"tags"` 
	AchievementType *string				`json:"achievement_type"`
}

type AchievementMongo struct {
	ID              primitive.ObjectID      `bson:"_id,omitempty"`
	StudentID       string                  `bson:"student_id"` 
	AchievementType string                  `bson:"achievement_type"`
	Title           string                  `bson:"title"`
	Description     string                  `bson:"description"`
	Details         map[string]interface{}  `bson:"details"`
	Tags            []string                `bson:"tags"`
	Attachments     []AchievementAttachment `bson:"attachments"` 
	Points          int                     `bson:"points"`
	CreatedAt       time.Time               `bson:"created_at"`
	UpdatedAt       time.Time               `bson:"updated_at"`
}

type AchievementAttachment struct {
	FileName   string    `bson:"file_name" json:"file_name"`
	FileURL    string    `bson:"file_url" json:"file_url"`
	FileType   string    `bson:"file_type" json:"file_type"`
	UploadedAt time.Time `bson:"uploaded_at" json:"uploaded_at"`
}

type AchievementReference struct {
	ID                 string    `json:"id"`
	StudentID          string    `json:"student_id"`
	MongoAchievementID string    `json:"mongo_achievement_id"`
	Status             string    `json:"status"` 
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type VerifyAchievementRequest struct {
	Points int `json:"points" validate:"required,min=1"`
}

type RejectAchievementRequest struct {
	RejectionNote string `json:"rejection_note" validate:"required,min=5"`
}

type PaginatedAchievements struct {
	Data       []AchievementListDTO `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}