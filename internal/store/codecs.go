package store

import (
	"encoding/json"
	"fmt"
)

func marshal(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode entity: %w", err)
	}
	return encoded, nil
}

func unmarshal(data []byte, target any) error {
	if err := json.Unmarshal(cloneBytes(data), target); err != nil {
		return fmt.Errorf("decode entity: %w", err)
	}
	return nil
}
