package maria

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mpkondrashin/certlist/pkg/model"
)

func TestMaria(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file path")
	}
	testDir := filepath.Dir(filename)
	tempDir := filepath.Join(testDir, "testing")
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to remove temp dir: %v", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	mariaDistib := "../../cmd/certlist/mariadb-11.7.2-winx64.zip"
	mariaDB := NewDB(mariaDistib, tempDir)
	t.Log("Extract database")
	if err := mariaDB.Extract(); err != nil {
		t.Logf("Extract: %v", err)
	}
	t.Log("Initialize database")
	if err := mariaDB.Init(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Log("Start database")
	if err := mariaDB.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		t.Log("Stop database")
		if err := mariaDB.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	time.Sleep(2 * time.Second)
	t.Logf("Connect to the database")
	db, err := mariaDB.Open("")
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if err := CreateDatabase(db); err != nil {
		t.Fatalf("Create Database: %v", err)
	}
	t.Logf("Close database")
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	t.Logf("Open database %s", DatabaseName)
	db, err = mariaDB.Open(DatabaseName)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Logf("Populate database")
	dumpFile := "noalerts.mysqldump"
	if err := mariaDB.Populate(dumpFile); err != nil {
		t.Fatalf("Populate database: %v", err)
	}
	for tpt, err := range model.RangeTptDevice(db, "") {
		if err != nil {
			t.Fatalf("RangeTptDevice: %v", err)
		}
		t.Logf("Tpt: %s", tpt.DisplayName.String)
	}
}
