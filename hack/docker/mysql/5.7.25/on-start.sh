#!/usr/bin/env bash

#set -eoux pipefail

# Environment variables passed from Pod env are as follows:
#
#   GROUP_NAME          = a uuid treated as the name of the replication group
#   BASE_NAME           = name of the StatefulSet (same as the name of CRD)
#   BASE_SERVER_ID      = server-id of the primary member
#   GOV_SVC             = the name of the governing service
#   POD_NAMESPACE       = the Pods' namespace
#   MYSQL_ROOT_USERNAME = root user name
#   MYSQL_ROOT_PASSWORD = root password

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

# get_host_name() expects only one argument and that is the index of the Pod of StatefulSet.
# And it forms the FQDN (Fully Qualified Domain Name) of the $1'th Pod of StatefulSet.
function get_host_name() {
  echo -n "$BASE_NAME-$1.$GOV_SVC.$NAMESPACE.svc.cluster.local"
}

# get the host names from stdin sent by peer-finder program
cur_hostname=$(hostname)
export cur_host=
log "INFO" "Reading standard input..."
while read -ra line; do
  if [[ "${line}" == *"${cur_hostname}"* ]]; then
    cur_host="$line"
    log "INFO" "I am $cur_host"
  fi
  peers=("${peers[@]}" "$line")
done
log "INFO" "Trying to start group with peers'${peers[*]}'"

# store the value for the variables those will be written in /etc/mysql/my.cnf file

# comma separated host names
export hosts=$(echo -n ${peers[*]} | sed -e "s/ /,/g")

# comma separated seed addresses of the hosts (host1:port1,host2:port2,...)
export seeds=$(echo -n ${hosts} | sed -e "s/,/:33060,/g" && echo -n ":33060")

# server-id calculated as $BASE_SERVER_ID + pod index
declare -i srv_id=$(hostname | sed -e "s/${BASE_NAME}-//g")
((srv_id += $BASE_SERVER_ID))

export cur_addr="${cur_host}:33060"

# the mysqld configurations have take by following
# 01. official doc: https://dev.mysql.com/doc/refman/5.7/en/group-replication-configuring-instances.html
# 02. digitalocean doc: https://www.digitalocean.com/community/tutorials/how-to-configure-mysql-group-replication-on-ubuntu-16-04
log "INFO" "Storing default mysqld config into /etc/mysql/my.cnf"
cat >>/etc/mysql/my.cnf <<EOL

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

log "INFO" "Starting mysql server with 'docker-entrypoint.sh mysqld $@'..."

# ensure the mysqld process be stopped
/etc/init.d/mysql stop

# run the mysqld process in background with user provided arguments if any
docker-entrypoint.sh mysqld $@ &
pid=$!
log "INFO" "The process id of mysqld is '$pid'"
#sleep 5

# wait for all mysql servers be running (alive)
for host in ${peers[*]}; do
  for i in {900..0}; do
    out=$(mysqladmin -u ${USER} --password=${PASSWORD} --host=${host} ping 2>/dev/null)
    if [[ "$out" == "mysqld is alive" ]]; then
      #      sleep 5
      break
    fi

    echo -n .
    sleep 1
  done

  if [[ "$i" == "0" ]]; then
    echo ""
    log "ERROR" "Server ${host} start failed..."
    exit 1
  fi
done

log "INFO" "All servers (${peers[*]}) are ready"

# now we need to configure a replication user for each server.
# the procedures for this have been taken by following
# 01. official doc (section from 17.2.1.3 to 17.2.1.5): https://dev.mysql.com/doc/refman/5.7/en/group-replication-user-credentials.html
# 02. digitalocean doc: https://www.digitalocean.com/community/tutorials/how-to-configure-mysql-group-replication-on-ubuntu-16-04
#####################################################################
# Begin initialization process                                      #
#####################################################################
export mysql_header="mysql -u ${USER}"

# this is to bypass the warning message for using password
export MYSQL_PWD=${PASSWORD}
export member_hosts=($(echo -n ${hosts} | sed -e "s/,/ /g"))

for host in ${member_hosts[*]}; do
  log "INFO" "Initializing the server (${host})..."

  mysql="$mysql_header --host=$host"

  out=$(${mysql} -N -e "select count(host) from mysql.user where mysql.user.user='repl';" | awk '{print$1}')
  if [[ "$out" -eq "0" ]]; then

    # is_new is an array,
    #                              | 1; if i'th host is created for the 1st time
    #       where ${is_new[$i]} =  |
    #                              | 0; otherwise (may rebooted)
    # So, for the first time creation a '1' otherwise a '0' will be appended
    is_new=("${is_new[@]}" "1")

    log "INFO" "Replication user not found and creating one..."
    ${mysql} -N -e "SET SQL_LOG_BIN=0;"
    ${mysql} -N -e "CREATE USER 'repl'@'%' IDENTIFIED BY 'password' REQUIRE SSL;"
    ${mysql} -N -e "GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';"
    ${mysql} -N -e "FLUSH PRIVILEGES;"
    ${mysql} -N -e "SET SQL_LOG_BIN=1;"

    ${mysql} -N -e "CHANGE MASTER TO MASTER_USER='repl', MASTER_PASSWORD='password' FOR CHANNEL 'group_replication_recovery';"
  else
    log "INFO" "Replication user info exists"
    is_new=("${is_new[@]}" "0")
  fi

  # ensure the group replication plugin be installed
  out=$(${mysql} -N -e 'SHOW PLUGINS;' | grep group_replication)
  if [[ -z "$out" ]]; then
    log "INFO" "Installing group replication plugin..."
    ${mysql} -e "INSTALL PLUGIN group_replication SONAME 'group_replication.so';"
  else
    log "INFO" "Already group replication plugin is installed"
  fi
