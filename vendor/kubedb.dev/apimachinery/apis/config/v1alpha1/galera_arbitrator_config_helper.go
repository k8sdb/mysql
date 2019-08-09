package v1alpha1

import (
	"fmt"
)

const (
	// GarbdListenPort is the port at which Galera Arbitrator Daemon (garbd) listen
	GarbdListenPort = 4444

	// GarbdXtrabackupSSTMethod is the name of the method or script that is
	// used during a State Snapshot Transfer to Galera Arbitrator Daemon (garbd).
	GarbdXtrabackupSSTMethod = "xtrabackup-v2"

	// GarbdXtrabackupSSTRequestSuffix denotes the suffix of sst request string for xtrabackup
	// used by Galera Arbitrator Daemon (garbd)
	GarbdXtrabackupSSTRequestSuffix = "/xtrabackup_sst//1"
	// GarbdLogFile is the name log file at which Galera Arbitrator Daemon (garbd) puts logs
	GarbdLogFile = "/tmp/garb.log"

	// GaleraParamsGarbdListenAddr defines an arbitrary listen socket address
	// that Galera Arbitrator Daemon (garbd) opens to communicate with the cluster
	// https://galeracluster.com/library/documentation/backup-cluster.html
	GaleraParamsGarbdListenAddr = "gmcast.listen_addr=tcp://0.0.0.0:" + string(GarbdListenPort)

	// SOCAT is needed after completing sst by Galera Arbitrator (garbd)
	// SOCATOptionTCPLISTEN is the SOCAT tcp listen option
	SOCATOptionTCPLISTEN = "TCP-LISTEN:" + string(GarbdListenPort)
	// SOCATOptionReUseAddr is the SOCAT reuseaddr option
	SOCATOptionReUseAddr = "reuseaddr"
	// SOCATOptionRetry is the default retry value for `socat` binary
	SOCATOptionRetry = 30
)

// ClusterAddressWithListenOption method returns the galera cluster address with
// the listening option (address at which Galera Cluster listens to connections from
// other nodes) for `--address` option in `garbd`
func (g *GaleraArbitratorConfiguration) ClusterAddressWithListenOption() string {
	if g == nil {
		return ""
	}

	return fmt.Sprintf("%s?%s", g.Address, GaleraParamsGarbdListenAddr)
}

// SSTRequestString method form the sst request string
// for `--sst` option in `garbd`
func (g *GaleraArbitratorConfiguration) SSTRequestString(host string) string {
	if g == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s:%d%s", g.SSTMethod, host, GarbdListenPort, GarbdXtrabackupSSTRequestSuffix)
}

// SOCATOption returns the option string used for `SOCAT` in the
// percona xtradb backup process
func SOCATOption(retry int32) string {
	return fmt.Sprintf("%s,%s,retry=%d", SOCATOptionTCPLISTEN, SOCATOptionReUseAddr, retry)
}
