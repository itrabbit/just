package just

import (
	"sync"
)

type ISerializer interface {
	DefaultContentType() string
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type ISerializerManager interface {
	NameDefaultSerializer() (string, bool)
	SetNameDefaultSerializer(string) ISerializerManager
	SetSerializer(string, []string, ISerializer) ISerializerManager
	SerializerNames() []string
	SerializerByName(string) ISerializer
	SerializerByContentType(string) ISerializer
}

// serializerManager менеджер сериализаторов
type serializerManager struct {
	sync.RWMutex
	nameDefaultSerializer    string
	serializersByName        map[string]ISerializer
	serializersByContentType map[string]ISerializer
}

func (m *serializerManager) SerializerNames() []string {
	names := make([]string, 0)
	for name, _ := range m.serializersByName {
		names = append(names, name)
	}
	return names
}

func (m *serializerManager) NameDefaultSerializer() (string, bool) {
	if len(m.nameDefaultSerializer) < 1 && len(m.serializersByName) > 0 {
		for name, _ := range m.serializersByName {
			return name, false
		}
	}
	return m.nameDefaultSerializer, true
}

func (m *serializerManager) SetNameDefaultSerializer(name string) ISerializerManager {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.serializersByName[name]; ok || len(name) < 1 {
		m.nameDefaultSerializer = name
	}
	return m
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

func (m *serializerManager) SerializerByName(name string) ISerializer {
	if m.serializersByName != nil {
		if s, ok := m.serializersByName[name]; ok {
			return s
		}
	}
	return nil
}

func (m *serializerManager) SerializerByContentType(contentType string) ISerializer {
	if m.serializersByContentType != nil {
		if s, ok := m.serializersByContentType[contentType]; ok {
			return s
		}
	}
	return nil
}
