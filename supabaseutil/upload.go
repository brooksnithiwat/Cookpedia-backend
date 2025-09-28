package supabaseutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// UploadFile uploads a file to Supabase Storage and returns the public URL
func UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID interface{}) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	bucket := os.Getenv("SUPABASE_BUCKET_NAME")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	filename := fmt.Sprintf("ProfileImage/profile_%v%s", userID, fileHeader.Filename)
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, filename)

	// Read file into buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, buf)
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

	// Public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, filename)
	return publicURL, nil
}
