package just

import "sync"

type ISerializer interface {
	DefaultContentType() string
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type ISerializerManager interface {
	SetSerializer(name string, contentTypes []string, serializer ISerializer) ISerializerManager
	GetSerializerByName(string) ISerializer
	GetSerializerByContentType(contentType string) ISerializer
}

type serializerManager struct {
	sync.RWMutex
	serializersByName        map[string]ISerializer
	serializersByContentType map[string]ISerializer
}

func (m *serializerManager) SetSerializer(name string, contentTypes []string, serializer ISerializer) ISerializerManager {
	m.RLock()
	defer m.RUnlock()
	if m.serializersByName == nil {
		m.serializersByName = make(map[string]ISerializer)
	}
	m.serializersByName[name] = serializer
	if m.serializersByContentType == nil {
		m.serializersByContentType = make(map[string]ISerializer)
	}
	if contentTypes != nil && len(contentTypes) > 0 {
		for _, contentType := range contentTypes {
			m.serializersByContentType[contentType] = serializer
		}
	}
	return m
}

func (m *serializerManager) GetSerializerByName(name string) ISerializer {
	m.Lock()
	defer m.Unlock()
	if m.serializersByName != nil {
		if s, ok := m.serializersByName[name]; ok {
			return s
		}
	}
	return nil
}

func (m *serializerManager) GetSerializerByContentType(contentType string) ISerializer {
	m.Lock()
	defer m.Unlock()
	if m.serializersByContentType != nil {
		if s, ok := m.serializersByContentType[contentType]; ok {
			return s
		}
	}
	return nil
}
