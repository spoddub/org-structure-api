package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"org-structure-api/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	replacer := strings.NewReplacer("/", "_", " ", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", replacer.Replace(t.Name()))

	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := database.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	require.NoError(t, database.AutoMigrate(&model.Department{}, &model.Employee{}))

	return database
}

func TestDepartmentRepositoryCreateGetExistsAndListChildren(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)
	require.NotZero(t, root.ID)
	require.Equal(t, "Engineering", root.Name)
	require.Nil(t, root.ParentID)

	exists, err := repo.Exists(ctx, root.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.Exists(ctx, 999)
	require.NoError(t, err)
	require.False(t, exists)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)
	require.Equal(t, root.ID, *child.ParentID)

	got, err := repo.GetByID(ctx, root.ID)
	require.NoError(t, err)
	require.Equal(t, root.ID, got.ID)
	require.Equal(t, "Engineering", got.Name)

	children, err := repo.ListChildren(ctx, root.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	require.Equal(t, child.ID, children[0].ID)
	require.Equal(t, "Backend", children[0].Name)
}

func TestDepartmentRepositoryUpdate(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)

	child.Name = "Backend Platform"
	child.ParentID = nil

	updated, err := repo.Update(ctx, child)
	require.NoError(t, err)

	require.Equal(t, child.ID, updated.ID)
	require.Equal(t, "Backend Platform", updated.Name)
	require.Nil(t, updated.ParentID)
}

func TestDepartmentRepositoryIsDescendant(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)

	grandChild, err := repo.Create(ctx, "Platform", &child.ID)
	require.NoError(t, err)

	isDescendant, err := repo.IsDescendant(ctx, root.ID, grandChild.ID)
	require.NoError(t, err)
	require.True(t, isDescendant)

	isDescendant, err = repo.IsDescendant(ctx, root.ID, child.ID)
	require.NoError(t, err)
	require.True(t, isDescendant)

	isDescendant, err = repo.IsDescendant(ctx, child.ID, root.ID)
	require.NoError(t, err)
	require.False(t, isDescendant)
}
