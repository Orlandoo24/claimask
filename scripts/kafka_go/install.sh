#!/bin/bash

# HOST_CONFIG=$(wget -O - "https://getip.clougence.com/" --no-check-certificate 2>/dev/null)
HOST_CONFIG=`ifconfig en0 | awk '/inet /{print $2}' | head -n 1`
echo HOST_CONFIG=${HOST_CONFIG} > .env

function install_before() {
    echo "Begin to create network and volume..." 
    docker network create mx-wk
    echo "" 
    echo -e "Create network \033[32mSUCCESS.\033[0m"
    echo ""
}

function install_kafka() {
    docker compose up -d kafka-0 kafka-1 kafka-2
    docker compose up -d kafkaui
    echo "" 
    echo -e "Kafka Web UI: \033[32m"${HOST_CONFIG}":7080\033[0m"
    echo -e "Install kafka cluster \033[32mSUCCESS.\033[0m"       
    echo ""
}


function __main() {
    install_before

    install_kafka

}


 __main
