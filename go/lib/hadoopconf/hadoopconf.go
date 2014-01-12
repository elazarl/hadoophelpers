package hadoopconf

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type HadoopConf struct {
	multiSourceConf
	CoreSite   *ConfWithDefault
	HdfsSite   *ConfWithDefault
	MapredSite *ConfWithDefault
	YarnSite   *ConfWithDefault
}

type HadoopDefaultConf struct {
	CoreSite   ConfSourcer
	HdfsSite   ConfSourcer
	MapredSite ConfSourcer
	YarnSite   ConfSourcer
}

func (c *HadoopConf) Save(backup bool) error {
	for _, conf := range []*ConfWithDefault{c.CoreSite, c.HdfsSite, c.MapredSite, c.YarnSite} {
		if err := conf.Conf.(*FileConfiguration).Save(backup); err != nil {
			return err
		}
	}
	return nil
}

func FromConf(coreSite *ConfWithDefault, hdfsSite *ConfWithDefault,
	mapredSite *ConfWithDefault, yarnSite *ConfWithDefault) *HadoopConf {
	confs := []ConfSourcer{coreSite, hdfsSite}
	if yarnSite != nil {
		confs = append(confs, yarnSite)
	}
	if mapredSite != nil {
		confs = append(confs, mapredSite)
	}
	return &HadoopConf{confs, coreSite, hdfsSite, mapredSite, yarnSite}
}

