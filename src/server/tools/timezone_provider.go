package tools

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	u "github.com/quollix/common/utils"
)

type TimezoneProvider interface {
	InitializeIanaTimezones() error
	ListIanaTimezones() []string
	IsIanaTimezoneValid(timezone string) bool
}

type TimezoneProviderImpl struct {
	ianaTimezonesMap map[string]any
	ianaTimezones    []string
}

func (p *TimezoneProviderImpl) InitializeIanaTimezones() error {
	u.Logger.Info("Initializing Iana Timezones")
	var zones []string
	root := "/usr/share/zoneinfo"
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return u.Logger.NewError(err.Error())
		}
		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return u.Logger.NewError(err.Error())
		}

		if strings.Contains(rel, "posix") ||
			strings.Contains(rel, "right") ||
			strings.HasPrefix(rel, "Etc/") ||
			strings.HasPrefix(rel, "SystemV") {
			return nil
		}

		zones = append(zones, filepath.ToSlash(rel))
		return nil
	})

	if err != nil {
		return err
	}
	sort.Strings(zones)
	p.ianaTimezones = zones
	p.ianaTimezonesMap = make(map[string]any)
	for _, tz := range zones {
		p.ianaTimezonesMap[tz] = nil
	}
	return nil
}

func (p *TimezoneProviderImpl) ListIanaTimezones() []string {
	return p.ianaTimezones
}

func (p *TimezoneProviderImpl) IsIanaTimezoneValid(timezone string) bool {
	_, ok := p.ianaTimezonesMap[timezone]
	return ok
}
