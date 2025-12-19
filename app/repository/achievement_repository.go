package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IAchievementRepository interface {
	Create(ctx context.Context, achRef *model.AchievementReference, achMongo *model.AchievementMongo) error
	GetRefByID(ctx context.Context, id string) (*model.AchievementReference, error)
	AddAttachment(ctx context.Context, mongoID string, attachment model.AchievementAttachment) error
	GetAll(ctx context.Context, page, pageSize int, search, studentIDFilter, advisorIDFilter, statusFilter string) ([]model.AchievementListDTO, int64, error)
	GetDetailByID(ctx context.Context, id string) (*model.AchievementDetailDTO, error)
	Update(ctx context.Context, id string, mongoID string, req *model.UpdateAchievementRequest) error
    Submit(ctx context.Context, id string) error
    SoftDelete(ctx context.Context, id string) error
	Verify(ctx context.Context, id string, lecturerID string, points int) error
    Reject(ctx context.Context, id string, lecturerID string, note string) error
}

type achievementRepository struct {
	pgDB    *sql.DB
	mongoDB *mongo.Database
}

func NewAchievementRepository(pgDB *sql.DB, mongoDB *mongo.Database) IAchievementRepository {
	return &achievementRepository{
		pgDB:    pgDB,
		mongoDB: mongoDB,
	}
}

// Create Achievement (Mongo & Postgres)
func (r *achievementRepository) Create(ctx context.Context, achRef *model.AchievementReference, achMongo *model.AchievementMongo) error {
	collection := r.mongoDB.Collection("achievements")

	result, err := collection.InsertOne(ctx, achMongo)
	if err != nil {
		return err
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("gagal mendapatkan object id")
	}

	achMongo.ID = oid
	achRef.MongoAchievementID = oid.Hex()

	query := `
        INSERT INTO achievement_references (student_id, mongo_achievement_id, status)
        VALUES ($1, $2, 'draft')
        RETURNING id, created_at, updated_at
    `

	err = r.pgDB.QueryRowContext(ctx, query, achRef.StudentID, achRef.MongoAchievementID).Scan(
		&achRef.ID, &achRef.CreatedAt, &achRef.UpdatedAt,
	)

	if err != nil {
		_, _ = collection.DeleteOne(ctx, bson.M{"_id": oid})
		return err
	}

	return nil
}

// Update Achievement
func (r *achievementRepository) Update(ctx context.Context, id string, mongoID string, req *model.UpdateAchievementRequest) error {
    query := `UPDATE achievement_references SET updated_at = NOW() WHERE id = $1`
    _, err := r.pgDB.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    oid, _ := primitive.ObjectIDFromHex(mongoID)
    
    updateFields := bson.M{"updated_at": time.Now()}
    if req.Title != nil { updateFields["title"] = *req.Title }
    if req.Description != nil { updateFields["description"] = *req.Description }
    if req.Tags != nil { updateFields["tags"] = req.Tags }
    if req.Details != nil { updateFields["details"] = req.Details }

    _, err = r.mongoDB.Collection("achievements").UpdateOne(
        ctx,
        bson.M{"_id": oid},
        bson.M{"$set": updateFields},
    )
    return err
}

