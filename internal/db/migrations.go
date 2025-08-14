package db

import (
	"context"
	"log"

	"github.com/mostafijurj/notification-service/internal/models"
	"gorm.io/gorm"
)

// RunMigrations applies GORM auto-migration and seeds initial data
func RunMigrations(ctx context.Context, db *gorm.DB) error {
	return RunMigrationsWithOptions(ctx, db, false)
}

func RunMigrationsWithOptions(ctx context.Context, db *gorm.DB, forceReset bool) error {
	log.Println("🔄 Running GORM auto-migrations...")

	// Set GORM to be more conservative with migrations
	db.Config.DisableForeignKeyConstraintWhenMigrating = true

	// Check if we're dealing with an existing database
	var tableCount int64
	db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)

	if forceReset && tableCount > 0 {
		log.Println("🗑️  Force reset requested, dropping all tables...")
		if err := dropAllTables(ctx, db); err != nil {
			log.Printf("⚠️  Warning: failed to drop tables: %v", err)
		}
		tableCount = 0
	}

	if tableCount > 0 {
		log.Println("📋 Existing database detected, using safe migration mode")
		// Use a more conservative approach for existing databases
		return runSafeMigrations(ctx, db)
	}

	// Fresh database - use standard AutoMigrate
	log.Println("🆕 Fresh database detected, using standard migration mode")
	return runStandardMigrations(ctx, db)
}

func dropAllTables(ctx context.Context, db *gorm.DB) error {
	// Drop tables in reverse dependency order
	tables := []string{
		"inapp_notifications",
		"delivery_attempts",
		"notifications",
		"device_tokens",
		"campaigns",
		"group_members",
		"groups",
		"user_dnd_windows",
		"user_channel_preferences",
		"templates",
		"notification_types",
	}

	for _, table := range tables {
		if db.WithContext(ctx).Migrator().HasTable(table) {
			log.Printf("🗑️  Dropping table: %s", table)
			if err := db.WithContext(ctx).Migrator().DropTable(table); err != nil {
				log.Printf("⚠️  Warning: failed to drop table %s: %v", table, err)
			}
		}
	}

	return nil
}

