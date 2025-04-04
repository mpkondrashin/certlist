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
	if runtime.GOOS == "darwin" {
		db.mariadbExe = "/opt/homebrew/bin/mariadb"
		db.mariadbdExe = "/opt/homebrew/bin/mariadbd"
		db.mariadbInstallDbExe = "/opt/homebrew/bin/mariadb-install-db"
	}
	return nil
}

func (db *DB) Init() error {
	c := exec.Command(db.mariadbInstallDbExe, "--datadir="+db.DataFolder())
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
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
	}
	//log.Println(db.mariadbdExe, strings.Join(options, " "))
	db.cmd = exec.Command(db.mariadbdExe, options...)
	db.cmd.Stdout = os.Stdout
	db.cmd.Stderr = os.Stderr
	return db.cmd.Start()
}

func (db *DB) Stop() error {
	err := db.cmd.Process.Kill()
	if err != nil {
		return err
	}
	return db.cmd.Wait()
}

/*
	func (db *DB) Socket() string {
		if runtime.GOOS == "windows" {
			return "npipe://" + Pipe
		}
		return filepath.Join(db.tempFolder, SocketName)
	}
*/
func (db *DB) Open(databasename string) (*sql.DB, error) {
	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%s)/%s", Port, databasename)
	return sql.Open("mysql", dsn)
}

var ErrPopulate = errors.New("failed to populate MariaDB")

func (db *DB) Populate(dumpFile string) error {
	command := []string{
		"-u=root",
		"--skip-password",
		"--host=127.0.0.1",
		"--port=" + Port,
		DatabaseName,
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
		return fmt.Errorf("%w: %s", ErrPopulate, errOutput.String())
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
	query := fmt.Sprintf("CREATE DATABASE %s", DatabaseName)
	_, err := db.Exec(query)
	return err
}
