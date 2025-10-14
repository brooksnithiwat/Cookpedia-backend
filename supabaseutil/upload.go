package supabaseutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// sanitizeFilename ลบอักขระที่ไม่ใช่ [a-zA-Z0-9._-]
var thaiToEng = map[rune]string{
	'ก': "k", 'ข': "kh", 'ค': "kh", 'ง': "ng",
	'จ': "ch", 'ฉ': "ch", 'ช': "ch", 'ซ': "s",
	'ด': "d", 'ต': "t", 'ถ': "th", 'ท': "th", 'น': "n",
	'บ': "b", 'ป': "p", 'ผ': "ph", 'พ': "ph", 'ม': "m",
	'ย': "y", 'ร': "r", 'ล': "l", 'ว': "w",
	'ส': "s", 'ศ': "s", 'ษ': "s", 'อ': "o", 'ฮ': "h",
	'า': "a", 'ิ': "i", 'ี': "i", 'ุ': "u", 'ู': "u",
	'เ': "e", 'แ': "ae", 'โ': "o", 'ใ': "ai", 'ไ': "ai",
	'ำ': "am", '่': "", '้': "", '๊': "", '๋': "",
	'์': "", ' ': "_",
}

func sanitizeFilename(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	var sb strings.Builder

	for _, r := range base {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			sb.WriteRune(r)
		} else if val, ok := thaiToEng[r]; ok {
			sb.WriteString(val)
		} else {
			sb.WriteRune('_') // กรณีอักขระอื่น ๆ
		}
	}

	return sb.String() + ext
}

func UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID interface{}, pathPrefix string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	bucket := os.Getenv("SUPABASE_BUCKET_NAME")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	// sanitize ชื่อไฟล์
	cleanName := sanitizeFilename(fileHeader.Filename)
	filename := fmt.Sprintf("%s_%v_%s", pathPrefix, userID, cleanName)
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, filename)

	// อ่านไฟล์ทั้งหมดลง buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}

	// ใช้ PUT แทน POST และ set header upsert
	req, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", anonKey))
	req.Header.Set("Content-Type", fileHeader.Header.Get("Content-Type"))
	req.Header.Set("x-upsert", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		return "", fmt.Errorf("upload failed: %v", body)
	}

	// คืน public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, filename)
	return publicURL, nil
}
