package repository

import (
	"context"
	"database/sql"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IReportRepository interface {
	GetGlobalStats(ctx context.Context) (*model.DashboardStatistics, error)
	GetStudentReport(ctx context.Context, studentID string) (*model.StudentReportDTO, error)
}

type reportRepository struct {
	pgDB    *sql.DB
	mongoDB *mongo.Database
}

func NewReportRepository(pgDB *sql.DB, mongoDB *mongo.Database) IReportRepository {
	return &reportRepository{pgDB, mongoDB}
}

// GetGlobalStats 
func (r *reportRepository) GetGlobalStats(ctx context.Context) (*model.DashboardStatistics, error) {
	stats := &model.DashboardStatistics{
		AchievementsByStatus: make(map[string]int),
	}

	r.pgDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM students").Scan(&stats.TotalStudents)
	r.pgDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM lecturers").Scan(&stats.TotalLecturers)

	rows, err := r.pgDB.QueryContext(ctx, "SELECT status, COUNT(*) FROM achievement_references GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalAch := 0
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats.AchievementsByStatus[status] = count
		totalAch += count
	}
	stats.TotalAchievements = totalAch

	return stats, nil
}

// GetStudentReport 
func (r *reportRepository) GetStudentReport(ctx context.Context, studentID string) (*model.StudentReportDTO, error) {
	report := &model.StudentReportDTO{
		PointsByType: make(map[string]int),
		GeneratedAt:  time.Now(),
	}

	query := `
        SELECT mongo_achievement_id 
        FROM achievement_references 
        WHERE student_id = $1 AND status = 'verified'
    `
	rows, err := r.pgDB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mongoIDs []primitive.ObjectID
	for rows.Next() {
		var mid string
		rows.Scan(&mid)
		if oid, err := primitive.ObjectIDFromHex(mid); err == nil {
			mongoIDs = append(mongoIDs, oid)
		}
	}

	if len(mongoIDs) == 0 {
		return report, nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": mongoIDs}}}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$achievement_type",
			"totalPoints": bson.M{"$sum": "$points"},
			"count":       bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.mongoDB.Collection("achievements").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalPoints := 0
	totalCount := 0

	for cursor.Next(ctx) {
		var result struct {
			Type        string `bson:"_id"`
			TotalPoints int    `bson:"totalPoints"`
			Count       int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			report.PointsByType[result.Type] = result.TotalPoints
			totalPoints += result.TotalPoints
			totalCount += result.Count
		}
	}

	report.TotalPoints = totalPoints
	report.TotalAchievements = totalCount

	return report, nil
}
