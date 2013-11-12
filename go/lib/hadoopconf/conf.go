package hadoopconf

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type Conf interface {
	Keys() []string
	Set(key, val string) (oldval string)
	Get(key string) string
}

type Source struct {
	Source     string
	SourceType SourceType
}

type ConfSourcer interface {
	Conf
	SourceGet(key string) (value string, src Source)
	Source() string
}

type ConfWithDefault struct {
	Conf    ConfSourcer
	Default ConfSourcer
}

func sourceGet(cs ConfSourcer, key string) (string, Source) {
	if cs == nil {
		return "", NoSource
	}
	return cs.SourceGet(key)
}

func (cwd *ConfWithDefault) Keys() []string {
	m := make(map[string]bool)
	for _, c := range []ConfSourcer{cwd.Conf, cwd.Default} {
		if c == nil {
			continue
		}
		for _, key := range c.Keys() {
			m[key] = true
		}
	}
	result := []string{}
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}

func (cwd *ConfWithDefault) Source() string {
	return cwd.Conf.Source() + " default: " + cwd.Default.Source()
}

func (cwd *ConfWithDefault) SourceGet(key string) (value string, src Source) {
	if cwd == nil {
		return "", NoSource
	}
	if v, src := sourceGet(cwd.Conf, key); v != "" {
		return v, src
	}
	return sourceGet(cwd.Default, key)
}

func (cwd *ConfWithDefault) Get(key string) (value string) {
	v, _ := cwd.SourceGet(key)
	return v
}

func (cwd *ConfWithDefault) Set(key, value string) (oldval string) {
	oldval = cwd.Get(key)
	cwd.Conf.Set(key, value)
	return oldval
}

type Property struct {
	Name        string `xml:"name"`
	Value       string `xml:"value"`
	Description string `xml:"description"`
}

type Configuration struct {
	XMLName  xml.Name    `xml:"configuration"`
	Property []*Property `xml:"property"`
}

type FileConfiguration struct {
	*Configuration
	Path     string
	modified bool
}

func NewFileConfiguration(path string) (*FileConfiguration, error) {
	if _, err := os.Open(path); os.IsNotExist(err) {
		return &FileConfiguration{&Configuration{}, path, false}, nil
	}
	conf, err := NewConfigurationFromFile(path)
	if err != nil {
		return nil, err
	}
	return &FileConfiguration{conf, path, false}, nil
}

func (fc *FileConfiguration) Keys() []string {
	keys := []string{}
	for _, p := range fc.Property {
		keys = append(keys, p.Name)
	}
	return keys
}

func (fc *FileConfiguration) Set(key, val string) (oldval string) {
	fc.modified = true
	return fc.Configuration.Set(key, val)
}

func (fc *FileConfiguration) Source() string {
	return fc.Path
}

func (fc *FileConfiguration) SourceGet(key string) (value string, source Source) {
	if p := fc.get(key); p != nil {
		return p.Value, Source{fc.Path, LocalFile}
	}
	return "", NoSource
}

func (fc *FileConfiguration) Save() error {
	if !fc.modified {
		return nil
	}
	if err := ioutil.WriteFile(fc.Path, fc.Bytes(), 0655); err != nil {
		return err
	}
	fc.modified = false
	return nil
}

type GeneratedConf struct {
	*Configuration
	ConfSource Source
}

func NewGeneratedConfFromBytes(source Source, b []byte) (*GeneratedConf, error) {
	conf, err := NewConfigurationFromByte(b)
	if err != nil {
		return nil, err
	}
	return NewGeneratedConf(source, conf), nil
}

func NewGeneratedConfFromString(source Source, s string) (*GeneratedConf, error) {
	return NewGeneratedConfFromBytes(source, []byte(s))
}

func NewGeneratedConf(source Source, conf *Configuration) *GeneratedConf {
	return &GeneratedConf{conf, source}
}

func (gc *GeneratedConf) Keys() []string {
	keys := []string{}
	for _, p := range gc.Property {
		keys = append(keys, p.Name)
	}
	return keys
}

func (gc *GeneratedConf) Source() string {
	return gc.ConfSource.Source
}

func (gc *GeneratedConf) SourceGet(key string) (value string, source Source) {
	p := gc.get(key)
	if p != nil {
		return p.Value, gc.ConfSource
	}
	return "", NoSource
}

type SourceType int

const (
	LocalFile SourceType = iota
	FileFromJar
	Generated
)

type multiSourceConf []ConfSourcer

var NoSource = Source{}

func (msc multiSourceConf) Keys() []string {
	m := make(map[string]bool)
	for _, c := range msc {
		for _, key := range c.Keys() {
			m[key] = true
		}
	}
	result := []string{}
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}

func (msc multiSourceConf) SourceGet(key string) (string, Source) {
	for _, s := range msc {
		v, src := s.SourceGet(key)
		if src != NoSource {
			return v, src
		}
	}
	return "", NoSource
}

func (msc multiSourceConf) SetIfExist(key, value string) (oldval string, src ConfSourcer) {
	for _, s := range msc {
		v := s.Get(key)
		if v != "" {
			return s.Set(key, value), src
		}
	}
	return "", nil
}

func (c *Configuration) get(key string) *Property {
	if c == nil {
		return nil
	}
	for _, prop := range c.Property {
		if prop.Name == key {
			return prop
		}
	}
	return nil
}

func (c *Configuration) getOrAdd(key string) *Property {
	v := c.get(key)
	if v == nil {
		v = &Property{key, "", ""}
		c.Property = append(c.Property, v)
	}
	return v
}

func (c *Configuration) Set(key, val string) (oldval string) {
	p := c.getOrAdd(key)
	oldval = p.Value
	p.Value = val
	return oldval
}

func (c *Configuration) Get(key string) string {
	if n := c.get(key); n != nil {
		return n.Value
	}
	return ""
}

func (c *Configuration) Bytes() []byte {
	t, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err) // should always be valid
	}
	return t
}

func (c *Configuration) String() string {
	return string(c.Bytes())
}

func NewConfigurationFromByte(b []byte) (*Configuration, error) {
	var c Configuration
	if err := xml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func NewConfigurationFromString(txt string) (*Configuration, error) {
	return NewConfigurationFromByte([]byte(txt))
}

func NewConfigurationFromFile(path string) (*Configuration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return NewConfigurationFromByte(b)
}
