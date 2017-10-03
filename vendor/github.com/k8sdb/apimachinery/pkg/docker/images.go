package docker

const (
	ImageOperator         = "kubedb/operator"
	ImagePostgresOperator = "kubedb/pg-operator"
	ImagePostgres         = "kubedb/postgres"
	ImageMySQLOperator    = "maruftuhin/mysql-operator" //TESTING
	//todo:
	// ImageMySQLOperator    = "kubedb/mysql-operator"
	ImageMySQL            = "maruftuhin/mysql"          //TESTING
	//todo:
	// ImageMySQL            = "library/mysql"
	ImageElasticOperator  = "kubedb/es-operator"
	ImageElasticsearch    = "kubedb/elasticsearch"
	ImageElasticdump      = "kubedb/elasticdump"
)

const (
	OperatorName       = "kubedb-operator"
	OperatorContainer  = "operator"
	OperatorPortName   = "web"
	OperatorPortNumber = 8080
)
