//go:build integration

package repository

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestDatabaseSnapshot_ResetRestoresSnapshotState(t *testing.T) {
	InitDeps()

	databaseSnapshotRepository := u.NewDatabaseSnapshotRepository("localhost", DatabaseConnector, DatabaseUtils)

	err := DatabaseConnector.Connect()
	if err != nil {
		t.Fatal(err)
	}

	_, err = DatabaseConnector.GetDB().Exec(`CREATE TABLE IF NOT EXISTS snapshot_test_items (item_id TEXT PRIMARY KEY)`)
	assert.Nil(t, err)

	_, err = DatabaseConnector.GetDB().Exec(`DELETE FROM snapshot_test_items`)
	assert.Nil(t, err)

	_, err = DatabaseConnector.GetDB().Exec(`INSERT INTO snapshot_test_items (item_id) VALUES ($1)`, "rowA")
	assert.Nil(t, err)

	assert.Nil(t, databaseSnapshotRepository.CreateDatabaseSnapshot())
	snapshotExists, err := databaseSnapshotRepository.SnapshotExists()
	assert.Nil(t, err)
	assert.True(t, snapshotExists)

	_, err = DatabaseConnector.GetDB().Exec(`DELETE FROM snapshot_test_items WHERE item_id = $1`, "rowA")
	assert.Nil(t, err)
	_, err = DatabaseConnector.GetDB().Exec(`INSERT INTO snapshot_test_items (item_id) VALUES ($1)`, "rowB")
	assert.Nil(t, err)

	assert.Nil(t, databaseSnapshotRepository.ResetDatabaseToSnapshot())

	var rowACount int
	err = databaseSnapshotRepository.DatabaseConnector.GetDB().QueryRow(`SELECT COUNT(*) FROM snapshot_test_items WHERE item_id = $1`, "rowA").Scan(&rowACount)
	assert.Nil(t, err)
	assert.Equal(t, 1, rowACount)

	var rowBCount int
	err = databaseSnapshotRepository.DatabaseConnector.GetDB().QueryRow(`SELECT COUNT(*) FROM snapshot_test_items WHERE item_id = $1`, "rowB").Scan(&rowBCount)
	assert.Nil(t, err)
	assert.Equal(t, 0, rowBCount)

	assert.Nil(t, databaseSnapshotRepository.DeleteDatabaseSnapshotIfExists())

	snapshotExists, err = databaseSnapshotRepository.SnapshotExists()
	assert.Nil(t, err)
	assert.False(t, snapshotExists)
}