done
#####################################################################
# End initialization process                                        #
#####################################################################

function find_group() {
  # TODO: Need to handle for multiple group existence
  group_found=0
  for host in $@; do

    export mysql="$mysql_header --host=${host}"
    # value may be 'UNDEFINED'
    primary_id=$(${mysql} -N -e "SHOW STATUS WHERE Variable_name = 'group_replication_primary_member';" | awk '{print $2}')
    if [[ -n "$primary_id" ]]; then
      ids=($(${mysql} -N -e "SELECT MEMBER_ID FROM performance_schema.replication_group_members WHERE MEMBER_STATE = 'ONLINE' OR MEMBER_STATE = 'RECOVERING';"))

      for id in ${ids[@]}; do
        if [[ "${primary_id}" == "${id}" ]]; then
          group_found=1
          primary_host=$(${mysql} -N -e "SELECT MEMBER_HOST FROM performance_schema.replication_group_members WHERE MEMBER_ID = '${primary_id}';" | awk '{print $1}')

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

export primary_host=$(get_host_name 0)
export found=$(find_group ${member_hosts[*]})

# filter the Pod index from the variable $primary_host
primary_idx=$(echo ${primary_host} | sed -e "s/.${GOV_SVC}.${NAMESPACE}.svc.cluster.local//g" | sed -e "s/${BASE_NAME}-//g")

# First start group replication inside the primary
#####################################################################
# Begin group replication bootstrap process if no group exists      #
#####################################################################
log "INFO" "Checking whether there exists any replication group or not..."
if [[ "$found" == "0" ]]; then
  mysql="$mysql_header --host=$primary_host"

  # get the member state from performance_schema.replication_group_members
  out=$(${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_HOST = '$primary_host';")
  if [[ -z "$out" || "$out" == "OFFLINE" ]]; then
    log "INFO" "No group is found and bootstrapping one on host '$primary_host'..."

    ${mysql} -N -e "STOP GROUP_REPLICATION;"

    # reset is needed for the first time creation
    if [[ "${is_new[$primary_idx]}" -eq "1" ]]; then
      log "INFO" "RESET MASTER in primary host $primary_host..."
      ${mysql} -N -e "RESET MASTER;"
    fi

    ${mysql} -N -e "SET GLOBAL group_replication_bootstrap_group=ON;"
    ${mysql} -N -e "START GROUP_REPLICATION;"
    ${mysql} -N -e "SET GLOBAL group_replication_bootstrap_group=OFF;"

    log "INFO" "A new group (name $GROUP_NAME) is bootstrapped on $primary_host"

  else
    log "INFO" "No group is found and member state is '$out' on host '$primary_host'..."
  fi

else
  log "INFO" "A group is found and the primary host is '$primary_host'..."
fi
#####################################################################
# End bootstrap process                                             #
#####################################################################

# Now start group replication inside the others
#####################################################################
# Begin joining others members to the group                         #
#####################################################################
declare -i host_idx=0

for host in ${member_hosts[*]}; do

  if [[ "$host" != "$primary_host" ]]; then
    mysql="$mysql_header --host=$host"

    # get the member state from performance_schema.replication_group_members
    out=$(${mysql} -N -e "SELECT MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_HOST = '$host';")

    if [[ -z "$out" || "$out" == "OFFLINE" ]]; then
      log "INFO" "Starting group replication on (${host})..."

      ${mysql} -N -e "STOP GROUP_REPLICATION;"

      # reset is needed for the first time creation
      if [[ "${is_new[$host_idx]}" -eq "1" ]]; then
        log "INFO" "RESET MASTER in host $host..."
        ${mysql} -N -e "RESET MASTER;"
      fi

      ${mysql} -N -e "START GROUP_REPLICATION;"

      log "INFO" "$host is joined the group $GROUP_NAME"

    else
      log "INFO" "Member state is '${out}' on host '${host}'..."
    fi
  fi

  ((host_idx++))
done
#####################################################################
# End joining process                                               #
#####################################################################

# wait for mysqld process running in background
log "INFO" "Waiting for mysqld server process running..."
wait $pid
