#!/usr/bin/env bash

#set -eoux pipefail

script_name=${0##*/}
NAMESPACE="$POD_NAMESPACE"
USER="$MYSQL_ROOT_USERNAME"
PASSWORD="$MYSQL_ROOT_PASSWORD"

function timestamp() {
    date +"%Y/%m/%d %T"
}

function log() {
    local type="$1"
    local msg="$2"
    echo "$(timestamp) [$script_name] [$type] $msg"
}

function get_host_name() {
    echo -n "$BASE_NAME-$1.$GOV_SVC.$NAMESPACE.svc.cluster.local"
}

# wait_for_each_instance_be_ready( group_size, base_name, gov_svc, ns)
# gr_size   = size of the replication group
# base_name = the name of MySQL CRD object
# gov_svc   = the governing service for the StatefulSet (representing the group)
# ns        = the namespace
function wait_for_each_instance_be_ready() {
    local gr_size=$1
    local base_name=$2
    local gov_svc=$3
    local ns=$4

    while true; do
        ready=0
        for ((i=0;i<$gr_size;i++)); do
            out=`hostname`
#            out=`kubectl exec -it -n $ns $base_name-$i -c mysql -- hostname`
            if [[ "$out" = "" ]]; then
                ready=1
                break
            fi
        done

        if [[ $ready = 0 ]]; then
            break
        fi
    done

    log "INFO" "All servers are ready :)"
}

# get_ip_whitelist( group_size, base_name, gov_svc, ns)
# gr_size   = size of the group
# base_name = the name of MySQL CRD object
# gov_svc   = the governing service for the StatefulSet (representing the group)
# ns        = the namespace
function get_ip_whitelist() {
    local gr_size=$1
    local base_name=$2
    local gov_svc=$3
    local ns=$4

    for ((i=0;i<$gr_size;i++)); do
        if (("$i" > "0")); then
            echo -n ",";
        fi
#        echo -n "$base_name-$i.$gov_svc.$ns.svc"
        get_host_name "$i"
    done
}

# ==============
# export GROUP_SIZE=5; export BASE_NAME=my-galera; export GOV_SVC=kubedb-gvr; export NAMESPACE=demo
# ==============

echo "GROUP_SIZE=$GROUP_SIZE"
echo "BASE_NAME=$BASE_NAME"
echo "GOV_SVC=$GOV_SVC"
echo "NAMESPACE=$POD_NAMESPACE"
echo "GROUP_NAME=$GROUP_NAME"
echo ""
echo "USER=$USER"
echo "PASSWORD=$PASSWORD"
echo ""

cur_hostname=$(hostname)
export cur_host=
log "INFO" "Reading standard input..."
while read -ra line; do
#    echo ">>>>>>>>> line: $line"
    if [[ "${line}" == *"${cur_hostname}"* ]]; then
        cur_host="$line"
        log "INFO" "I am $cur_host"
    fi
    peers=("${peers[@]}" "$line")
done
#echo "================= args: '${peers[*]}'"

#wait_for_each_instance_be_ready ${GROUP_SIZE} ${BASE_NAME} ${GOV_SVC} ${NAMESPACE}

#export hosts=`get_ip_whitelist ${GROUP_SIZE} ${BASE_NAME} ${GOV_SVC} ${NAMESPACE}`
export hosts=`echo -n ${peers[*]} | sed -e "s/ /,/g"`
#echo ">>>>>> ips: '$hosts'"
export seeds=`echo -n ${hosts} | sed -e "s/,/:33060,/g" && echo -n ":33060"`
#echo ">>>>>> seeds: $seeds"
declare -i srv_id=`hostname | sed -e "s/${BASE_NAME}-//g"`
((srv_id+=1))
#echo ">>>>>> srv_id: $srv_id"
#export cur_host=`echo -n "$(hostname).${GOV_SVC}.${NAMESPACE}.svc.cluster.local"`
#echo ">>>>>> cur_host: $cur_host"
export cur_addr="${cur_host}:33060"
#echo ">>>>>> cur_addr: $cur_addr"

#echo "/etc/mysql/my.cnf contents are as follows:"
log "INFO" "Storing default mysqld config into /etc/mysql/my.cnf"
cat >> /etc/mysql/my.cnf <<EOL

[mysqld]

# General replication settings
gtid_mode = ON
enforce_gtid_consistency = ON
master_info_repository = TABLE
relay_log_info_repository = TABLE
binlog_checksum = NONE
log_slave_updates = ON
log_bin = binlog
binlog_format = ROW
transaction_write_set_extraction = XXHASH64
loose-group_replication_bootstrap_group = OFF
loose-group_replication_start_on_boot = OFF
loose-group_replication_ssl_mode = REQUIRED
loose-group_replication_recovery_use_ssl = 1

# Shared replication group configuration
loose-group_replication_group_name = "${GROUP_NAME}"
#loose-group_replication_ip_whitelist = "${hosts}"
loose-group_replication_group_seeds = "${seeds}"

# Single or Multi-primary mode? Uncomment these two lines
# for multi-primary mode, where any host can accept writes
#loose-group_replication_single_primary_mode = OFF
#loose-group_replication_enforce_update_everywhere_checks = ON

# Host specific replication configuration
server_id = ${srv_id}
bind-address = "${cur_host}"
report_host = "${cur_host}"
loose-group_replication_local_address = "${cur_addr}"
EOL

## TODO: remove this
#echo ""
#cat /etc/mysql/my.cnf
#echo ""

log "INFO" "Starting mysql server..."
/etc/init.d/mysql stop
docker-entrypoint.sh mysqld >/dev/null 2>&1 &
pid=$!
echo $pid
sleep 5

for host in ${peers[*]}; do
    for i in {60..0} ; do
        out=`mysqladmin -u ${USER} --password=${PASSWORD} --host=${host} ping 2> /dev/null`
        if [[ "$out" == "mysqld is alive" ]]; then
            sleep 5
            break
        fi

        echo -n .
        sleep 1
    done

    if [[ "$i" = "0" ]]; then
        log "ERROR" "Server ${host} start failed..."
        exit 1
    fi
done

log "INFO" "All servers (${peers[*]}) are ready"

export mysql_header="mysql -u ${USER} --password=${PASSWORD}"
export member_hosts=( `echo -n ${hosts} | sed -e "s/,/ /g"` )

for host in ${member_hosts[*]} ; do
#    echo ">>>>>>> host: $host"
    log "INFO" "Initializing the server (${host})..."

    mysql="$mysql_header --host=$host"
#    echo "+++++++++++ $mysql"

    out=`${mysql} -N -e "select count(host) from mysql.user where mysql.user.user='repl';" 2> /dev/null | awk '{print$1}'`
    if [[ "$out" -eq "0" ]]; then
        is_new=("${is_new[@]}" "1")

        log "INFO" "Replication user not found and creating one..."
        ${mysql} -N -e "SET SQL_LOG_BIN=0;" 2> /dev/null
        ${mysql} -N -e "CREATE USER 'repl'@'%' IDENTIFIED BY 'password' REQUIRE SSL;" 2> /dev/null
        ${mysql} -N -e "GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';" 2> /dev/null
        ${mysql} -N -e "FLUSH PRIVILEGES;" 2> /dev/null
        ${mysql} -N -e "SET SQL_LOG_BIN=1;" 2> /dev/null

    #    echo ">>>> plugins are as follows:"
    #    out=`${mysql_header} -N -e "SHOW PLUGINS;" 2> /dev/null || true`
    #    echo ">>>> plugins= $out"
    else
        log "INFO" "Replication user info exists"
        is_new=("${is_new[@]}" "0")
    fi

    #mysql -u root --password=uWuj7-dbvefZVnJx
    ## mysql>
    #SET SQL_LOG_BIN=0;
    #CREATE USER 'repl'@'%' IDENTIFIED BY 'password' REQUIRE SSL;
    #GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
    #FLUSH PRIVILEGES;
    ## -> from here
    #SET SQL_LOG_BIN=1;

    ${mysql} -N -e "CHANGE MASTER TO MASTER_USER='repl', MASTER_PASSWORD='password' FOR CHANNEL 'group_replication_recovery';" 2> /dev/null
    out=`${mysql} -N -e 'SHOW PLUGINS;' 2> /dev/null | grep group_replication`
#    echo ">>>>>>>>> plugins: $out"
    if [[ -z "$out" ]]; then
        log "INFO" "Installing group replication plugin..."
        ${mysql} -e "INSTALL PLUGIN group_replication SONAME 'group_replication.so';" 2> /dev/null
    else
        log "INFO" "Already group replication plugin is installed"
    fi

##     TODO: it is optional. So remove this
#    sleep 5
#    echo ">>>> plugins are as follows:"
#    out=`${mysql} -N -e "SHOW PLUGINS;" 2> /dev/null`
#    echo ">>>> plugins $out"
done

function find_group() {
    # TODO: Need to handle for multiple group existence
    group_found=0
    for host in $@; do

        export mysql="$mysql_header --host=${host}"
        # value may be 'UNDEFINED'
        primary_id=`${mysql} -N -e "SHOW STATUS WHERE Variable_name = 'group_replication_primary_member';" 2>/dev/null | awk '{print $2}'`
#        ${mysql_header} -N -e "SELECT MEMBER_PORT FROM performance_schema.replication_group_members;"

        if [[ -n "$primary_id" ]]; then
            ids=( `${mysql} -N -e "SELECT MEMBER_ID FROM performance_schema.replication_group_members WHERE MEMBER_STATE = 'ONLINE' OR MEMBER_STATE = 'RECOVERING';" 2>/dev/null` )

            for id in ${ids[@]}; do
                if [[ "${primary_id}" == "${id}" ]]; then
                    group_found=1
                    primary_host=`${mysql} -N -e "SELECT MEMBER_HOST FROM performance_schema.replication_group_members WHERE MEMBER_ID = '${primary_id}';" 2>/dev/null | awk '{print $1}'`

                    break
                fi
            done
        fi

        if [[ "$group_found" -eq "1" ]]; then
            break
        fi
    done

    echo -n "${group_found}"
}

export primary_host=`get_host_name 0`
export found=`find_group ${member_hosts[*]}`
primary_idx=`echo ${primary_host} | sed -e "s/[a-z.-]//g"`

log "INFO" "Checking whether there exists any replication group or not..."
if [[ "$found" = "0" ]]; then
    mysql="$mysql_header --host=$primary_host"
#    echo "+++++++++++ $mysql"

#    out=`${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members;"`
    out=`${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_HOST = '$primary_host';"`
#    echo ">>>>>>> state: $out"
    if [[ -z "$out" || "$out" = "OFFLINE" ]]; then
        log "INFO" "No group is found and bootstrapping one on host '$primary_host'..."

        ${mysql} -N -e "STOP GROUP_REPLICATION;" # 2>/dev/null
#        ${mysql} -N -e 'SET GLOBAL group_replication_ip_whitelist="$hosts"' # 2>/dev/null
#        ${mysql} -N -e 'SET GLOBAL group_replication_group_seeds="$seeds"' # 2>/dev/null
#        echo "
#        ++++++++++ primary_idx: $primary_idx
#        ++++++++++ is_new[$primary_idx]: ${is_new[$primary_idx]}
#        "
        if [[ "${is_new[$tmp]}" -eq "1" ]]; then
            log "INFO" "RESET MASTER in primary host $primary_host..."
            ${mysql} -N -e "RESET MASTER;" # 2>/dev/null
        fi
#        ${mysql} -N -e "RESET MASTER;" # 2>/dev/null
        ${mysql} -N -e "SET GLOBAL group_replication_bootstrap_group=ON;" # 2>/dev/null
        ${mysql} -N -e "START GROUP_REPLICATION;" # 2>/dev/null
        ${mysql} -N -e "SET GLOBAL group_replication_bootstrap_group=OFF;" # 2>/dev/null
    else
        log "INFO" "No group is found and member state is unknown on host '$primary_host'..."
    fi
else
    log "INFO" "A group is found and the primary host is '$primary_host'..."
fi

declare -i tmp=0
for host in ${member_hosts[*]}; do
    if [[ "$host" != "$primary_host" ]]; then
        mysql="$mysql_header --host=$host"
#        echo "+++++++++++ $mysql"

#        out=`${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members;"`
        out=`${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_HOST = '$host';"`
#        echo ">>>>>>> state: $out"
        if [[ -z "$out" || "$out" = "OFFLINE" ]]; then
            log "INFO" "Starting group replication on (${host})..."

            ${mysql} -N -e "STOP GROUP_REPLICATION;" # 2>/dev/null
#            echo "
#            ++++++++++ tmp: $tmp
#            ++++++++++ is_new[$tmp]: ${is_new[$tmp]}
#            "
            if [[ "${is_new[$tmp]}" -eq "1" ]]; then
                log "INFO" "RESET MASTER in host $host..."
                ${mysql} -N -e "RESET MASTER;" # 2>/dev/null
            fi
#            ${mysql} -N -e "RESET MASTER;" # 2>/dev/null
            ${mysql} -N -e "START GROUP_REPLICATION;" # 2>/dev/null
        else
            log "INFO" "Member state is unknown on host '${host}'..."
        fi
    fi
    ((tmp++))
done

while true; do
    echo -n .
    sleep 1
done

#my="mysql -u root --password=uWuj7-dbvefZVnJx"
#h0=my-galera-0.kubedb-gvr.demo.svc.cluster.local
#h1=my-galera-1.kubedb-gvr.demo.svc.cluster.local
#h2=my-galera-2.kubedb-gvr.demo.svc.cluster.local
#h3=my-galera-3.kubedb-gvr.demo.svc.cluster.local
#$my --host=$h0 -e 'SELECT * FROM performance_schema.replication_group_members;'
#$my --host=$h1 -e 'SELECT * FROM performance_schema.replication_group_members;'
#$my --host=$h2 -e 'SELECT * FROM performance_schema.replication_group_members;'
#$my --host=$h0 -e 'CREATE DATABASE playground; CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id)); INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");'
#$my --host=$h0 -e 'SELECT * FROM playground.equipment;'
#$my --host=$h1 -e 'SELECT * FROM playground.equipment;'
#$my --host=$h2 -e 'SELECT * FROM playground.equipment;'


# =========================================