func anyRegexpMatch(s string, res []*regexp.Regexp) bool {
	for _, re := range res {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func ConfsFromJar(jar string, files ...string) ([]ConfSourcer, error) {
	rv := make([]ConfSourcer, len(files))
	r, err := zip.OpenReader(jar)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	m := make(map[string]int)
	for i, name := range files {
		m[name] = i
	}
	for _, f := range r.File {
		if ix, ok := m[f.Name]; ok {
			r, err := f.Open()
			if err != nil {
				return nil, err
			}
			b, err := ioutil.ReadAll(r)
			if err != nil {
				return nil, err
			}
			if rv[ix], err = NewGeneratedConfFromBytes(Source{filepath.Join(jar, f.Name), FileFromJar}, b); err != nil {
				return nil, err
			}
		}
	}
	return rv, nil
}

func ConfFromJar(jar string, file string) (ConfSourcer, error) {
	confs, err := ConfsFromJar(jar, file)
	if err != nil {
		return nil, err
	}
	if confs[0] == nil {
		return nil, errors.New("Cannot find " + file + " in " + jar)
	}
	return confs[0], nil
}

func globRegexp(basedirs []string, res []*regexp.Regexp) (string, error) {
	for _, glob := range basedirs {
		dirs, err := filepath.Glob(glob)
		if err != nil {
			return "", err
		}
		for _, dir := range dirs {
			jars, err := ioutil.ReadDir(dir)
			if err != nil {
				return "", err
			}
			for _, jar := range jars {
				if anyRegexpMatch(jar.Name(), res) {
					return filepath.Join(dir, jar.Name()), nil
				}
			}
		}
	}
	return "", errors.New(fmt.Sprintln("in", basedirs, "jar was not found", res))
}

func getDefault(filename string, res []*regexp.Regexp, basedirs ...string) (ConfSourcer, error) {
	jar, err := globRegexp(basedirs, res)
	if err != nil {
		return nil, err
	}
	return ConfFromJar(jar, filename)
}

func getCoreDefault(basedirs ...string) (ConfSourcer, error) {
	hadoopCommonRegexp := []*regexp.Regexp{
		regexp.MustCompile(`hadoop-(common|core)-[0-9.]+-?([a-zA-Z0-9._]+)?\.jar`),
	}
	jar, err := globRegexp(basedirs, hadoopCommonRegexp)
	if err != nil {
		return nil, err
	}
	return ConfFromJar(jar, "core-default.xml")
}

func getConf(confName string, globs ...string) (*FileConfiguration, error) {
	for _, glob := range globs {
		dirs, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		for _, dir := range dirs {
			conf, err := NewFileConfiguration(filepath.Join(dir, confName))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			return conf, nil
		}
	}
	return nil, errors.New(fmt.Sprintln("cannot find file", confName, "in any of", globs))
}

type returnPanic struct {
	err error
}

func re(pats ...string) []*regexp.Regexp {
	res := []*regexp.Regexp{}
	for _, pat := range pats {
		res = append(res, regexp.MustCompile(pat))
	}
	return res
}

func Jars(basedir string) (*HadoopDefaultConf, error) {
	coreDefault, err := getDefault("core-default.xml", re(`hadoop-(common|core)-[0-9.]+-?([a-zA-Z0-9._]+)?\.jar`, "hadoop-common.jar"), basedir,
		filepath.Join(basedir, "share/hadoop/common"),
		"/usr/lib/hadoop",
		"/share/hadoop/common")
	if err != nil {
		return nil, err
	}
	hdfsDefault, err := getDefault("hdfs-default.xml", re(`hadoop-(hdfs|core)-[0-9.]+-?([a-zA-Z0-9._]+)?\.jar`), basedir,
		filepath.Join(basedir, "share/hadoop/hdfs"),
		filepath.Join(basedir, "hadoop-hdfs"),
		"/share/hadoop/hdfs",
		"/usr/lib/hadoop-hdfs")
	if err != nil {
		return nil, err
	}
	// dfs.*.{kerberos,https}.principal dfs.*.keytab.file does not appear in hdfs-default.xml for some reason
	for _, role := range []string {"namenode", "namenode.secondary", "datanode"} {
		hdfsDefault.Set("dfs." + role + ".keytab.file", "")
		hdfsDefault.Set("dfs." + role + ".kerberos.principal", "")
		hdfsDefault.Set("dfs." + role + ".https.principal", "")
	}
	hdfsDefault.Set("dfs.datanode.hostname", "")
	mapredDefault, err := getDefault("mapred-default.xml", re(`hadoop-(mapreduce-client-)?core-[0-9.]+-?([a-zA-Z0-9._]+)?\.jar`), basedir,
		filepath.Join(basedir, "hadoop-0.20-mapreduce"),
		filepath.Join(basedir, "hadoop-mapreduce"),
		filepath.Join(basedir, "share/hadoop/mapreduce"),
		"/share/hadoop/mapreduce",
		"/usr/lib/hadoop-0.20-mapreduce",
		"/usr/lib/hadoop-mapreduce")
	if mapredDefault == nil {
		fmt.Println("got", err)
	}
	yarnDefault, _ := getDefault("yarn-default.xml", re(`hadoop-yarn-common-[0-9.]+-?([a-zA-Z0-9._]+)?\.jar`), basedir,
		filepath.Join(basedir, "share/hadoop/yarn"),
		filepath.Join(basedir, "hadoop-yarn"),
		"/usr/lib/hadoop-yarn",
		"/share/hadoop/yarn")
	return &HadoopDefaultConf{
		CoreSite:   coreDefault,
		HdfsSite:   hdfsDefault,
		MapredSite: mapredDefault,
		YarnSite:   yarnDefault,
	}, nil
}

func New(basedir string, defaultConf *HadoopDefaultConf) (conf *HadoopConf, err error) {
	j := func(s string) string {
		return filepath.Join(basedir, s)
	}
	coreSite, err := getConf("core-site.xml", j("etc/hadoop"), j("conf"), basedir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(coreSite.Path); err != nil {
		return nil, err
	}
	hdfsSite, err := getConf("hdfs-site.xml", j("etc/hadoop"), j("conf"), basedir)
	if err != nil {
		return nil, err
	}
	mapredSite, _ := getConf("mapred-site.xml", j("etc/hadoop"), j("conf"), basedir)
	yarnSite, _ := getConf("yarn-site.xml", j("etc/hadoop"), j("conf"), basedir)
	return FromConf(&ConfWithDefault{Default: defaultConf.CoreSite, Conf: coreSite},
		&ConfWithDefault{Default: defaultConf.HdfsSite, Conf: hdfsSite},
		&ConfWithDefault{Default: defaultConf.MapredSite, Conf: mapredSite},
		&ConfWithDefault{Default: defaultConf.YarnSite, Conf: yarnSite},
	), nil
}
