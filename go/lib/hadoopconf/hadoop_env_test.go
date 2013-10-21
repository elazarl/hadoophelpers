package hadoopconf

import (
	"path/filepath"
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestEnvRegexp(t *testing.T) {
	Terst(t)
	m := exportLine.FindStringSubmatch(`export HADOOP_NAMENODE_OPTS="-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS"`)
	if !IsNot(len(m), 0) {
		return
	}
	Is(m[1], "HADOOP_NAMENODE_OPTS")
	Is(m[2], "-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS")
}

func TestHadoopEnv(t *testing.T) {
	Terst(t)
	defer restoreConf()
	env, err := NewEnv(filepath.Join(tempDir, hadoop1))
	FailOnErr(err)
	Is(len(env), 1)
	Is(env.Get("HADOOP_NAMENODE_OPTS").Val, "-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS")
	Is(env.Get("HADOOP_OPTS").Val, "")
	Is(env.Get("HADOOP_OPT"), (*Var)(nil))

	env, err = NewEnv(filepath.Join(tempDir, hadoop2))
	FailOnErr(err)
	Is(len(env), 4)
	Is(env.Get("HADOOP_CLIENT_OPTS").Val, "-Xmx512m $HADOOP_CLIENT_OPTS")
	Is(env.Get("HADOOP_OPTS").Val, "$HADOOP_OPTS -Djava.net.preferIPv4Stack=true")
	Is(env.Get("JSVC_HOME").Val, "")
	Is(env.Get("HADOOP_JOB_HISTORYSERVER_HEAPSIZE").Val, "1000")
	Is(env.Get("HADOOP_OPT"), (*Var)(nil))
}
