package helper

import (
	"fmt"
	"mime/multipart"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
)

func ValidateFile(fileHeader *multipart.FileHeader, maxBytes int64, allowedTypes []string) error {
	if fileHeader.Size > maxBytes {
		maxMB := maxBytes / (1024 * 1024)
		return model.NewValidationError(fmt.Sprintf("Ukuran file terlalu besar. Maksimal %dMB", maxMB))
	}

	contentType := fileHeader.Header.Get("Content-Type")
	isValidType := false
	
	for _, t := range allowedTypes {
		if t == contentType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		allowedStr := strings.Join(allowedTypes, ", ")
		return model.NewValidationError(fmt.Sprintf("Format file tidak didukung. Tipe yang diizinkan: %s", allowedStr))
	}

	return nil
}