package smsbackup

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/mpkondrashin/certlist/pkg/model"
)

/*
[ServerName] IPS
[IP] IPS Management Address
[OS] OS Version
[Port] 443
[IssuerName] from the certificate
[ExpirationDate] from the certificate
[EffectiveDate] from the certificate
[KeySize0] key size from the certificate
[SerialNumber] the certificate
[Thumbprint] - HASH from the certificate
[SignatureAlgorithm] algorithm of the signature - from the certificate
[SubjectName] Subject of the certificate
[Version] of the certificate (appers to be 3)
*/

type ReportLine struct {
	CertName           string
	SubjectName        string
	IssuerName         string
	ExpirationDate     string
	EffectiveDate      string
	KeySize0           string
	SerialNumber       string
	Thumbprint         string
	SignatureAlgorithm string
	Version            string
	SSLServerProxies   string
	IpsName            string
	ManagmentIP        string
	Tos                string
}

func GenerateReport(db *sql.DB) (report []ReportLine, err error) {
	for cert, err := range model.RangeNamedCertificate(db, "PRIVATE_KEY_EXPECTED=1") {
		if err != nil {
			return nil, err
		}

		line := ReportLine{
			CertName:   cert.Name,
			Thumbprint: cert.Thumbprint,
		}
		err = GetX509(&line, cert.CertBytes)
		if err != nil {
			return nil, err
		}
		err := GetIPS(db, cert.ID, &line)
		if err != nil {
			return nil, err
		}
		err = GetSSLServer(db, cert.ID, &line)
		if err != nil {
			return nil, err
		}
		report = append(report, line)
	}
	return report, nil
}

var ErrFailedToParsePEMCertificate = fmt.Errorf("failed to parse certificate PEM")

// getKeySize returns the key size in bits from an x509 certificate
func getKeySize(cert *x509.Certificate) int {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return pub.N.BitLen()
	case *ecdsa.PublicKey:
		return pub.Curve.Params().BitSize
	default:
		return -1 // Unknown or unsupported key type
	}
}

func GetX509(line *ReportLine, certData []byte) error {
	block, _ := pem.Decode([]byte(certData))
	if block == nil {
		return ErrFailedToParsePEMCertificate
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	line.IssuerName = cert.Issuer.String()
	line.ExpirationDate = cert.NotAfter.String()
	line.EffectiveDate = cert.NotBefore.String()
	line.KeySize0 = strconv.Itoa(getKeySize(cert))
	line.SerialNumber = cert.SerialNumber.String()
	line.SignatureAlgorithm = cert.SignatureAlgorithm.String()
	line.SubjectName = cert.Subject.String()
	line.Version = strconv.Itoa(cert.Version)

	return nil
}

func GetIPS(db *sql.DB, certID int, line *ReportLine) error {
	where := fmt.Sprintf("NAMED_CERTIFICATE_ID = '%d'", certID)
	for deviceCertificate, err := range model.RangeDeviceCertificate(db, where) {
		if err != nil {
			return err
		}
		err := AddDeviceInfo(db, deviceCertificate.DeviceShortID, line)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddDeviceInfo(db *sql.DB, deviceShortID uint, line *ReportLine) error {
	where := fmt.Sprintf("SHORT_ID = '%d'", deviceShortID)
	for device, err := range model.RangeTptDevice(db, where) {
		if err != nil {
			return err
		}
		line.IpsName = device.DisplayName.String
		line.ManagmentIP = device.IPAddress.String
		_ = device.DeviceMode.String
		line.Tos = device.SoftwareVersion.String
		_ = device.Location.String
	}
	return nil
}

func GetSSLServer(db *sql.DB, certID int, line *ReportLine) error {
	where := fmt.Sprintf("NAMED_CERTIFICATE_ID = %d", certID)
	for sslServerCert, err := range model.RangeSslServerCertificates(db, where) {
		if err != nil {
			return err
		}
		where = fmt.Sprintf("SSL_SERVER_ID = '%s'", sslServerCert.SslServerID)
		for sslServer, err := range model.RangeSslServer(db, where) {
			if err != nil {
				return err
			}
			line.SSLServerProxies += sslServer.Name + ","
		}
	}
	if len(line.SSLServerProxies) > 0 {
		line.SSLServerProxies = line.SSLServerProxies[:len(line.SSLServerProxies)-1] // remove last comma
	}
	return nil
}
