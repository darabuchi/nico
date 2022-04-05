package adapter

import (
	"bytes"
)

type ExtraInfo struct {
	Country      string `json:"country,omitempty"`
	Region       string `json:"region,omitempty"`
	City         string `json:"city,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	CountryEmoji string `json:"country_emoji,omitempty"`
}

func ParseClash4Extra(m map[string]any) ExtraInfo {
	p := ExtraInfo{}

	if val, ok := m["country"]; ok {
		p.Country, _ = val.(string)
	}

	if val, ok := m["region"]; ok {
		p.Region, _ = val.(string)
	}

	if val, ok := m["city"]; ok {
		p.City, _ = val.(string)
	}

	if val, ok := m["country_code"]; ok {
		p.CountryCode, _ = val.(string)
	}

	if val, ok := m["country_emoji"]; ok {
		p.CountryEmoji, _ = val.(string)
	}

	return p
}

// CountryEmoji Country-Region-City Name
func (p ExtraInfo) GenNameTpl() string {
	var b bytes.Buffer
	if p.CountryEmoji != "" {
		b.WriteString(p.CountryEmoji)
	}

	return b.String()
}
