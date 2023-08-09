package pro

import (
	"errors"
)

var KeyNameMap map[string]string
var NameKeyMap map[string][]string

// GetNameByKey 根据Key的前缀获取项目名
func GetNameByKey(key string) string {
	return KeyNameMap[key]
}

// GetKeyByName 根据项目名获取Key的前缀
func GetKeyByName(name string) ([]string, error) {
	if _, ok := NameKeyMap[name]; ok {
		return NameKeyMap[name], nil
	} else {
		return nil, errors.New("没有该项目的Key映射")
	}
}

func init() {
	initKeyNameMap()
	initNameKeyMap()
}

func initKeyNameMap() {
	KeyNameMap = make(map[string]string)
	KeyNameMap["HADOOP"] = "Hadoop"
	KeyNameMap["MAPREDUCE"] = "Hadoop"
	KeyNameMap["YARN"] = "Hadoop"
	KeyNameMap["HDFS"] = "Hadoop"
	KeyNameMap["HBASE"] = "Hbase"
	KeyNameMap["HIVE"] = "Hive"
	KeyNameMap["KAFKA"] = "Kafka"
	KeyNameMap["IMPALA"] = "Impala"
	KeyNameMap["AMBARI"] = "Ambari"
	KeyNameMap["DAFFODIL"] = "Daffodil"
	KeyNameMap["OAK"] = "Jackrabbit-Oak"
	KeyNameMap["SVN"] = "Subversion"
	KeyNameMap["WICKET"] = "Wicket"
	KeyNameMap["ARROW"] = "Arrow"
	KeyNameMap["CB"] = "Cordova"
	KeyNameMap["AMQ"] = "ActiveMQ"
	KeyNameMap["GUACAMOLE"] = "Guacamole"
	KeyNameMap["TS"] = "Traffic Server"
	KeyNameMap["CLOUDSTACK"] = "CloudStack"
	KeyNameMap["MESOS"] = "Mesos"
	KeyNameMap["DRILL"] = "Drill"
	KeyNameMap["LUCENE"] = "Lucene"
	KeyNameMap["SOLR"] = "Solr"
	KeyNameMap["HUDI"] = "Hudi"
	KeyNameMap["IGNITE"] = "Ignite"
	KeyNameMap["NIFI"] = "NiFi"
	KeyNameMap["OFBIZ"] = "OFBiz"
	KeyNameMap["SPARK"] = "Spark"
	KeyNameMap["AXIS2"] = "Axis2"
	KeyNameMap["GROOVY"] = "Groovy"
	KeyNameMap["THRIFT"] = "Thrift"
	KeyNameMap["GEODE"] = "Geode"
	KeyNameMap["NETBEANS"] = "Netbeans"
	KeyNameMap["HDDS"] = "Ozone"
	KeyNameMap["CAMEL"] = "Camel"
	KeyNameMap["FLINK"] = "Flink"
	KeyNameMap["QPID"] = "Qpid"
}

func initNameKeyMap() {
	NameKeyMap = make(map[string][]string)
	NameKeyMap["hadoop"] = append(NameKeyMap["hadoop"], []string{"HADOOP", "MAPREDUCE", "YARN", "HDFS"}...)
	NameKeyMap["hbase"] = append(NameKeyMap["hbase"], []string{"HADOOP", "HBASE"}...)
	NameKeyMap["hive"] = append(NameKeyMap["hive"], []string{"HADOOP", "HIVE"}...)
	NameKeyMap["kafka"] = append(NameKeyMap["kafka"], "KAFKA")
	NameKeyMap["impala"] = append(NameKeyMap["impala"], "IMPALA")
	NameKeyMap["ambari"] = append(NameKeyMap["ambari"], "AMBARI")
	NameKeyMap["daffodil"] = append(NameKeyMap["daffodil"], "DAFFODIL")
	NameKeyMap["jackrabbit-oak"] = append(NameKeyMap["jackrabbit-oak"], "OAK")
	NameKeyMap["subversion"] = append(NameKeyMap["subversion"], "SVN")
	NameKeyMap["wicket"] = append(NameKeyMap["wicket"], "WICKET")
	NameKeyMap["arrow"] = append(NameKeyMap["arrow"], "ARROW")
	NameKeyMap["cordova"] = append(NameKeyMap["cordova"], "CB")
	NameKeyMap["activemq"] = append(NameKeyMap["activemq"], "AMQ")
	NameKeyMap["guacamole"] = append(NameKeyMap["guacamole"], "GUACAMOLE")
	NameKeyMap["trafficserver"] = append(NameKeyMap["trafficserver"], "TS")
	NameKeyMap["cloudstack"] = append(NameKeyMap["cloudstack"], "CLOUDSTACK")
	NameKeyMap["mesos"] = append(NameKeyMap["mesos"], "MESOS")
	NameKeyMap["drill"] = append(NameKeyMap["drill"], "DRILL")
	NameKeyMap["lucene"] = append(NameKeyMap["lucene"], []string{"LUCENE", "SOLR"}...)
	NameKeyMap["solr"] = append(NameKeyMap["solr"], []string{"LUCENE", "SOLR"}...)
	NameKeyMap["hudi"] = append(NameKeyMap["hudi"], "HUDI")
	NameKeyMap["ignite"] = append(NameKeyMap["ignite"], "IGNITE")
	NameKeyMap["nifi"] = append(NameKeyMap["nifi"], "NIFI")
	NameKeyMap["ofbiz"] = append(NameKeyMap["ofbiz"], "OFBIZ")
	NameKeyMap["spark"] = append(NameKeyMap["spark"], "SPARK")
	NameKeyMap["axis2"] = append(NameKeyMap["axis2"], "AXIS2")
	NameKeyMap["groovy"] = append(NameKeyMap["groovy"], "GROOVY")
	NameKeyMap["thrift"] = append(NameKeyMap["thrift"], "THRIFT")
	NameKeyMap["geode"] = append(NameKeyMap["geode"], "GEODE")
	NameKeyMap["netbeans"] = append(NameKeyMap["netbeans"], "NETBEANS")
	NameKeyMap["ozone"] = append(NameKeyMap["ozone"], "HDDS")
	NameKeyMap["camel"] = append(NameKeyMap["camel"], "CAMEL")
	NameKeyMap["flink"] = append(NameKeyMap["flink"], "FLINK")
	NameKeyMap["qpid"] = append(NameKeyMap["qpid"], "QPID")
}
