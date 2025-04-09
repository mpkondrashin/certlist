
# CertList — Get List of All TLS Certificates Used For Server SSL Inspection 

**CertList generates CSV file with all server TLS inspection certificates with info on which IPS boxes they are used**

## Report Format

Resulting CSV file has following columns:
-	CertName - certificate name as it was provided in SMS console
-	SubjectName - certificate X.500 subject
-	IssuerName - certificate X.500 issuer
-	ExpirationDate - date after which the certificate is no longer valid
-	EffectiveDate - date before which the certificate is not yet valid    
-	KeySize0 - size of the key (1024, 2048, ...)
-	SerialNumber - serial number of the certificate 
-	Thumbprint - certificate thumbprint
-	SignatureAlgorithm - used cryptographic algorithm (for example SHA-256)
-	Version - used certificate version 
-	SSLServerProxies - name of the SSL server proxies names configured in SMS and using this certificate   
-	IpsName - name of Tipping Point IPS using this certificate
-	ManagmentIP - Tipping Point IPS management IP address
-	Tos - Tipping Point IPS TOS version

**Note:** if same certificate used in mode than one Server SSL Profile/IPS, then they all will be listed in CSV file separated by comma.

## How to use:
1. Create API Key
2. Edit config file
3. Run certlist.exe with appropriate options

### Create API Key
1. Open Tipping Point SMS
2. Go to Admin -> Authentication and Authorization -> Users
3. Create new user and add it to the superuser group, or create special role (see below).
4. Save API Key

#### Minimal permissions

Role Capabilities:
- Admin -> Admin Section -> SMS Management -> Access Management -> Access SMS Web Services
- Admin -> Administer the SMS -> Database Management -> Manage Database -> Backup Database

### Write config file
Create following minimal configuration file:
```yaml
sms:
  address: 1.2.3.4 <- replace with SMS server address
  api_key: A95BE8AB-AE11-45C5-B813-A9A2FDC27E5B <- replace with generated API Key
  ignore_tls_errors: false <- change to true if your SMS does not have valid certificate
```

### Run CertList

1. Download CertList distributive from  [releases](https://github.com/mpkondrashin/certlist/releases/latest).
2. Unpack ZIP archive
3. Run following command:
```commandline
certlist.exe --output report.csv
```

**Note:** No need to unpack mariadb-latest.zip archive.zip.

## Configuration

CertList provides following ways to provide options:
1. Configuration file config.yaml. Application seeks for this file in its current folder or folder of CertList executable
2. Environment variables
3. Command line parameters

Full config.yaml file explained:
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
On Windows, system TEMP folder should be on same drive and certlist.exe program (actually as current folder). If it is not so, "temp" parameter of configuration can be used, e.g. "temp: D:\TEMP" in config.yaml or TEMP/TMP environment variable can be set.

### Bidirectional connectivity
Bidirectional connectivity must be provided from host running CertList to SMS and back (see System Requirements).

### IPv6
If using IPv6 address for SMS, please put it in square brackets.

### Client SSL Certificates
The script does not provide any info on usage of the Client SSL Inspection certificates, though the certificates themselves will be listed.

### Running time
Although the task seem to be trivial, certlist can run over 10 minutes.