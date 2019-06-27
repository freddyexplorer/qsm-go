#! /bin/bash

QSM_ENV_NUMBER=${QSM_ENV_NUMBER:=1}
dbUser="not-a-user"
dbPassword=""
dbName="not-a-database"
dbLoc="was-not-set"
dbConfFile="was-not-set"

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
. $curDir/functions.sh

dbError=13

ensureUser() {
    echo "INFO: Checking user $dbUser"
    checkUser=$(psql -qAt postgres -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser already exists"
    else
        echo "INFO: Creating user $dbUser"
        psql -qAt postgres -c "create user $dbUser with encrypted password '$dbPassword';"
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to create user $dbUser"
            exit $dbError
        fi
        echo "INFO: User $dbUser created"
    fi
}

ensureDb() {
    echo "INFO: Checking db $dbName"
    checkDb=$(psql -qAt postgres -c "SELECT 1 FROM pg_database WHERE datname='$dbName';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName already exists"
    else
        echo "INFO: Creating database $dbName"
        psql -qAt postgres <<EOF
create database $dbName;
grant all privileges on database $dbName to $dbUser;
EOF
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to create database $dbName"
            exit $dbError
        fi
        echo "INFO: Database $dbName created"
    fi
}

dropUser() {
    echo "INFO: Dropping user $dbUser"
    checkUser=$(psql -qAt postgres -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser exists => deleting it"
        psql -qAt postgres -c "drop user $dbUser;"
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to drop user $dbUser"
            exit $dbError
        fi
        echo "INFO: User $dbUser deleted"
    else
        echo "INFO: User $dbUser already deleted"
    fi
}

dropDb() {
    echo "INFO: Dropping db $dbName"
    checkDb=$(psql -qAt postgres -c "SELECT 1 FROM pg_database WHERE datname='$dbName';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName exists => Dropping it"
        psql -qAt postgres <<EOF
drop database $dbName;
EOF
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to drop database $dbName"
            exit $dbError
        fi
        echo "INFO: Database $dbName dropped"
    else
        echo "INFO: Database $dbName already dropped"
    fi
}

case "$1" in
	check)
	    checkDbConf || exit $?
        echo "INFO: Checking all good for QSM dev using:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		ensureRunningPg && ensureUser && ensureDb
		exit $?
		;;
	drop)
	    checkDbConf || exit $?
        echo "INFO: Dropping QSM environment:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
	    ensureRunningPg && dropDb && dropUser && rm $dbConfFile
		exit $?
		;;
	start)
		ensureRunningPg
		exit $?
		;;
	stop)
		pg_ctl -D $dbLoc stop
		exit $?
		;;
	conf)
	    checkDbConf || exit $?
        echo "INFO: Checked configuration for QSM environment:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		exit 0
		;;
	rmconf)
		rm $dbConfFile
		exit $?
		;;
	status)
	    checkDbConf || exit $?
        echo "INFO: Checking PostgreSQL using:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		pg_ctl -D $dbLoc status
		exit $?
		;;
	*)
		echo "ERROR: Command $1 unknown"
		exit 1
esac

