package hadoopconf

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/robertkrimen/terst"
)

var tempDir = "/tmp/gohadoopconf-test"

var hadoops = []string{
	hadoop2,
	hadoop1,
}

const (
	hadoop2 = "hadoop-2.1.0-beta"
	hadoop1 = "hadoop-1.2.1"
)

func bash(script string) {
	cmd := exec.Command("bash", "-x", "-c", script)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		log.Fatalln("error running script:", script, err)
	}
}

func init() {
	os.MkdirAll(filepath.Join(tempDir, "tmp"), 0755)
	if stat, err := os.Stat(filepath.Join(tempDir, hadoop2)); os.IsNotExist(err) {
		bash(`FILE="` + hadoop2 + `";
		(cd tmp && curl -vL "http://apache.mivzakim.net/hadoop/common/${FILE}/${FILE}.tar.gz"|tar xzf -) \
		&& cp -r tmp/$FILE/etc tmp/$FILE/etc_orig && mv tmp/"$FILE" .`)
	} else if !stat.IsDir() {
		panic(filepath.Join(tempDir, stat.Name()) + " is file instead of dir")
	}
	if stat, err := os.Stat(filepath.Join(tempDir, hadoop1)); os.IsNotExist(err) {
		bash(`FILE="` + hadoop1 + `";
		(cd tmp && curl -vL "http://apache.mivzakim.net/hadoop/common/${FILE}/${FILE}-bin.tar.gz"|tar xzf -) \
		&& cp -r tmp/$FILE/conf tmp/$FILE/conf_orig && mv tmp/"$FILE" .`)
	} else if !stat.IsDir() {
		panic(filepath.Join(tempDir, stat.Name()) + " is file instead of dir")
	}
}

func restoreConf() {
	bash("rm -rf " + filepath.Join(tempDir, hadoop2, "etc") + " " + filepath.Join(tempDir, hadoop1, "conf"))
	bash("cp -r " + filepath.Join(tempDir, hadoop2, "etc_orig") + " " + filepath.Join(tempDir, hadoop2, "etc"))
	bash("cp -r " + filepath.Join(tempDir, hadoop1, "conf_orig") + " " + filepath.Join(tempDir, hadoop1, "conf"))
}

func FailOnErr(err error) {
	if err != nil {
		FailNow(err)
	}
}

type expectValSrc struct {
	v   string
	src Source
}

func ValSrc(v string, src Source) expectValSrc {
	return expectValSrc{v, src}
}

func (vs expectValSrc) Is(v, src string) {
	Is(vs.v, v)
	if src == "" {
		Is(vs.src, NoSource)
	} else {
		Is(filepath.Base(vs.src.Source), src)
	}
}

func (vs expectValSrc) Empty() {
	vs.Is("", "")
}

func TestBasicHadoopConf(t *testing.T) {
	Terst(t)

	jars, err := Jars(filepath.Join(tempDir, hadoop2))
	FailOnErr(err)
	c, err := New(filepath.Join(tempDir, hadoop2), jars)
	FailOnErr(err)
	ValSrc(c.SourceGet("hadoop.common.configuration.version")).Is("0.23.0", "core-default.xml")
	ValSrc(c.SourceGet("hadoop.hdfs.configuration.version")).Is("1", "hdfs-default.xml")
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("32768", "hdfs-default.xml")
	ValSrc(c.SourceGet("mapreduce.jobtracker.jobhistory.task.numberprogresssplits")).Is("12", "mapred-default.xml")
	ValSrc(c.SourceGet("yarn.ipc.serializer.type")).Is("protocolbuffers", "yarn-default.xml")

	c.SetIfExist("dfs.default.chunk.view.size", "1")
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("1", "hdfs-site.xml")
	Is(c.CoreSite.Set("in.core.site", "right here"), "")
	ValSrc(c.SourceGet("in.core.site")).Is("right here", "core-site.xml")

	jars, err = Jars(filepath.Join(tempDir, hadoop1))
	FailOnErr(err)
	c, err = New(filepath.Join(tempDir, hadoop1), jars)
	FailOnErr(err)
	ValSrc(c.SourceGet("mapred.job.shuffle.input.buffer.percent")).Is("0.70", "mapred-default.xml")
	ValSrc(c.SourceGet("io.sort.factor")).Is("10", "mapred-default.xml")
	ValSrc(c.SourceGet("hadoop.common.configuration.version")).Empty()
	ValSrc(c.SourceGet("hadoop.hdfs.configuration.version")).Empty()
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("32768", "hdfs-default.xml")

	c.SetIfExist("dfs.default.chunk.view.size", "1")
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("1", "hdfs-site.xml")
	Is(c.CoreSite.Set("in.core.site", "right here"), "")
	ValSrc(c.SourceGet("in.core.site")).Is("right here", "core-site.xml")
}

