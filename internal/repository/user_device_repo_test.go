package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestUpsertUserDevice(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	fcmToken := "token123"
	deviceType := "ios"

	device, err := repo.UpsertUserDevice(ctx, userID, fcmToken, deviceType, userID)
	if err != nil {
		t.Fatalf("Failed to upsert device: %v", err)
	}

	if device.ID == uuid.Nil {
		t.Error("Device ID should not be nil")
	}
	if device.FCMToken != fcmToken {
		t.Errorf("FCMToken mismatch: got %v, want %v", device.FCMToken, fcmToken)
	}
}

func TestUpsertUserDeviceIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	fcmToken := "token123"
	deviceType := "ios"

	// First upsert
	device1, _ := repo.UpsertUserDevice(ctx, userID, fcmToken, deviceType, userID)

	// Second upsert with same token
	newDeviceType := "android"
	device2, err := repo.UpsertUserDevice(ctx, userID, fcmToken, newDeviceType, userID)
	if err != nil {
		t.Fatalf("Failed to upsert device: %v", err)
	}

	// Should be same ID
	if device1.ID != device2.ID {
		t.Errorf("ID should be the same: got %v, want %v", device2.ID, device1.ID)
	}
	// Device type should be updated
	if device2.DeviceType != newDeviceType {
		t.Errorf("DeviceType should be updated: got %v, want %v", device2.DeviceType, newDeviceType)
	}
}

func TestGetUserDeviceByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	fcmToken := "token456"
	deviceType := "ios"
	created, _ := repo.UpsertUserDevice(ctx, userID, fcmToken, deviceType, userID)

	device, err := repo.GetUserDeviceByID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}

	if device.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", device.ID, created.ID)
	}
}

func TestGetUserDeviceByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	_, err := repo.GetUserDeviceByID(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrUserDeviceNotFound) {
		t.Errorf("Expected ErrUserDeviceNotFound, but got %v", err)
	}
}

func TestGetUserDevicesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	token1 := "token1"
	token2 := "token2"
	deviceType := "ios"
	repo.UpsertUserDevice(ctx, userID, token1, deviceType, userID)
	repo.UpsertUserDevice(ctx, userID, token2, deviceType, userID)

	devices, err := repo.GetUserDevicesByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to get devices: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
}

func TestGetUserDeviceByFCMToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	fcmToken := "unique_token"
	deviceType := "ios"
	repo.UpsertUserDevice(ctx, userID, fcmToken, deviceType, userID)

	device, err := repo.GetUserDeviceByFCMToken(ctx, fcmToken, userID)
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}

	if device.FCMToken != fcmToken {
		t.Errorf("FCMToken mismatch: got %v, want %v", device.FCMToken, fcmToken)
	}
}

func TestDeleteUserDevice(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	fcmToken := "delete_token"
	deviceType := "ios"
	device, _ := repo.UpsertUserDevice(ctx, userID, fcmToken, deviceType, userID)
	deviceID := device.ID

	err := repo.DeleteUserDeviceByID(ctx, deviceID, userID)
	if err != nil {
		t.Fatalf("Failed to delete device: %v", err)
	}

	_, err = repo.GetUserDeviceByID(ctx, deviceID, userID)
	if !errors.Is(err, ErrUserDeviceNotFound) {
		t.Errorf("Expected ErrUserDeviceNotFound, but got %v", err)
	}
}

func TestDeleteUserDeviceNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	err := repo.DeleteUserDeviceByID(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrUserDeviceNotFound) {
		t.Errorf("Expected ErrUserDeviceNotFound, but got %v", err)
	}
}

func TestDeleteUserDevicesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserDeviceRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	token1 := "token1"
	token2 := "token2"
	deviceType := "ios"
	repo.UpsertUserDevice(ctx, userID, token1, deviceType, userID)
	repo.UpsertUserDevice(ctx, userID, token2, deviceType, userID)

	err := repo.DeleteUserDevicesByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to delete devices: %v", err)
	}

	devices, _ := repo.GetUserDevicesByUserID(ctx, userID, userID)
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices after deletion, got %d", len(devices))
	}
}
