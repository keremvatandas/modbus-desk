package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Profile struct {
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	UnitID      int    `json:"unitId"`
	TimeoutMs   int    `json:"timeoutMs"`
	Retries     int    `json:"retries"`
	AddressBase int    `json:"addressBase"`
	DataType    string `json:"dataType"`
	ByteOrder   string `json:"byteOrder"`
}

type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ModbusDesk", "profiles.json"), nil
}

func (s *Store) List() ([]Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	sort.Slice(profiles, func(i, j int) bool {
		return strings.ToLower(profiles[i].Name) < strings.ToLower(profiles[j].Name)
	})
	return profiles, nil
}

func (s *Store) Save(profile Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile.Name = strings.TrimSpace(profile.Name)
	profile.Host = strings.TrimSpace(profile.Host)
	if profile.Name == "" {
		return errors.New("profile name is required")
	}
	if profile.Host == "" {
		return errors.New("profile host is required")
	}
	if profile.Version == 0 {
		profile.Version = 1
	}
	if profile.Port == 0 {
		profile.Port = 502
	}
	if profile.TimeoutMs == 0 {
		profile.TimeoutMs = 3000
	}
	if profile.DataType == "" {
		profile.DataType = "uint16"
	}
	if profile.ByteOrder == "" {
		profile.ByteOrder = "ABCD"
	}

	profiles, err := s.readLocked()
	if err != nil {
		return err
	}
	replaced := false
	for i := range profiles {
		if strings.EqualFold(profiles[i].Name, profile.Name) {
			profiles[i] = profile
			replaced = true
			break
		}
	}
	if !replaced {
		profiles = append(profiles, profile)
	}
	return s.writeLocked(profiles)
}

func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("profile name is required")
	}

	profiles, err := s.readLocked()
	if err != nil {
		return err
	}
	filtered := profiles[:0]
	for _, profile := range profiles {
		if !strings.EqualFold(profile.Name, name) {
			filtered = append(filtered, profile)
		}
	}
	return s.writeLocked(filtered)
}

func (s *Store) readLocked() ([]Profile, error) {
	if s.path == "" {
		return nil, errors.New("profile store path is empty")
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Profile{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []Profile{}, nil
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("read profiles: %w", err)
	}
	return profiles, nil
}

func (s *Store) writeLocked(profiles []Profile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o600)
}
