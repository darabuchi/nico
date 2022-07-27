package adapter

import (
	"bytes"
	"reflect"
	"strings"
	"time"
)

var (
	skipUniqueKeyMap map[string]bool
	extraInfoTypeMap map[string]reflect.Kind
)

func init() {
	t := reflect.TypeOf(ExtraInfo{})
	skipUniqueKeyMap = make(map[string]bool, t.NumField()+2)
	skipUniqueKeyMap["name"] = true
	skipUniqueKeyMap["unique_id"] = true
	
	extraInfoTypeMap = make(map[string]reflect.Kind, t.NumField())
	
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" {
			continue
		}
		
		for _, item := range strings.Split(tag, ",") {
			if item == "omitempty" {
				continue
			}
			skipUniqueKeyMap[item] = true
			t.Kind()
		}
		
	}
}

type ExtraInfo struct {
	Country      string `json:"country,omitempty"`
	Region       string `json:"region,omitempty"`
	City         string `json:"city,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	CountryEmoji string `json:"country_emoji,omitempty"`
	
	Delay       time.Duration `json:"delay,omitempty"`
	GoogleDelay time.Duration `json:"google,omitempty"`
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
