package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mpkondrashin/certlist/pkg/maria"
	"github.com/mpkondrashin/certlist/pkg/prompt"

	"github.com/mpkondrashin/certalert/pkg/secureftp"
	"github.com/mpkondrashin/certalert/pkg/sms"
	"github.com/mpkondrashin/certlist/pkg/smsbackup"
)

const (
	DefaultUsernameLength = 16
	DefaultPasswordLength = 16
)

const (
	EnvPrefix = "CERTLIST"
)

const (
	ConfigFileName = "config"
	ConfigFileType = "yaml"
)

const (
	flagTempDir = "temp"
	flagOutput  = "output"

	flagSMSAddress         = "sms.address"
	flagSMSAPIKey          = "sms.api_key"
	flagSMSIgnoreTLSErrors = "sms.ignore_tls_errors"

	flagSFTPUsernameLength = "sftp.username_length"
	flagSFTPPasswordLength = "sftp.password_length"

	flagMariaDB   = "debug.mariadb"
	flagBackup    = "debug.backup"
	flagNoCleanup = "debug.nocleanup"
)

func Configure() {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)

	fs.String(flagTempDir, "", "Folder for temporary files")
	fs.String(flagOutput, "", "Output filename")

	fs.String(flagSMSAddress, "", "Tipping Point SMS address")
	fs.String(flagSMSAPIKey, "", "Tipping Point SMS API Key")
	fs.Bool(flagSMSIgnoreTLSErrors, false, "Ignore SMS TLS errors")

	fs.Int(flagSFTPUsernameLength, DefaultUsernameLength, "sFTP username length")
	fs.Int(flagSFTPPasswordLength, DefaultPasswordLength, "sFTP password length")

	fs.String(flagMariaDB, maria.MariaDBZip, "MariaDB ZIP file")
	fs.String(flagBackup, "", "SMS Backup File")
	fs.Bool(flagNoCleanup, false, "Keep temporary folder")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlags(fs); err != nil {
		log.Fatal(err)
	}
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)
	path, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(path)
		viper.AddConfigPath(dir)
	}
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			log.Fatal(err)
		}
	}
	err = prompt.Mandatory(fs, flagOutput, flagSMSAddress, flagSMSAPIKey)
	if err != nil {
		log.Fatal(err)
	}
	if viper.GetString(flagOutput) == "" {
		Panic("missing %s", flagOutput)
	}
	if viper.GetString(flagSMSAddress) == "" {
		Panic("missing %s", flagSMSAddress)
	}
	if viper.GetString(flagSMSAPIKey) == "" {
		Panic("missing %s", flagSMSAPIKey)
	}
}

func GetSMS() *sms.SMS {
	auth := sms.NewAPIKeyAuthorization(viper.GetString(flagSMSAPIKey))
	smsClient := sms.New("https://"+viper.GetString(flagSMSAddress), auth)
	return smsClient.SetInsecureSkipVerify(viper.GetBool(flagSMSIgnoreTLSErrors))
}

func GetLocalAddress() string {
	smsAddress := viper.GetString(flagSMSAddress)
	log.Printf("Dial SMS (%s)", smsAddress)
	localIP, err := GetOutboundIP(smsAddress + ":443")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SMS connection succeeded")
	log.Printf("Local address %v", localIP)
	return localIP.String()
}

func GetBackupFileName() string {
	backupBaseName := strings.ToLower(RandStringBytesRmndr(16))
	return backupBaseName + ".gz"
}

func FilterBackupPath(backupPath string) string {
	if runtime.GOOS != "windows" {
		return backupPath
	}
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	currentDrive := path[:2]
	if !strings.HasPrefix(backupPath, currentDrive) {
		Panic("TEMP is on %s drive and not on current drive: %s", backupPath[:2], currentDrive)
	}
	backupPath = backupPath[2:]
	return strings.ReplaceAll(backupPath, "\\", "/")
}

func RunBackup(smsClient *sms.SMS, username, password, localIP, backupPath string) {
	backupPath = FilterBackupPath(backupPath)
	//log.Printf("RunBackup(%v, %s, %s, %s, %s)", smsClient, username, password, localIP, backupPath)
	location := fmt.Sprintf("%s:%s", localIP, backupPath)
	password = url.QueryEscape(password)
	options := sms.NewBackupDatabaseOptionsSFTP(location, username, password)
	options.SetSSLPrivateKeys(false).SetTimestamp(false).SetEvents(false)
	log.Printf("Initiate backup: %v -> %s", smsClient, localIP)
	err := smsClient.BackupDatabase(options)
	if err != nil {
		Panic("backup database: %v", err)
	}
}

