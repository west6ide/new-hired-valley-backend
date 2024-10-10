package courses

import (
	"encoding/json"
	"net/http"
)

type VimeoUploadResponse struct {
	Uri string `json:"uri"`
}

func UploadToVimeo(videoFilePath string) (string, error) {
	request, err := http.NewRequest("POST", "https://api.vimeo.com/me/videos", nil)
	if err != nil {
		return "", err
	}

	// Токен доступа
	request.Header.Set("Authorization", "bearer 807fd2798cdabc7a1d385c72eb2e5137")
	request.Header.Set("Content-Type", "multipart/form-data")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var uploadResponse VimeoUploadResponse
	if err := json.NewDecoder(response.Body).Decode(&uploadResponse); err != nil {
		return "", err
	}

	return uploadResponse.Uri, nil
}
