package utils

import (
    "crypto/rand"
    "fmt"
    "github.com/google/uuid"
)

func GenerateUUID() uuid.UUID {
    return uuid.New()
}

func RandomBytes(length int) ([]byte, error) {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        return nil, err
    }
    return bytes, nil
}

func UUIDFromString(str string) (uuid.UUID, error) {
    return uuid.Parse(str)
}

func UUIDToString(uuid uuid.UUID) string {
    return uuid.String()
}

func FormatUUID(uuid uuid.UUID) string {
    return fmt.Sprintf("%s", uuid)
}