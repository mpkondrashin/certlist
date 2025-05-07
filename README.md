
# CertList — Get List of All TLS Certificates Used For Server SSL Inspection 

<p align="center">
  <img src="media/certlist.jpeg" width="600"/>
</p>

**CertList generates CSV file with all server TLS inspection certificates with info on which IPS boxes they are used**

## Report Format

Resulting CSV file has following columns:
-	IpsName - name of Tipping Point IPS using this certificate
-	ManagmentIP - Tipping Point IPS management IP address
-	Tos - Tipping Point IPS TOS version
-	StartPort - Start Port (usually 443)
-	IssuerName - certificate X.500 issuer
-	ExpirationDate - date after which the certificate is no longer valid
-	EffectiveDate - date before which the certificate is not yet valid    
-	KeySize0 - size of the key (1024, 2048, ...)
-	SerialNumber - serial number of the certificate 
-	Thumbprint - certificate thumbprint
-	SignatureAlgorithm - used cryptographic algorithm (for example SHA-256)
-	SubjectName - certificate X.500 subject
-	Version - used certificate version 
-	SSLServerProxies - name of the SSL server proxies names configured in SMS and using this certificate   
-	CertName - certificate name as it was provided in SMS console

If ```--strict``` protion provided list of the parameters will be the following:
- [ServerName] IPS
- [IP] IPS Management Address
- [OS] OS Version
- [Port] 443
- [IssuerName] from the certificate
- [ExpirationDate] from the certificate
- [EffectiveDate] from the certificate
- [KeySize0] key size from the certificate
- [SerialNumber] the certificate
- [Thumbprint] - HASH from the certificate
- [SignatureAlgorithm] algorithm of the signature - from the certificate
- [SubjectName] Subject of the certificate
- [Version] of the certificate (appers to be 3)

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
certlist.exe --output.filename report.csv
```

**Note:** No need to unpack mariadb-latest.zip archive.zip.

## Configuration

CertList provides following ways to provide options:
1. Configuration file config.yaml. Application seeks for this file in its current folder or folder of CertList executable
2. Environment variables
3. Command line parameters
4. CLI input

Full config.yaml file explained:
```yaml
temp: # use this folder for temporary files instead of %TEMP%
output:
  filename: # output filename
  strict: # true/false - another CSV format
  semicolon: # true/false - use semicolon instead of comma as separator
  no_tz: # true/false - do not include timezone in dates
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

If one of the mandatory parameters of the CertList is missing, it will prompt for the value.

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

### Multiply TPS boxes
If the same certificate is used more than on one TPS box, the CSV
report will contain information on the same certificate for each box.

### Platform support
Only Windows is supported.

### Local firewall
Local firewall may affect ability of Certlist to operate. It is recommended to turn it off.