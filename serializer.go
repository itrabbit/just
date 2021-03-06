package just

import (
	"sync"
)

// Input interface serializer.
type ISerializeInput interface {
	Data() interface{}
	Options() interface{}
}

// Serializer interface.
type ISerializer interface {
	Name() string                                    // Name serializer.
	Charset() string                                 // Get Charset serializer.
	DefaultContentType(withCharset bool) string      // The default content type.
	Serialize(interface{}) ([]byte, error)           // Serialize obj/objs to bytes.
	Deserialize([]byte, interface{}) error           // Deserialize bytes tp obj/objs.
	Response(status int, data interface{}) IResponse // Serialize obj/objs to IResponse.
}

// Serializer manager interface to manage the serializers.
type ISerializerManager interface {
	Names() []string
	DefaultName() (string, bool)
	SetDefaultName(string) ISerializerManager
	SetSerializer(string, []string, ISerializer) ISerializerManager
	Serializer(n string, byContent bool) ISerializer
}

type serializerManager struct {
	sync.RWMutex
	nameDefaultSerializer string
	mapByName             map[string]ISerializer
	mapByContentType      map[string]ISerializer
}

func (m *serializerManager) Names() []string {
	names := make([]string, 0)
	for name := range m.mapByName {
		names = append(names, name)
	}
	return names
}

func (m *serializerManager) DefaultName() (string, bool) {
	if len(m.nameDefaultSerializer) < 1 && len(m.mapByName) > 0 {
		for name := range m.mapByName {
			return name, false
		}
	}
	return m.nameDefaultSerializer, true
}

func (m *serializerManager) SetDefaultName(name string) ISerializerManager {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.mapByName[name]; ok || len(name) < 1 {
		m.nameDefaultSerializer = name
	}
	return m
}

func (m *serializerManager) SetSerializer(name string, contentTypes []string, serializer ISerializer) ISerializerManager {
	if name != "default" {
		m.RLock()
		defer m.RUnlock()
		if m.mapByName == nil {
			m.mapByName = make(map[string]ISerializer)
		}
		m.mapByName[name] = serializer
		if m.mapByContentType == nil {
			m.mapByContentType = make(map[string]ISerializer)
		}
		if contentTypes != nil && len(contentTypes) > 0 {
			for _, contentType := range contentTypes {
				m.mapByContentType[contentType] = serializer
			}
		}
	}
	return m
}

func (m *serializerManager) Serializer(n string, byContent bool) ISerializer {
	if byContent {
		if m.mapByContentType != nil {
			if s, ok := m.mapByContentType[n]; ok {
				return s
			}
		}
	} else if m.mapByName != nil {
		if s, ok := m.mapByName[n]; ok {
			return s
		}
	}
	return nil
}
