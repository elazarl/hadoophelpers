package hadoopconf

import (
	"path/filepath"
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestEnvExportParse(t *testing.T) {
	Terst(t)
	v := parseExport(0, `export HADOOP_NAMENODE_OPTS="-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS"`)
	if IsNot(v, nil) {
		Is(v.Name, "HADOOP_NAMENODE_OPTS")
		Is(v.Val, "-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS")
	}
	v = parseExport(0, `export HADOOP_OPTS=${foo:"bar gar"}`)
	if IsNot(v, nil) {
		Is(v.Name, "HADOOP_OPTS")
		Is(v.Val, "${foo:\"bar gar\"}")
	}
}

func TestHadoopEnv(t *testing.T) {
	Terst(t)
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

func TestHadoopEnvWrite(t *testing.T) {
	Terst(t)
	defer restoreConf()
	env, err := NewEnv(filepath.Join(tempDir, hadoop1))
	FailOnErr(err)
	Is(len(env), 1)
	env.Get("HADOOP_NAMENODE_OPTS").Append("-Xms100")
	env.Get("HADOOP_OPTS").Append("-Xms100")

	Is(env.Get("HADOOP_NAMENODE_OPTS").Val, "-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS -Xms100")
	Is(env.Get("HADOOP_OPTS").Val, "-Xms100")
	Is(env.Get("HADOOP_OPT"), (*Var)(nil))
	FailOnErr(env.Save())
	// reevaluate tests after loading file from disk
	env, err = NewEnv(filepath.Join(tempDir, hadoop1))
	FailOnErr(err)
	Is(env.Get("HADOOP_NAMENODE_OPTS").Val, "-Dcom.sun.management.jmxremote $HADOOP_NAMENODE_OPTS -Xms100")
	Is(env.Get("HADOOP_OPTS").Val, "-Xms100")
	Is(env.Get("HADOOP_OPT"), (*Var)(nil))

	env, err = NewEnv(filepath.Join(tempDir, hadoop2))
	FailOnErr(err)
	Is(len(env), 4)
	env.Get("HADOOP_CLIENT_OPTS").Update("-Xmx", "-Xmx1024m")
	Is(env.Get("HADOOP_CLIENT_OPTS").Val, "-Xmx1024m $HADOOP_CLIENT_OPTS")
	env.Get("JSVC_HOME").Update("/home/jsvc", "/home/jsvc")
	Is(env.Get("JSVC_HOME").Val, "/home/jsvc")
	FailOnErr(env.Save())
	// reevaluate tests after loading file from disk
	env, err = NewEnv(filepath.Join(tempDir, hadoop2))
	FailOnErr(err)
	Is(env.Get("HADOOP_CLIENT_OPTS").Val, "-Xmx1024m $HADOOP_CLIENT_OPTS")
	Is(env.Get("JSVC_HOME").Val, "/home/jsvc")
}
