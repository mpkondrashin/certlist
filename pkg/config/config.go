package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mpkondrashin/certlist/pkg/maria"
	"github.com/mpkondrashin/certlist/pkg/prompt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	TempDir = "temp"

	OutputFilename  = "output.filename"
	OutputStrict    = "output.strict"
	OutputSemicolon = "output.semicolon"
	OutputNoTZ      = "output.no_tz"

	SMSAddress         = "sms.address"
	SMSAPIKey          = "sms.api_key"
	SMSIgnoreTLSErrors = "sms.ignore_tls_errors"

	SFTPUsernameLength = "sftp.username_length"
	SFTPPasswordLength = "sftp.password_length"

	MariaDB   = "debug.mariadb"
	Backup    = "debug.backup"
	NoCleanup = "debug.nocleanup"
)

func Configure() {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)

	fs.String(TempDir, "", "Folder for temporary files")
	fs.String(OutputFilename, "", "Output filename")
	fs.Bool(OutputStrict, false, "Generate strict version of report")
	fs.Bool(OutputSemicolon, false, "Use semicolon instead of comma as separator")
	fs.Bool(OutputNoTZ, false, "Do not include timezone in dates")

	fs.String(SMSAddress, "", "Tipping Point SMS address")
	fs.String(SMSAPIKey, "", "Tipping Point SMS API Key")
	fs.Bool(SMSIgnoreTLSErrors, false, "Ignore SMS TLS errors")

	fs.Int(SFTPUsernameLength, DefaultUsernameLength, "sFTP username length")
	fs.Int(SFTPPasswordLength, DefaultPasswordLength, "sFTP password length")

	fs.String(MariaDB, maria.MariaDBZip, "MariaDB ZIP file")
	fs.String(Backup, "", "SMS Backup File")
	fs.Bool(NoCleanup, false, "Keep temporary folder")

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
	mandatory := []string{
		OutputFilename,
	}
	if viper.GetString(Backup) == "" {
		mandatory = append(mandatory, SMSAddress, SMSAPIKey)
	}
	err = prompt.Mandatory(fs, mandatory...)
	if err != nil {
		log.Fatal(err)
	}
	/*
		if viper.GetString(flagOutput) == "" {
			Panic("missing %s", flagOutput)
		}
		if viper.GetString(flagSMSAddress) == "" {
			Panic("missing %s", flagSMSAddress)
		}
		if viper.GetString(flagSMSAPIKey) == "" {
			Panic("missing %s", flagSMSAPIKey)
		}*/
}
