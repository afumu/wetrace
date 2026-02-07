package security

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

// EncryptTime encrypts a timestamp using the machineID as a salt.
// Format: Base64( XOR( timestamp_string, machineID_part ) )
func EncryptTime(ts int64, machineID string) string {
	if len(machineID) < 8 {
		machineID = "default_salt_if_machine_id_is_too_short"
	}
	salt := machineID[:8]
	tsStr := fmt.Sprintf("%d", ts)

	encrypted := xorString(tsStr, salt)
	return base64.StdEncoding.EncodeToString([]byte(encrypted))
}

// DecryptTime decrypts the obfuscated time string.
func DecryptTime(encryptedStr string, machineID string) (int64, error) {
	if len(machineID) < 8 {
		machineID = "default_salt_if_machine_id_is_too_short"
	}
	salt := machineID[:8]

	decodedBytes, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return 0, err
	}

	decryptedStr := xorString(string(decodedBytes), salt)

	ts, err := strconv.ParseInt(decryptedStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid time format after decryption")
	}

	return ts, nil
}

func xorString(input, key string) string {
	output := make([]byte, len(input))
	keyLen := len(key)
	for i := 0; i < len(input); i++ {
		output[i] = input[i] ^ key[i%keyLen]
	}
	return string(output)
}
