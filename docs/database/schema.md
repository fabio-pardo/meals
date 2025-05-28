# Database Schema

## Overview

The Meals application uses PostgreSQL as the primary database with GORM as the ORM. All tables use GORM's standard conventions including soft deletes, timestamps, and auto-incrementing primary keys.

## Tables

### users
Primary user authentication and profile data linked to OAuth2 providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing user ID |
| created_at | TIMESTAMP | NOT NULL | Record creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| provider | VARCHAR | NOT NULL | OAuth2 provider (e.g., "google") |
| email | VARCHAR | UNIQUE, NOT NULL | User's email address |
| name | VARCHAR | NULL | Full name from OAuth2 |
| first_name | VARCHAR | NULL | First name from OAuth2 |
| last_name | VARCHAR | NULL | Last name from OAuth2 |
| nick_name | VARCHAR | NULL | Nickname from OAuth2 |
| description | VARCHAR | NULL | User description from OAuth2 |
| access_token | VARCHAR | NOT NULL | OAuth2 access token |
| access_token_secret | VARCHAR | NULL | OAuth2 access token secret |
| refresh_token | VARCHAR | NULL | OAuth2 refresh token |
| expires_at | TIMESTAMP | NOT NULL | Token expiration time |
| id_token | VARCHAR | NOT NULL | OAuth2 ID token |
| user_id | VARCHAR(50) | UNIQUE, NOT NULL | External OAuth2 user ID |
| user_type | VARCHAR(20) | DEFAULT 'customer' | User role: admin, driver, customer |

**Indexes:**
- `idx_users_email` (UNIQUE)
- `idx_users_user_id` (UNIQUE)
- `idx_users_deleted_at`

**Business Rules:**
- New users default to 'customer' type
- Email must be unique across all users
- OAuth2 user_id must be unique across all providers

### sessions
User authentication sessions for cookie-based auth.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing session ID |
| created_at | TIMESTAMP | NOT NULL | Session creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| token | VARCHAR | UNIQUE, NOT NULL | Session token |
| expires_at | TIMESTAMP | NOT NULL | Session expiration time |
| user_identifier | VARCHAR(50) | NOT NULL | References users.user_id |

**Indexes:**
- `idx_sessions_token` (UNIQUE)
- `idx_sessions_user_identifier`
- `idx_sessions_deleted_at`

**Foreign Keys:**
- `user_identifier` → `users.user_id` (CASCADE UPDATE, CASCADE DELETE)

**Business Rules:**
- Sessions automatically expire based on expires_at
- Expired sessions are cleaned up periodically
- One user can have multiple active sessions

### user_profiles
Extended user information including delivery addresses and preferences.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing profile ID |
| created_at | TIMESTAMP | NOT NULL | Record creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| user_id | INTEGER | NOT NULL | References users.id |

**Indexes:**
- `idx_user_profiles_user_id`
- `idx_user_profiles_deleted_at`

**Foreign Keys:**
- `user_id` → `users.id` (NO CASCADE - profiles persist if user is soft deleted)

**Business Rules:**
- One-to-one relationship with users
- Profiles are optional and created on-demand
- Additional fields can be added for driver-specific data

### meals
Individual meal definitions with pricing and details.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing meal ID |
| created_at | TIMESTAMP | NOT NULL | Record creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| name | VARCHAR(255) | NOT NULL | Meal name |
| price | DECIMAL | NOT NULL | Meal price |

**Indexes:**
- `idx_meals_name`
- `idx_meals_deleted_at`

**Business Rules:**
- Meal names should be descriptive
- Prices are stored as decimal for accuracy
- Soft delete preserves meal history in orders

### menus
Weekly meal collections that group multiple meals together.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing menu ID |
| created_at | TIMESTAMP | NOT NULL | Record creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| name | VARCHAR | NOT NULL | Menu name |
| description | VARCHAR | NULL | Menu description |
| week_start_date | DATE | NOT NULL | Start date of menu week |
| week_end_date | DATE | NOT NULL | End date of menu week |

