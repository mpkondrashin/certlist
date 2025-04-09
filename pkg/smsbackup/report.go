package smsbackup

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"strconv"
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
	id                 string
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

func NewReportLine(id, name, thumbprint string) *ReportLine {
	return &ReportLine{
		id:         id,
		CertName:   name,
		Thumbprint: thumbprint,
	}
}

func (r *ReportLine) GetX509(certData []byte) error {
	block, _ := pem.Decode(certData)
	if block == nil {
		return ErrFailedToParsePEMCertificate
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	r.IssuerName = cert.Issuer.String()
	r.ExpirationDate = cert.NotAfter.String()
	r.EffectiveDate = cert.NotBefore.String()
	r.KeySize0 = strconv.Itoa(getKeySize(cert))
	r.SerialNumber = cert.SerialNumber.String()
	r.SignatureAlgorithm = cert.SignatureAlgorithm.String()
	r.SubjectName = cert.Subject.String()
	r.Version = strconv.Itoa(cert.Version)

	return nil
}

func (r *ReportLine) GetIPS(db *sql.DB) error {
	query := "SELECT DEVICE_SHORT_ID FROM DEVICE_CERTIFICATE WHERE NAMED_CERTIFICATE_ID = ?"
	rows, err := db.Query(query, r.id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var DeviceShortID string
		if err := rows.Scan(&DeviceShortID); err != nil {
			return err
		}

		if err := r.AddDeviceInfo(db, DeviceShortID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *ReportLine) AddDeviceInfo(db *sql.DB, deviceShortID string) error {
	query := "SELECT DISPLAY_NAME,SOFTWARE_VERSION,IP_ADDRESS FROM TPT_DEVICE WHERE SHORT_ID = ?"
	rows, err := db.Query(query, deviceShortID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		return rows.Scan(&r.IpsName, &r.Tos, &r.ManagmentIP)
	}
	return rows.Err()
	//_ = device.DeviceMode.String
	//_ = device.Location.String
}

func (r *ReportLine) GetSSLServer(db *sql.DB) error {
	query := "SELECT SSL_SERVER_ID FROM SSL_SERVER_CERTIFICATES WHERE NAMED_CERTIFICATE_ID = ?"
	rows, err := db.Query(query, r.id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sslServerID string
		err := rows.Scan(&sslServerID)
		if err != nil {
			return err
		}
		if err := r.GetSSLServerForID(db, sslServerID); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(r.SSLServerProxies) > 0 {
		r.SSLServerProxies = r.SSLServerProxies[:len(r.SSLServerProxies)-1] // remove last comma
	}

	return nil
}

func (r *ReportLine) GetSSLServerForID(db *sql.DB, sslServerID string) error {
	query := "SELECT NAME FROM SSL_SERVER WHERE SSL_SERVER_ID = ?"
	rows, err := db.Query(query, sslServerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sslServerName string
		err := rows.Scan(&sslServerName)
		if err != nil {
			return err
		}
		r.SSLServerProxies += sslServerName + ","
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func GenerateReport(db *sql.DB) (report []ReportLine, err error) {
	query := "SELECT ID,NAME,THUMBPRINT,CERT_BYTES FROM NAMED_CERTIFICATE WHERE PRIVATE_KEY_EXPECTED=1"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var certID, certName, certThumbprint, certCertBytes string
		err := rows.Scan(&certID, &certName, &certThumbprint, &certCertBytes)
		if err != nil {
			return nil, err
		}
		reportLine := NewReportLine(certID, certName, certThumbprint)

		if err := reportLine.GetX509([]byte(certCertBytes)); err != nil {
			return nil, err
		}

		if err := reportLine.GetIPS(db); err != nil {
			return nil, err
		}

		if err := reportLine.GetSSLServer(db); err != nil {
			return nil, err
		}
		report = append(report, *reportLine)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
