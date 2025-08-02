#!/bin/bash

# 启动脚本 - 等待数据库就绪后启动应用

set -e

echo "Starting Trojan Panel..."

# 等待MySQL数据库就绪
echo "Waiting for MySQL database to be ready..."
until mysql -h"$host" -P"$port" -u"$user" -p"$password" -e "SELECT 1;" &> /dev/null; do
    echo "MySQL is unavailable - sleeping"
    sleep 2
done

echo "MySQL is up - continuing..."

# 等待Redis就绪
echo "Waiting for Redis to be ready..."
until redis-cli -h "$redisHost" -p "$redisPort" -a "$redisPassword" ping &> /dev/null; do
    echo "Redis is unavailable - sleeping"
    sleep 2
done

echo "Redis is up - continuing..."

# 启动应用
echo "Starting Trojan Panel application..."
exec ./trojan-panel \
    -host="$mariadb_ip" \
    -port="$mariadb_port" \
    -user="$mariadb_user" \
    -password="$mariadb_pas" \
    -redisHost="$redis_host" \
    -redisPort="$redis_port" \
    -redisPassword="$redis_pass" \
    -serverPort="$server_port"