**Indexes:**
- `idx_menus_week_start_date`
- `idx_menus_week_end_date`
- `idx_menus_deleted_at`

**Business Rules:**
- Week end date must be after start date
- Menus typically span 7 days
- Multiple menus can exist for different weeks

### menu_meals
Junction table linking menus to meals with delivery day information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-incrementing ID |
| created_at | TIMESTAMP | NOT NULL | Record creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| delivery_day | VARCHAR(20) | NOT NULL | Day of week for delivery |
| menu_id | INTEGER | NOT NULL | References menus.id |
| meal_id | INTEGER | NOT NULL | References meals.id |

**Indexes:**
- `idx_menu_meals_menu_id`
- `idx_menu_meals_meal_id`
- `idx_menu_meals_delivery_day`
- `idx_menu_meals_deleted_at`

**Foreign Keys:**
- `menu_id` → `menus.id` (CASCADE UPDATE, CASCADE DELETE)
- `meal_id` → `meals.id` (RESTRICT DELETE, CASCADE UPDATE)

**Business Rules:**
- Same meal can appear multiple times in a menu for different days
- Delivery day should be a valid day of the week
- Deleting a menu cascades to menu_meals
- Deleting a meal is restricted if referenced in menu_meals

## Relationships

### User → Session (One-to-Many)
- One user can have multiple active sessions
- Sessions are automatically cleaned up on user deletion
- Foreign key: `sessions.user_identifier` → `users.user_id`

### User → UserProfile (One-to-One)
- Optional profile for extended user information
- Profile persists even if user is soft deleted
- Foreign key: `user_profiles.user_id` → `users.id`

### Menu → MenuMeal (One-to-Many)
- One menu contains multiple meal assignments
- Menu deletion cascades to menu_meals
- Foreign key: `menu_meals.menu_id` → `menus.id`

### Meal → MenuMeal (One-to-Many)
- One meal can be used in multiple menus
- Meal deletion is restricted if referenced
- Foreign key: `menu_meals.meal_id` → `meals.id`

## Data Integrity

### Soft Deletes
All tables use GORM's soft delete functionality:
- Records are marked with `deleted_at` timestamp instead of being physically deleted
- Soft deleted records are excluded from normal queries
- Use `.Unscoped()` to include soft deleted records

### Timestamps
All tables automatically maintain:
- `created_at`: Set on record creation
- `updated_at`: Updated on every save operation
- `deleted_at`: Set when record is soft deleted

### Constraints
- **Unique Constraints**: Enforced at database level for emails and tokens
- **Foreign Key Constraints**: Maintain referential integrity
- **NOT NULL Constraints**: Ensure required fields are always populated

## Indexing Strategy

### Primary Indexes
- All tables have auto-incrementing primary keys
- Primary keys are automatically indexed

### Foreign Key Indexes
- All foreign key columns are indexed for join performance
- Composite indexes on frequently queried combinations

### Business Logic Indexes
- `users.email` - Unique index for login lookups
- `sessions.token` - Unique index for session validation
- `menus.week_start_date` - Range queries for menu selection
- `menu_meals.delivery_day` - Filtering meals by delivery day

## Migration Strategy

### Schema Evolution
- Use GORM AutoMigrate for development
- Production migrations should be explicit SQL scripts
- Always backup before schema changes
- Test migrations on staging environment first

### Data Migration
- Preserve existing data during schema changes
- Use transactions for multi-table migrations
- Validate data integrity after migrations
- Plan rollback procedures for failed migrations

## Performance Considerations

### Query Optimization
- Use appropriate indexes for common query patterns
- Preload related data to avoid N+1 queries
- Use pagination for large result sets
- Monitor slow query logs

### Connection Management
- Configure appropriate connection pool size
- Set reasonable connection timeouts
- Monitor connection usage patterns
- Use read replicas for read-heavy workloads

### Caching Strategy
- Redis for session storage
- Consider caching frequently accessed meals/menus
- Cache user profile data for authenticated requests
- Implement cache invalidation strategies 