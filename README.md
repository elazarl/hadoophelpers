hadoophelpers
=============

Collection of helper utilities for hadoop

## hadoopconf

A simple command line utility to get and set configuration of hadoop
clusters.

Get tool with:

    go get github.com/elazarl/hadoophelpers/go/hadoopconf

Usage Example

    # $GOPATH/bin/hadoopconf --conf /opt/hadoop-2.1.0-beta get '*zook*'
    core-default.xml ha.zookeeper.acl                = world:anyone:rwcda
    core-default.xml ha.zookeeper.quorum             =
    core-default.xml ha.zookeeper.session-timeout.ms = 5000
    core-default.xml ha.zookeeper.auth               =
    core-default.xml ha.zookeeper.parent-znode       = /hadoop-ha
    # $GOPATH/bin/hadoopconf --conf /opt/hadoop-2.1.0-beta
    hadoopconf> get *cert
    core-default.xml hadoop.ssl.require.client.cert = false
    hadoopconf> set hadoop.ssl.require.client.cert=true
    hadoopconf>
    # cat /opt/hadoop-2.1.0-beta/etc/hadoop/core-site.xml
    <configuration>
      <property>
        <name>hadoop.ssl.require.client.cert</name>
        <value>true</value>
        <description></description>
      </property>
    </configuration> 

One can also inspect environment variables

    $ $GOPATH/bin/hadoopconf -c /tmp/gohadoopconf-test/hadoop-1.2.1 env '*TRACKER*'
    hadoop-env.sh HADOOP_JOBTRACKER_OPTS  = -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS
    hadoop-env.sh HADOOP_TASKTRACKER_OPTS =
    $ $GOPATH/bin/hadoopconf -c /tmp/gohadoopconf-test/hadoop-1.2.1
    hadoopconf> envadd HADOOP_JOBTRACKER_OPTS -Dfoo
    hadoopconf> env HADOOP_JOBTRACKER_OPTS
    hadoop-env.sh HADOOP_JOBTRACKER_OPTS was -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS
                                         now -Dfoo -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS
    hadoopconf> env HADOOP_JOBTRA<tab>
    HADOOP_JOBTRACKER_OPTS         HADOOP_LOG_DIR                 HADOOP_CLASSPATH               HADOOP_SLAVE_SLEEP             HADOOP_TASKTRACKER_OPTS
    HADOOP_OPTS                    HADOOP_SSH_OPTS                HADOOP_NICENESS                HADOOP_DATANODE_OPTS           HADOOP_SECONDARYNAMENODE_OPTS
    HADOOP_MASTER                  HADOOP_PID_DIR                 JAVA_HOME                      HADOOP_NAMENODE_OPTS
    HADOOP_SLAVES                  HADOOP_HEAPSIZE                HADOOP_IDENT_STRING            HADOOP_BALANCER_OPTS
    hadoopconf> env HADOOP_JOBTRACKER_OPTS
    hadoop-env.sh HADOOP_JOBTRACKER_OPTS = -Dfoo -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS
    hadoopconf> envdel HADOOP_JOBTRACKER_OPTS -Dfoo
    hadoop-env.sh HADOOP_JOBTRACKER_OPTS was -Dfoo -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS
                                         now -Dcom.sun.management.jmxremote $HADOOP_JOBTRACKER_OPTS

