package hadoopconf

import (
	. "github.com/robertkrimen/terst"
	"testing"
)

func TestExampleConf(t *testing.T) {
	Terst(t)
	conf, err := NewConfigurationFromString(coreSite)
	Is(err, nil)
	Is(conf.Get("hadoop.rpc.protection"), "authentication")
	Is(conf.Get("koko"), "")
	Is(conf.Set("koko", "bobo"), "")
	Is(conf.Get("koko"), "bobo")
	ser := conf.String()
	newconf, err := NewConfigurationFromString(ser)
	Is(newconf.Get("momo"), "")
	Is(newconf.Get("koko"), "bobo")
	Is(newconf.Get("hadoop.rpc.protection"), "authentication")
}

func TestDefaultConf(t *testing.T) {
	Terst(t)
	coreSite, err := NewGeneratedConfFromString(Source{"coreSite", Generated}, coreSite)
	Is(err, nil)
	coreDefault, err := NewGeneratedConfFromString(Source{"coreDefault", Generated}, coreDefault)
	Is(err, nil)
	conf := &ConfWithDefault{coreSite, coreDefault}
	Is(conf.Get("nonexitsing.conf"), "")
	Is(conf.Get("custom.property"), "custom.value")
	Is(conf.Get("hadoop.security.authentication"), "kerberos")
	Is(conf.Get("hadoop.common.configuration.version"), "0.23.0")

	var src Source
	_, src = conf.SourceGet("custom.property")
	Is(src.Source, "coreSite")
	_, src = conf.SourceGet("hadoop.security.authentication")
	Is(src.Source, "coreSite")
	_, src = conf.SourceGet("hadoop.common.configuration.version")
	Is(src.Source, "coreDefault")

	v, src := conf.SourceGet("nonexisting.property")
	Is(v, "")
	Is(src, NoSource)
}
