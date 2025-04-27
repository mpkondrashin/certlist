package maria

import (
	"archive/zip"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DatabaseName = "sms"
	MariaDBZip   = "mariadb-latest.zip" //mariadb-11.7.2-winx64.zip"
	Port         = "33060"
)

type DB struct {
	distribPath         string
	tempFolder          string
	mariadbExe          string
	mariadbdExe         string
	mariadbInstallDbExe string
	cmd                 *exec.Cmd
}

func NewDB(distribPath, tempFolder string) *DB {
	return &DB{
		distribPath: distribPath,
		tempFolder:  tempFolder,
	}
}
func (db *DB) DataFolder() string {
	return filepath.Join(db.tempFolder, "data")
}

func (db *DB) MariaFolder() string {
	return filepath.Join(db.tempFolder, "maria")
}

func (db *DB) Extract() error {
	if runtime.GOOS == "darwin" {
		db.mariadbExe = "/opt/homebrew/bin/mariadb"
		db.mariadbdExe = "/opt/homebrew/bin/mariadbd"
		db.mariadbInstallDbExe = "/opt/homebrew/bin/mariadb-install-db"
		return nil
	}

	if err := os.MkdirAll(db.MariaFolder(), 0755); err != nil {
		return fmt.Errorf("failed to create folder for maria: %w", err)
	}
	paths, err := Unzip(db.distribPath, db.MariaFolder(), []string{
		"mysql.exe",
		"mysqld.exe",
		"mysql_install_db.exe",
	})
	/*
			"mariadb.exe",
		"mariadbd.exe",
		"mariadb-install-db.exe",

	*/
	if err != nil {
		return fmt.Errorf("failed to unzip MariaDB: %w", err)
	}
	if paths[0] == "" {
		return fmt.Errorf("failed to find mariadb.exe in the zip file")
	}
	db.mariadbExe = paths[0]
	if paths[1] == "" {
		return fmt.Errorf("failed to find mariadb.exe in the zip file")
	}
	db.mariadbdExe = paths[1]
	if paths[2] == "" {
		return fmt.Errorf("failed to find mariadb-install-db.exe in the zip file")
	}
	db.mariadbInstallDbExe = paths[2]
	return nil
}

func (db *DB) Init() error {
	c := exec.Command(db.mariadbInstallDbExe, "--datadir="+db.DataFolder())
	var sb strings.Builder
	c.Stdout = &sb
	c.Stderr = &sb
	err := c.Run()
	if err != nil {
		return fmt.Errorf("failed to init MariaDB: %w\n%s %s\n%s", err, db.mariadbInstallDbExe, "--datadir="+db.DataFolder(), sb.String())
	}
	return nil
}

func (db *DB) Start() error {
	options := []string{
		"--datadir=" + db.DataFolder(),
		"--bind-address=127.0.0.1",
		"--port=" + Port,
		"--log-error=" + filepath.Join(db.tempFolder, "mariadb_error.log"),
		"--pid-file=" + filepath.Join(db.tempFolder, "mariadb.pid"),
		"--skip-grant-tables",
		"--console",
		"--silent-startup",
	}
	//log.Println(db.mariadbdExe, strings.Join(options, " "))
	db.cmd = exec.Command(db.mariadbdExe, options...)
	var sb strings.Builder
	db.cmd.Stdout = &sb
	db.cmd.Stderr = &sb
	err := db.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start MariaDB: %w\n%s %s\n%s", err, db.mariadbdExe, strings.Join(options, " "), sb.String())
	}
	return nil
}

func (db *DB) Stop() error {
	err := db.cmd.Process.Kill()
	if err != nil {
		return err
	}
	return db.cmd.Wait()
}

func (db *DB) Open(databasename string) (*sql.DB, error) {
	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%s)/%s", Port, databasename)
	return sql.Open("mysql", dsn)
}

var ErrPopulate = errors.New("failed to populate MariaDB")

func (db *DB) Populate(dumpFile string, databaseName string) error {
	command := []string{
		"-u=root",
		"--skip-password",
		"--host=127.0.0.1",
		"--port=" + Port,
		databaseName,
	}
	//log.Println(db.mariadbExe, strings.Join(command, " "))
	c := exec.Command(db.mariadbExe, command...)
	in, err := os.Open(dumpFile)
	if err != nil {
		return fmt.Errorf("failed to open dump file: %w", err)
	}
	defer in.Close()
	var errOutput strings.Builder
	c.Stdin = in
	c.Stdout = os.Stdout
	c.Stderr = &errOutput //os.Stderr
	if err := c.Run(); err != nil {
		if strings.Contains(errOutput.String(), DatabaseName+".alerts") {
			return nil
		}
		return fmt.Errorf("%s %s\n%w: %s", db.mariadbExe, strings.Join(command, " "), ErrPopulate, errOutput.String())
	}
	return nil
}

func Unzip(src, dest string, searchFor []string) (paths []string, err error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	paths = make([]string, len(searchFor))
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		for i, each := range searchFor {
			if strings.EqualFold(filepath.Base(f.Name), each) {
				paths[i] = fpath
			}
		}
		// Prevent Zip Slip (https://snyk.io/research/zip-slip-vulnerability)
		if !filepath.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid file path: %s", fpath)
		}
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return nil, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return nil, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return nil, err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return nil, err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return nil, err
		}
	}
	return
}

func CreateDatabase(db *sql.DB) error {
	query := fmt.Sprintf("DROP DATABASE %s", DatabaseName)
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("CREATE DATABASE %s", DatabaseName)
	_, err = db.Exec(query)
	return err
}