func Val(v string, conf ConfSourcer) string {
	return v
}

func TestHadoopConfWrite(t *testing.T) {
	Terst(t)
	defer restoreConf()

	jars, err := Jars(filepath.Join(tempDir, hadoop2))
	FailOnErr(err)
	c, err := New(filepath.Join(tempDir, hadoop2), jars)
	FailOnErr(err)
	ValSrc(c.SourceGet("hadoop.common.configuration.version")).Is("0.23.0", "core-default.xml")
	Is(Val(c.SetIfExist("hadoop.common.configuration.version", "oldie")), "0.23.0")
	Is(Val(c.SetIfExist("hadoop.hdfs.configuration.version", "1")), "1")
	Is(Val(c.SetIfExist("yarn.ipc.serializer.type", "writables")), "protocolbuffers")
	Is(c.Save(), nil)
	_, err = os.Stat(filepath.Join(tempDir, hadoop2, "etc", "hadoop", "mapred-site.xml"))
	Is(os.IsNotExist(err), true)

	c, err = New(filepath.Join(tempDir, hadoop2), jars)
	FailOnErr(err)
	ValSrc(c.SourceGet("hadoop.common.configuration.version")).Is("oldie", "core-site.xml")
	ValSrc(c.SourceGet("hadoop.hdfs.configuration.version")).Is("1", "hdfs-site.xml")
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("32768", "hdfs-default.xml")
	ValSrc(c.SourceGet("mapreduce.jobtracker.jobhistory.task.numberprogresssplits")).Is("12", "mapred-default.xml")
	Is(Val(c.SetIfExist("mapreduce.jobtracker.jobhistory.task.numberprogresssplits", "14")), "12")
	ValSrc(c.SourceGet("yarn.ipc.serializer.type")).Is("writables", "yarn-site.xml")
	Is(c.Save(), nil)
	_, err = os.Stat(filepath.Join(tempDir, hadoop2, "etc", "hadoop", "mapred-site.xml"))
	Is(err, nil)

	jars, err = Jars(filepath.Join(tempDir, hadoop1))
	FailOnErr(err)
	c, err = New(filepath.Join(tempDir, hadoop1), jars)
	FailOnErr(err)
	ValSrc(c.SourceGet("mapred.job.shuffle.input.buffer.percent")).Is("0.70", "mapred-default.xml")
	ValSrc(c.SourceGet("io.sort.factor")).Is("10", "mapred-default.xml")
	ValSrc(c.SourceGet("hadoop.common.configuration.version")).Empty()
	ValSrc(c.SourceGet("hadoop.hdfs.configuration.version")).Empty()
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("32768", "hdfs-default.xml")

	c.SetIfExist("dfs.default.chunk.view.size", "1")
	ValSrc(c.SourceGet("dfs.default.chunk.view.size")).Is("1", "hdfs-site.xml")
	Is(c.CoreSite.Set("in.core.site", "right here"), "")
	ValSrc(c.SourceGet("in.core.site")).Is("right here", "core-site.xml")
}
