# Database Migrations

This directory contains database migrations for the Digital Discovery project.

## Structure

```
migrations/
├── README.md                    # This file
├── backups/                     # Database backups
│   └── backup_YYYYMMDD_HHMMSS.sql
├── 000001_init_schema.up.sql   # Initial schema migration
├── 000001_init_schema.down.sql # Rollback for initial schema
└── ...                         # Future migrations
```

## Naming Convention

Migration files follow the naming pattern:
- `{version}_{description}.up.sql` - Forward migration
- `{version}_{description}.down.sql` - Rollback migration

Example:
- `000001_init_schema.up.sql`
- `000001_init_schema.down.sql`

## Migration Commands

### Basic Commands
```bash
# Apply all migrations
make migrate-up

# Rollback last migration
make migrate-down

# Apply one migration
make migrate-up-1

# Rollback one migration
make migrate-down-1

# Create new migration
make migrate-create

# Check migration status
make migrate-status
```

### Advanced Commands
```bash
# Verify migration files
make migrate-verify

# Fix SQL formatting
make migrate-fix

# Create database backup
make migrate-backup

# Open database shell
make db-shell

# Backup with metadata
make db-backup

# Restore from backup
make db-restore
```

## Best Practices

### 1. Migration Design
- **Atomic Changes**: Each migration should handle one logical change
- **Idempotent**: Migrations should be idempotent (can be run multiple times safely)
- **Self-Contained**: Don't rely on external scripts or data
- **Reversible**: Always provide a down migration
- **Backward Compatible**: Consider existing data and code

### 2. Naming and Organization
- Use descriptive names: `create_users_table`, `add_email_to_users`
- Keep migrations small and focused
- Use sequential numbering to maintain order
- Group related changes in a single migration

### 3. SQL Best Practices
```sql
-- Good Practice Example
BEGIN;

-- Add new column with a default value
ALTER TABLE users 
    ADD COLUMN email VARCHAR(255);

-- Update existing records
UPDATE users 
    SET email = 'unknown@example.com' 
    WHERE email IS NULL;

-- Add not-null constraint after data is updated
ALTER TABLE users 
    ALTER COLUMN email SET NOT NULL;

COMMIT;
```

### 4. Safety Measures
- Always test migrations in development first
- Create database backups before migrations
- Use transactions for complex migrations
- Add appropriate indexes before adding foreign keys
- Consider data volume when adding constraints

### 5. Performance Considerations
- Use batching for large data updates
- Add indexes after bulk data loading
- Consider running intensive migrations during off-peak hours
- Monitor lock times and transaction duration

### 6. Common Patterns

#### Adding Columns
```sql
-- Safe pattern for adding columns
ALTER TABLE table_name 
    ADD COLUMN new_column data_type,
    ADD COLUMN another_column data_type;
```

#### Modifying Columns
```sql
-- Safe pattern for modifying columns
-- 1. Add new column
ALTER TABLE users ADD COLUMN email_new VARCHAR(255);

-- 2. Copy data
UPDATE users SET email_new = email;

-- 3. Drop old column
ALTER TABLE users DROP COLUMN email;

-- 4. Rename new column
ALTER TABLE users RENAME COLUMN email_new TO email;
```

#### Adding Constraints
```sql
-- Safe pattern for adding constraints
-- 1. Add constraint as NOT VALID
ALTER TABLE users ADD CONSTRAINT users_email_check 
    CHECK (email ~* '^.+@.+\..+$') NOT VALID;

-- 2. Validate existing data
ALTER TABLE users VALIDATE CONSTRAINT users_email_check;
```

### 7. Troubleshooting
- Keep track of migration versions
- Use `migrate-status` to check current state
- Maintain backup before critical migrations
- Document any manual steps required

### 8. Version Control
- Never modify committed migrations
- Create new migrations for changes
- Include both up and down migrations
- Document breaking changes

### 9. Production Deployment
1. Always backup database first
2. Test migrations in staging environment
3. Plan for rollback scenarios
4. Monitor system during migration
5. Have database experts available during migration
6. Document deployment steps

### 10. Monitoring and Maintenance
- Regular cleanup of old backups
- Monitor migration performance
- Keep track of schema versions
- Document database changes

## Tools and Scripts

### Backup Script
- Located at `backup.sh`
- Creates timestamped backups
- Maintains backup history
- Includes metadata

### Restore Script
- Located at `restore.sh`
- Interactive restoration
- Safety checks
- Creates pre-restore backup

## Additional Resources
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [Database Migration Patterns](https://www.postgresql.org/docs/current/ddl.html) 