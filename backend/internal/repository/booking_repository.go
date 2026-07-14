package repository

import (
	"context"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		collection: db.Collection("bookings"),
	}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) (string, error) {
	res, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *BookingRepository) FindByID(ctx context.Context, bookingID string) (*models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, err
	}
	var booking models.Booking
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// FindAll คือ query สำหรับ Admin Dashboard พร้อม filter
// รับ filter เป็น bson.M ตรงๆ จาก handler ชั้นบน เพื่อให้ยืดหยุ่นต่อ requirement
// ข้อ 2.2 ("Filter อย่างน้อย 1 อย่าง เช่น by movie / date / user") โดยไม่ต้องแก้
// signature ของ function นี้ทุกครั้งที่อยากเพิ่ม filter ใหม่
func (r *BookingRepository) FindAll(ctx context.Context, filter bson.M, limit, skip int64) ([]models.Booking, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}
