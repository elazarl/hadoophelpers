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
