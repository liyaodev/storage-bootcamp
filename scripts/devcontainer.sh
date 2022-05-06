#!/bin/bash

if [ "${1-}" = "up" ]; then
    mkdir -p "${STORAGE_BOOTCAMP_ROOT_DIR:-.}/volumes/vscode-extensions"
    chmod -R 777 "${STORAGE_BOOTCAMP_ROOT_DIR:-.}/volumes"

    docker-compose -f ${STORAGE_BOOTCAMP_ROOT_DIR:-.}/docker-compose-devcontainer.yml up -d
fi

if [ "${1-}" = "down" ]; then
    docker-compose -f ${STORAGE_BOOTCAMP_ROOT_DIR:-.}/docker-compose-devcontainer.yml down
    rm -rf "${STORAGE_BOOTCAMP_ROOT_DIR:-.}/volumes"
fi
