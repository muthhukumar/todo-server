package backup

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"time"
	"todo-server/models"
)

func BackupTasks(DB *sql.DB, emailAuth models.EmailAuth) {
	csvData, err := generateCSVData(DB)

	if err != nil {
		log.Println("Failed to generate csv data", err)
		return
	}

	encodedCsv := base64.StdEncoding.EncodeToString(csvData.Bytes())

	if success := sendEmail(emailAuth, "Attached is the backup CSV file you requested.", encodedCsv); !success {
		log.Println("Sending backup email failed")

		return
	}

	log.Println("Backup successful")
}

func sendEmail(emailAuth models.EmailAuth, body string, encodedCsv string) bool {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	addr := fmt.Sprintf("%v:%v", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", emailAuth.FromEmail, emailAuth.Password, smtpHost)

	now := time.Now()
	subject := fmt.Sprintf("Backup file generated on %s", now.Format("Monday, January 2, 2006 3:04 PM"))

	emailContent := createEmailWithAttachment(emailAuth.FromEmail, emailAuth.ToEmail, subject, body, "backup.csv", encodedCsv)

	err := smtp.SendMail(addr, auth, emailAuth.FromEmail, []string{emailAuth.ToEmail}, emailContent.Bytes())

	if err != nil {
		log.Println("Failed to send backup email", err.Error())
		return false
	}

	return true
}

func generateCSVData(DB *sql.DB) (*bytes.Buffer, error) {
	query := "select * from tasks"

	rows, err := DB.Query(query)
	defer rows.Close()

	if err != nil {
		log.Println("Failed to execute query to backup tasks")
		return nil, err
	}

	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)
	defer writer.Flush()

	columns, err := rows.Columns()

	if err != nil {
		log.Println("Getting all the rows failed", err.Error())
		return nil, err
	}

	if err := writer.Write(columns); err != nil {
		log.Println("Writing columns to write failed", err.Error())
		return nil, err
	}

	for rows.Next() {
		columns := make([]interface{}, len(columns))
		columnsPoints := make([]interface{}, len(columns))

		for i := range columns {
			columnsPoints[i] = &columns[i]
		}

		if err := rows.Scan(columnsPoints...); err != nil {
			log.Println("Failed to scan rows", err.Error())
			return nil, err
		}

		record := make([]string, len(columns))

		for i, col := range columns {
			if col != nil {
				record[i] = fmt.Sprintf("%v", col)
			}
		}

		if err := writer.Write(record); err != nil {
			log.Println("Writing record to csv writer failed", err.Error())
			return nil, err
		}
	}

	return &csvBuffer, nil
}

func createEmailWithAttachment(from, to, subject, body, filename, encodedAttachment string) *bytes.Buffer {
	var emailBuffer bytes.Buffer
	boundary := "boundary12345"
	writer := multipart.NewWriter(&emailBuffer)

	emailBuffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	emailBuffer.WriteString(fmt.Sprintf("To: %s\r\n", to))
	emailBuffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	emailBuffer.WriteString(fmt.Sprintf("MIME-Version: 1.0\r\n"))
	emailBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	emailBuffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(body)
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString(fmt.Sprintf("Content-Type: text/csv; name=\"%s\"\r\n", filename))
	emailBuffer.WriteString("Content-Transfer-Encoding: base64\r\n")
	emailBuffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(encodedAttachment)
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	writer.Close()
	return &emailBuffer
}
