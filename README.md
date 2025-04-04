
# CertList — Get List of All TLS Certificates Used For Server SSL Inspection 

**CertList generates CSV file with all server TLS inspection certificates with info on which IPS boxes they are used**

## How to use:
1. Create API Key
2. Write config file
3. Run certlist.exe --output &lt;output filename&gt;

### Create API Key
1. Open Tipping Point SMS
2. Go to Admin -> Authentication and Authorization -> Users
3. Create new user and add it to the superuser group
4. Save API Key

### Write config file
Create following minimal configuration file:
```yaml
sms:
  address: 1.2.3.4
  api_key: A95BE8AB-AE11-45C5-B813-A9A2FDC27E5B
  ignore_tls_errors: true
```

### Run CertList

Download certlist executable  [releases](https://github.com/mpkondrashin/certlist/releases/latest)

Run following command:
```commandline
certlist.exe
```

## Options

CertList provides following ways to provide options:
1. Configuration file config.yaml. Application seeks for this file in its current folder or folder of CertList executable
2. Environment variables
3. Command line parameters

Full config file explained:
```yaml
temp: # use this folder for temporary files instead of %TEMP%
output: # output filename
sms:
  address: # IP address or DNS name
  api_key: # SMS API Key
  ignore_tls_errors: false # Can be set to true if SMS does not have correct certificate
sftp:
  username_length: # sftp username length
  password_length: # sftp password length
debug:
  mariadb: # MariaDB portable ZIP file to use instead of mariadb-latest.zip
  backup: # SMS backup file to use instead of downloading it from SMS
  nocleanup: # Do not remove temporary files
```

To set these parameters through command line, use following notation: <section>.<parameter>. For example to ignore TLS errors use following command line option:
```commandline 
certalert --sms.ignore_tls_errors
```

To set these parameters through environment variable, add CERTLIST_ prefix and put "_" (underscore) between section and option. Example for API Key:
```commandline
CERTALERT_SMS_API_KEY=A95BE8AB-AE11-45C5-B813-A9A2FDC27E5B
```

## System Requirements

- OS: Windows
- CPU: x86_64
- HDD: 2GB of free space
- Network:
  - Opened port 443 to the direction of the SMS
  - Opened port 22 to the direction of the CertList

## Known Issues

### On Windows - %TEMP% and certlist.exe on same drive!
On Windows, system TEMP folder should be on same drive and certlist.exe program (actually as current folder). If it is not so, "temp" parameter of configuration can be used, e.g. "temp: D:\TEMP" in config.yaml or TMP environment variable can be set.

### Bidirectional connectivity
Bidirectional connectivity must be provided from host running CertList to SMS and back (see System Requirements).

### IPv6
If using IPv6 address for SMS, please put it in square brackets.