func GetOutboundIP(address string) (net.IP, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	return localAddr.IP, nil
}

func RandStringBytesRmndr(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func GetTempDir() string {
	tempDir := viper.GetString(flagTempDir)
	if tempDir == "" {
		var err error
		tempDir, err = os.MkdirTemp(viper.GetString(flagTempDir), "cl-*")
		if err != nil {
			Panic("TempDir: %v", err)
		}
	}
	log.Printf("Temp folder: %s", tempDir)
	return tempDir
}

func LogSize(backupPath string) {
	info, err := os.Stat(backupPath)
	if err != nil {
		Panic("stat: %v", err)
	}
	log.Printf("Got backup file: %s", formatFileSize(info.Size()))
}

func main() {
	log.Println("CertList Started")
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("%v %v", time.Now(), r)
			err := os.WriteFile("error.txt", []byte(message), 0664)
			if err != nil {
				log.Println(r)
			}
		}
		log.Println("Exiting")
	}()
	Configure()
	tempDir := GetTempDir()
	if !viper.GetBool(flagNoCleanup) {
		defer func() {
			log.Printf("Remove temporary folder %s", tempDir)
			err := os.RemoveAll(tempDir)
			if err != nil {
				log.Print(err)
			}
		}()
	}
	backupName := GetBackupFileName()
	backupPath := filepath.Join(tempDir, backupName)
	if viper.GetString(flagBackup) != "" {
		backupPath = viper.GetString(flagBackup)
	} else {
		localIP := GetLocalAddress()
		log.Printf("Run local sFTP server")
		port := 22
		username := RandStringBytesRmndr(viper.GetInt(flagSFTPUsernameLength))
		password := RandStringBytesRmndr(viper.GetInt(flagSFTPPasswordLength))
		go secureftp.Run(username, password, localIP, port)
		smsClient := GetSMS()
		log.Printf("Run backup")
		RunBackup(smsClient, username, password, localIP, backupPath)
	}
	LogSize(backupPath)
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	mariaDistib := filepath.Join(filepath.Dir(exePath), viper.GetString(flagMariaDB))
	mariaDB := maria.NewDB(mariaDistib, tempDir)
	log.Print("Extract MariaDB")
	if err := mariaDB.Extract(); err != nil {
		Panic("extract: %v", err)
	}
	log.Print("Initialize MariaDB")
	if err := mariaDB.Init(); err != nil {
		Panic("initialize: %v", err)
	}
	log.Print("Start MariaDB")
	if err := mariaDB.Start(); err != nil {
		Panic("start: %v", err)
	}
	defer func() {
		log.Print("Stop MariaDB")
		err := mariaDB.Stop()
		if err != nil {
			log.Print(err)
		}
	}()
	time.Sleep(2 * time.Second)
	log.Print("Extract dump")
	dumpFile, err := smsbackup.ExtractDump(backupPath)
	if err != nil {
		Panic("ExtractDump: %v", err)
	}
	log.Print("Connect to the MariaDB")
	db, err := mariaDB.Open("")
	if err != nil {
		Panic("connect: %v", err)
	}
	log.Print("Ping MariaDB")
	if err := db.Ping(); err != nil {
		Panic("Failed to ping database: %v", err)
	}
	log.Print("Create database")
	if err := maria.CreateDatabase(db); err != nil {
		Panic("create Database: %v", err)
	}
	log.Print("Close MariaDB connection")
	if err := db.Close(); err != nil {
		Panic("close database: %v", err)
	}
	log.Print("Populate database")
	if err = mariaDB.Populate(dumpFile); err != nil {
		Panic("populate database: %v", err)
	}
	log.Printf("Connect to database %s", maria.DatabaseName)
	db, err = mariaDB.Open(maria.DatabaseName)
	if err != nil {
		Panic("connect to %s: %v", maria.DatabaseName, err)
	}
	log.Print("Generate report")
	report, err := smsbackup.GenerateReport(db)
	if err != nil {
		Panic("GenerateReport: %v", err)
	}
	log.Print("Write report")
	if err := SaveCSV(viper.GetString(flagOutput), report); err != nil {
		Panic("SaveCSV: %v", err)
	}
	log.Printf("Report saved to %s", viper.GetString(flagOutput))
}

func Panic(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Println(msg)
	panic(msg)
}
