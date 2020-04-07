#!/bin/bash

function startMicroservice {
    checkStatus
    local status=$?
    if (( $status > 0 )); then
        echo "The docker composer is already running"
    else
        echo "Docker composer starting..."
        sudo docker-compose up
    fi
}

function stopMicroservice {
    echo "Docker composer stopping..."
    sudo docker-compose kill
}

function statusMicroservice() {
    checkStatus
    local status=$?
    if (( $status > 0 )); then
        echo "RUNNING"
    else
        echo "NOT RUNNING"
    fi
}

function checkStatus() {
    countEtcd=$(sudo docker ps | grep -c etcd_goApp)
    countApp=$(sudo docker ps | grep -c goApp)

    local _status=$(($countEtcd > 0 && $countApp > 0))
    return $_status
}

"$@"