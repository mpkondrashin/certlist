package smsbackup

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
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
	//Id                 string
	IpsName            string `csv:"[ServerName]"`
	ManagmentIP        string `csv:"[IP]"`
	Tos                string `csv:"[OS]"`
	StartPort          string `csv:"[Port]"`
	IssuerName         string `csv:"[IssuerName]"`
	ExpirationDate     string `csv:"[ExpirationDate]"`
	EffectiveDate      string `csv:"[EffectiveDate]"`
	KeySize0           string `csv:"[KeySize0]"`
	SerialNumber       string `csv:"[SerialNumber]"`
	Thumbprint         string `csv:"[Thumbprint]"`
	SignatureAlgorithm string `csv:"[SignatureAlgorithm]"`
	SubjectName        string `csv:"[SubjectName]"`
	Version            string `csv:"[Version]"`
	// Extra
	SSLServerProxies string
	CertName         string
}

/*
	func NewReportLine(id, name, thumbprint string) *ReportLine {
		return &ReportLine{
			Id:         id,
			CertName:   name,
			Thumbprint: thumbprint,
		}
	}
*/

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
	r.ExpirationDate = cert.NotAfter.Format("2006-01-02 15:04:05")
	r.EffectiveDate = cert.NotBefore.Format("2006-01-02 15:04:05")
	r.KeySize0 = strconv.Itoa(getKeySize(cert))
	r.SerialNumber = cert.SerialNumber.String()
	r.SignatureAlgorithm = cert.SignatureAlgorithm.String()
	r.SubjectName = cert.Subject.String()
	r.Version = strconv.Itoa(cert.Version)
	return nil
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

var query = `
SELECT
	nc.NAME,
	nc.THUMBPRINT,
	nc.CERT_BYTES,
	GROUP_CONCAT(ss.NAME SEPARATOR ',') AS SSL_SERVER_PROXIES,
	COALESCE(td.DISPLAY_NAME,''),
	COALESCE(sslp.START_PORT,''),
	COALESCE(td.IP_ADDRESS,''),
	COALESCE(td.SOFTWARE_VERSION,'')
FROM NAMED_CERTIFICATE nc
LEFT JOIN DEVICE_CERTIFICATE dc ON nc.ID = dc.NAMED_CERTIFICATE_ID
LEFT JOIN TPT_DEVICE td ON dc.DEVICE_SHORT_ID = td.SHORT_ID
LEFT JOIN SSL_SERVER_CERTIFICATES ssc ON nc.ID = ssc.NAMED_CERTIFICATE_ID
LEFT JOIN SSL_SERVER ss ON ssc.SSL_SERVER_ID = ss.SSL_SERVER_ID
LEFT JOIN SSL_SERVER_PORT sslp ON ss.SSL_SERVER_ID = sslp.SSL_SERVER_ID
WHERE nc.PRIVATE_KEY_EXPECTED=1
GROUP BY ss.NAME;`

//GROUP BY nc.ID,nc.NAME,nc.THUMBPRINT,nc.CERT_BYTES;`

// NAMED_CERTIFICATE.ID,
func GenerateReport_(db *sql.DB) (report []ReportLine, err error) {
	/*	query := `
		SELECT
		    nc.NAME as cert_name,
		    nc.THUMBPRINT,
		    nc.CERT_BYTES,
		    GROUP_CONCAT(DISTINCT ss.PROXY_NAME) as ssl_server_proxies,
		    GROUP_CONCAT(DISTINCT td.IPS_NAME) as ips_name,
		    GROUP_CONCAT(DISTINCT td.MANAGMENT_IP) as managment_ip,
		    GROUP_CONCAT(DISTINCT td.TOS) as tos
		FROM NAMED_CERTIFICATE nc
		LEFT JOIN DEVICE_CERTIFICATE dc ON nc.ID = dc.NAMED_CERTIFICATE_ID
		LEFT JOIN TPT_DEVICE td ON dc.DEVICE_SHORT_ID = td.SHORT_ID
		LEFT JOIN SSL_SERVER_CERTIFICATES ssc ON nc.ID = ssc.NAMED_CERTIFICATE_ID
		LEFT JOIN SSL_SERVER ss ON ssc.SSL_SERVER_ID = ss.SSL_SERVER_ID
		WHERE nc.PRIVATE_KEY_EXPECTED = 1
		GROUP BY nc.NAME, nc.THUMBPRINT, nc.CERT_BYTES`*/
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var reportLine ReportLine
		//&reportLine.Id,
		var cn, tp, ccb, ssp, in, sp, mip, tos sql.NullString
		err := rows.Scan(&cn, &tp, &ccb, &ssp, &in, &sp, &mip, &tos)
		if err != nil {
			return nil, err
		}
		if cn.Valid {
			reportLine.CertName = cn.String
		}
		if tp.Valid {
			reportLine.Thumbprint = tp.String
		}
		if ssp.Valid {
			reportLine.SSLServerProxies = ssp.String
		}
		if in.Valid {
			reportLine.IpsName = in.String
		}
		if sp.Valid {
			reportLine.StartPort = sp.String
		}
		if mip.Valid {
			reportLine.ManagmentIP = mip.String
		}
		if tos.Valid {
			reportLine.Tos = tos.String
		}
		if err := reportLine.GetX509([]byte(ccb.String)); err != nil {
			return nil, err
		}
		log.Printf("Certificate name: %s", reportLine.CertName)
		report = append(report, reportLine)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return report, nil
}

/*
func (r *ReportLine) GetIPS(db *sql.DB) error {
	query := "SELECT DEVICE_SHORT_ID FROM DEVICE_CERTIFICATE WHERE NAMED_CERTIFICATE_ID = ?"
	rows, err := db.Query(query, r.Id)
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
	rows, err := db.Query(query, r.Id)
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
		if err := rows.Scan(&sslServerName); err != nil {
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
*/