// Get Reference
func (r *achievementRepository) GetRefByID(ctx context.Context, id string) (*model.AchievementReference, error) {
	query := `SELECT id, student_id, mongo_achievement_id, status FROM achievement_references WHERE id = $1`

	var ach model.AchievementReference
	err := r.pgDB.QueryRowContext(ctx, query, id).Scan(&ach.ID, &ach.StudentID, &ach.MongoAchievementID, &ach.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ach, nil
}

// Add Attachment 
func (r *achievementRepository) AddAttachment(ctx context.Context, mongoID string, attachment model.AchievementAttachment) error {
	collection := r.mongoDB.Collection("achievements")

	oid, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid}
	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

// GetAllAchievement 
func (r *achievementRepository) GetAll(ctx context.Context, page, pageSize int, search, studentIDFilter, advisorIDFilter, statusFilter string) ([]model.AchievementListDTO, int64, error) {
	offset := (page - 1) * pageSize

	baseQuery := `
        FROM achievement_references ar
        JOIN students s ON ar.student_id = s.id
        JOIN users u ON s.user_id = u.id
        WHERE 1=1
    `
	var args []interface{}
	argCounter := 1

	if advisorIDFilter != "" {
		baseQuery += fmt.Sprintf(" AND s.advisor_id = $%d", argCounter)
		args = append(args, advisorIDFilter)
		argCounter++
	}

	if studentIDFilter != "" {
		baseQuery += fmt.Sprintf(" AND ar.student_id = $%d", argCounter)
		args = append(args, studentIDFilter)
		argCounter++
	}

	if statusFilter != "" {
		baseQuery += fmt.Sprintf(" AND ar.status = $%d", argCounter)
		args = append(args, statusFilter)
		argCounter++
	}

	if search != "" {
		baseQuery += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR s.student_id ILIKE $%d)", argCounter, argCounter)
		args = append(args, "%"+search+"%")
		argCounter++
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.pgDB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := `
        SELECT ar.id, ar.mongo_achievement_id, ar.status, ar.created_at, 
               s.student_id, u.full_name 
    ` + baseQuery + fmt.Sprintf(" ORDER BY ar.created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)

	args = append(args, pageSize, offset)

	rows, err := r.pgDB.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var achievements []model.AchievementListDTO
	var mongoIDs []primitive.ObjectID
	mongoMap := make(map[string]int)

	for rows.Next() {
		var a model.AchievementListDTO
		var mongoIDStr string

		if err := rows.Scan(&a.ID, &mongoIDStr, &a.Status, &a.CreatedAt, &a.StudentID, &a.StudentName); err != nil {
			return nil, 0, err
		}

		a.MongoID = mongoIDStr
		achievements = append(achievements, a)

		if oid, err := primitive.ObjectIDFromHex(mongoIDStr); err == nil {
			mongoIDs = append(mongoIDs, oid)
			mongoMap[mongoIDStr] = len(achievements) - 1
		}
	}

	if len(mongoIDs) == 0 {
		return achievements, total, nil
	}

	cursor, err := r.mongoDB.Collection("achievements").Find(ctx, bson.M{"_id": bson.M{"$in": mongoIDs}})
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var m model.AchievementMongo
		if err := cursor.Decode(&m); err != nil {
			continue
		}

		if idx, ok := mongoMap[m.ID.Hex()]; ok {
			achievements[idx].Title = m.Title
			achievements[idx].AchievementType = m.AchievementType
			achievements[idx].Points = m.Points
		}
	}

	return achievements, total, nil
}

// GetDetailByID 
func (r *achievementRepository) GetDetailByID(ctx context.Context, id string) (*model.AchievementDetailDTO, error) {
	query := `
        SELECT ar.id, ar.mongo_achievement_id, ar.status, ar.rejection_note, ar.created_at, ar.updated_at,
               s.id, s.student_id, u.full_name, u.email, s.program_study, s.academic_year, s.advisor_id
        FROM achievement_references ar
        JOIN students s ON ar.student_id = s.id
        JOIN users u ON s.user_id = u.id
        WHERE ar.id = $1
    `
	var d model.AchievementDetailDTO
	var mongoIDStr string
	var rejectionNote, advisorID sql.NullString

	err := r.pgDB.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &mongoIDStr, &d.Status, &rejectionNote, &d.CreatedAt, &d.UpdatedAt,
		&d.Student.ID, &d.Student.StudentID, &d.Student.FullName, &d.Student.Email, &d.Student.ProgramStudy, &d.Student.AcademicYear, &advisorID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if rejectionNote.Valid {
		d.RejectionNote = &rejectionNote.String
	}
	if advisorID.Valid {
		d.Student.AdvisorID = &advisorID.String
	}

	oid, err := primitive.ObjectIDFromHex(mongoIDStr)
	if err != nil {
		return nil, err
	}

	var m model.AchievementMongo
	err = r.mongoDB.Collection("achievements").FindOne(ctx, bson.M{"_id": oid}).Decode(&m)
	if err != nil {
		return nil, err
	}

	d.AchievementType = m.AchievementType
	d.Title = m.Title
	d.Description = m.Description
	d.Details = m.Details
	d.Tags = m.Tags
	d.Attachments = m.Attachments

	return &d, nil
}

// Submit Achievement
func (r *achievementRepository) Submit(ctx context.Context, id string) error {
    query := `
        UPDATE achievement_references 
        SET status = 'submitted', 
            submitted_at = NOW(), 
            updated_at = NOW() 
        WHERE id = $1
    `
    _, err := r.pgDB.ExecContext(ctx, query, id)
    return err
}

// Soft Delete 
func (r *achievementRepository) SoftDelete(ctx context.Context, id string) error {
    query := `
        UPDATE achievement_references 
        SET status = 'deleted', 
            updated_at = NOW() 
        WHERE id = $1
    `
    _, err := r.pgDB.ExecContext(ctx, query, id)
    return err
}

// Verify Achievement
func (r *achievementRepository) Verify(ctx context.Context, id string, lecturerID string, points int) error {
    query := `
        UPDATE achievement_references 
        SET status = 'verified', 
            verified_at = NOW(), 
            verified_by = $1,
            updated_at = NOW()
        WHERE id = $2
        RETURNING mongo_achievement_id
    `
    
    var mongoIDStr string
    err := r.pgDB.QueryRowContext(ctx, query, lecturerID, id).Scan(&mongoIDStr)
    if err != nil {
        return err
    }

    oid, err := primitive.ObjectIDFromHex(mongoIDStr)
    if err != nil {
        return err
    }

    collection := r.mongoDB.Collection("achievements")
    filter := bson.M{"_id": oid}
    update := bson.M{
        "$set": bson.M{
            "points":     points,
            "updated_at": time.Now(),
        },
    }

    result, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }

    fmt.Printf("✅ MongoDB Updated - MatchedCount: %d, ModifiedCount: %d\n", result.MatchedCount, result.ModifiedCount)
    return nil
}

// Reject
func (r *achievementRepository) Reject(ctx context.Context, id string, lecturerID string, note string) error {
    query := `
        UPDATE achievement_references 
        SET status = 'rejected', 
            rejection_note = $1,
            verified_at = NOW(),
            verified_by = $2,
            updated_at = NOW()
        WHERE id = $3
    `
    _, err := r.pgDB.ExecContext(ctx, query, note, lecturerID, id)
    return err
}