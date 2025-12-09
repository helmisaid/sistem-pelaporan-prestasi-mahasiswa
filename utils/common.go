package utils

import "github.com/google/uuid"

func GetStringOrDefault(ptr *string, defaultVal string) string {
    if ptr != nil {
        return *ptr
    }
    return defaultVal
}

func GetBoolOrDefault(ptr *bool, defaultVal bool) bool {
    if ptr != nil {
        return *ptr
    }
    return defaultVal
}

func ParseUUIDSafe(s string) uuid.UUID {
    parsed, err := uuid.Parse(s)
    if err != nil {
        return uuid.Nil
    }
    return parsed
}