func fixConstraintIssues(ctx context.Context, db *gorm.DB) error {
	log.Println("🔧 Checking for constraint issues...")

	// List of constraints that might cause issues
	problematicConstraints := []string{
		"uni_notification_types_key",
		"uni_templates_type_key_channel_locale_version",
		"uni_user_channel_preferences_user_type_channel",
		"uni_user_dnd_windows_user_id",
		"uni_groups_name",
		"uni_group_members_group_user",
		"uni_notifications_idempotency_key",
		"uni_device_tokens_provider_token",
	}

	for _, constraint := range problematicConstraints {
		// Check if constraint exists
		var exists bool
		db.WithContext(ctx).Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE constraint_name = ? AND table_schema = 'public'
			)`, constraint).Scan(&exists)

		if exists {
			log.Printf("⚠️  Warning: found problematic constraint: %s", constraint)
			// Try to drop it gracefully
			if err := db.WithContext(ctx).Exec("ALTER TABLE " + getTableNameForConstraint(constraint) + " DROP CONSTRAINT IF EXISTS " + constraint).Error; err != nil {
				log.Printf("⚠️  Warning: failed to drop constraint %s: %v", constraint, err)
			} else {
				log.Printf("✅ Successfully dropped constraint: %s", constraint)
			}
		}
	}

	return nil
}

func getTableNameForConstraint(constraint string) string {
	// Map constraints to their table names
	constraintMap := map[string]string{
		"uni_notification_types_key":                     "notification_types",
		"uni_templates_type_key_channel_locale_version":  "templates",
		"uni_user_channel_preferences_user_type_channel": "user_channel_preferences",
		"uni_user_dnd_windows_user_id":                   "user_dnd_windows",
		"uni_groups_name":                                "groups",
		"uni_group_members_group_user":                   "group_members",
		"uni_notifications_idempotency_key":              "notifications",
		"uni_device_tokens_provider_token":               "device_tokens",
	}

	if tableName, exists := constraintMap[constraint]; exists {
		return tableName
	}

	return "unknown_table"
}

func runStandardMigrations(ctx context.Context, db *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&models.NotificationType{},
		&models.Template{},
		&models.UserChannelPreference{},
		&models.UserDNDWindow{},
		&models.Group{},
		&models.GroupMember{},
		&models.Campaign{},
		&models.Notification{},
		&models.DeliveryAttempt{},
		&models.DeviceToken{},
		&models.InAppNotification{},
	}

	for _, model := range modelsToMigrate {
		if err := db.WithContext(ctx).AutoMigrate(model); err != nil {
			log.Printf("⚠️  Warning: failed to migrate %T: %v", model, err)
		}
	}

	log.Println("✅ Standard migrations completed")

	// Seed initial data
	if err := seedInitialData(ctx, db); err != nil {
		log.Printf("⚠️  Warning: failed to seed initial data: %v", err)
	}

	return nil
}

func runSafeMigrations(ctx context.Context, db *gorm.DB) error {
	// For existing databases, create tables only if they don't exist
	log.Println("🔒 Running safe migrations for existing database...")

	// First, fix any constraint issues
	if err := fixConstraintIssues(ctx, db); err != nil {
		log.Printf("⚠️  Warning: failed to fix constraint issues: %v", err)
	}

	// Create tables one by one with error handling
	tables := []struct {
		name  string
		model interface{}
	}{
		{"notification_types", &models.NotificationType{}},
		{"templates", &models.Template{}},
		{"user_channel_preferences", &models.UserChannelPreference{}},
		{"user_dnd_windows", &models.UserDNDWindow{}},
		{"groups", &models.Group{}},
		{"group_members", &models.GroupMember{}},
		{"campaigns", &models.Campaign{}},
		{"notifications", &models.Notification{}},
		{"delivery_attempts", &models.DeliveryAttempt{}},
		{"device_tokens", &models.DeviceToken{}},
		{"inapp_notifications", &models.InAppNotification{}},
	}

	for _, table := range tables {
		if !db.WithContext(ctx).Migrator().HasTable(table.model) {
			log.Printf("📝 Creating table: %s", table.name)
			if err := db.WithContext(ctx).Migrator().CreateTable(table.model); err != nil {
				log.Printf("⚠️  Warning: failed to create table %s: %v", table.name, err)
			}
		} else {
			log.Printf("✅ Table already exists: %s", table.name)
		}
	}

	log.Println("✅ Safe migrations completed")

	// Seed initial data
	if err := seedInitialData(ctx, db); err != nil {
		log.Printf("⚠️  Warning: failed to seed initial data: %v", err)
	}

	return nil
}

// seedInitialData creates initial notification types and templates
func seedInitialData(ctx context.Context, db *gorm.DB) error {
	log.Println("🌱 Seeding initial data...")

	// Check if data already exists
	var count int64
	db.Model(&models.NotificationType{}).Count(&count)
	if count > 0 {
		log.Println("ℹ️  Initial data already exists, skipping seed")
		return nil
	}

	// Create notification types
	notificationTypes := []models.NotificationType{
		{Key: "otp", Category: "transactional"},
		{Key: "order_confirmation", Category: "transactional"},
		{Key: "password_reset", Category: "transactional"},
		{Key: "welcome", Category: "promotional"},
		{Key: "marketing", Category: "promotional"},
		{Key: "system_alert", Category: "system"},
		{Key: "maintenance", Category: "system"},
		{Key: "new_follower", Category: "activity"},
		{Key: "like_post", Category: "activity"},
	}

	if err := db.CreateInBatches(notificationTypes, 10).Error; err != nil {
		return err
	}

	// Create basic templates
	otpSubject := "Your OTP Code"
	welcomeSubject := "Welcome to Our Service!"
	orderSubject := "Order Confirmation"

	templates := []models.Template{
		{
			TypeKey:  "otp",
			Channel:  "email",
			Locale:   "en",
			Subject:  &otpSubject,
			Body:     "Your OTP code is: {{otp_code}}. Valid for 5 minutes.",
			Version:  1,
			IsActive: true,
		},
		{
			TypeKey:  "otp",
			Channel:  "sms",
			Locale:   "en",
			Body:     "Your OTP code is: {{otp_code}}. Valid for 5 minutes.",
			Version:  1,
			IsActive: true,
		},
		{
			TypeKey:  "welcome",
			Channel:  "email",
			Locale:   "en",
			Subject:  &welcomeSubject,
			Body:     "Hi {{user_name}}, welcome to our service! We're excited to have you on board.",
			Version:  1,
			IsActive: true,
		},
		{
			TypeKey:  "order_confirmation",
			Channel:  "email",
			Locale:   "en",
			Subject:  &orderSubject,
			Body:     "Your order #{{order_id}} has been confirmed. Total: {{amount}}",
			Version:  1,
			IsActive: true,
		},
	}

	if err := db.CreateInBatches(templates, 10).Error; err != nil {
		return err
	}

	// Create default groups
	premiumDesc := "Premium subscription users"
	merchantDesc := "Merchant accounts"
	adminDesc := "System administrators"

	groups := []models.Group{
		{Name: "Premium Users", Description: &premiumDesc},
		{Name: "Merchants", Description: &merchantDesc},
		{Name: "Administrators", Description: &adminDesc},
	}

	if err := db.CreateInBatches(groups, 10).Error; err != nil {
		return err
	}

	log.Println("✅ Initial data seeded successfully")
	return nil
}
