package adapter

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/darabuchi/log"
)

type Cache interface {
	Reset()

	Store(key string, val any)

	Load(key string) (val any, err error)
	LoadBool(key string) bool
	LoadUint16(key string) uint16
	LoadFloat64(key string) float64

	Del(key string)
}

var (
	ErrCacheNotFound = errors.New("cache not found")
)

type AdapterCache struct {
	lock sync.RWMutex

	m map[string]any
}

func NewAdapterCache() Cache {
	p := &AdapterCache{
		m: map[string]any{},
	}

	return p
}

func (p *AdapterCache) Store(key string, val any) {
	p.set(key, val)
}

func (p *AdapterCache) set(key string, val any) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.m[key] = val
}

func (p *AdapterCache) Load(key string) (any, error) {
	return p.get(key)
}

func (p *AdapterCache) LoadBool(key string) bool {
	val, err := p.get(key)
	if err != nil {
		log.Debugf("err:%v", err)
		return false
	}

	switch x := val.(type) {
	case bool:
		return x
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return x != 0
	case string:
		switch strings.ToLower(x) {
		case "true", "1":
			return true
		case "false", "0":
			return false
		case "":
			return false
		default:
			return true
		}
	case []byte:
		switch string(bytes.ToLower(x)) {
		case "true", "1":
			return true
		case "false", "0":
			return false
		case "":
			return false
		default:
			return true
		}
	default:
		return val == nil
	}
}

func (p *AdapterCache) LoadUint16(key string) uint16 {
	val, err := p.get(key)
	if err != nil {
		log.Debugf("err:%v", err)
		return 0
	}

	switch x := val.(type) {
	case bool:
		if x {
			return 1
		}
		return 0
	case int:
		return uint16(x)
	case int8:
		return uint16(x)
	case int16:
		return uint16(x)
	case int32:
		return uint16(x)
	case int64:
		return uint16(x)
	case uint:
		return uint16(x)
	case uint8:
		return uint16(x)
	case uint16:
		return x
	case uint32:
		return uint16(x)
	case uint64:
		return uint16(x)
	case float32:
		return uint16(x)
	case float64:
		return uint16(x)
	case string:
		val, err := strconv.ParseUint(x, 10, 16)
		if err != nil {
			log.Debugf("err:%v", err)
			return 0
		}
		return uint16(val)
	case []byte:
		val, err := strconv.ParseUint(string(x), 10, 16)
		if err != nil {
			log.Debugf("err:%v", err)
			return 0
		}
		return uint16(val)
	default:
		log.Panic("know how to handle")
		return 0
	}
}

func (p *AdapterCache) LoadFloat64(key string) float64 {
	val, err := p.get(key)
	if err != nil {
		log.Debugf("err:%v", err)
		return 0
	}

	switch x := val.(type) {
	case bool:
		if x {
			return 1
		}
		return 0
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case string:
		val, err := strconv.ParseUint(x, 10, 16)
		if err != nil {
			log.Debugf("err:%v", err)
			return 0
		}
		return float64(val)
	case []byte:
		val, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			log.Debugf("err:%v", err)
			return 0
		}
		return val
	default:
		log.Panic("know how to handle")
		return 0
	}
}

func (p *AdapterCache) get(key string) (any, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if val, ok := p.m[key]; ok {
		return val, nil
	}

	return nil, ErrCacheNotFound
}

func (p *AdapterCache) Reset() {
	p.lock.Lock()
	defer p.lock.RUnlock()

	p.m = map[string]any{}
}

func (p *AdapterCache) Del(key string) {
	p.del(key)
}

func (p *AdapterCache) del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.m, key)
}
