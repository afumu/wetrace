#!/bin/bash

set -eu
set -o pipefail
[ "$#" = "1" ] && [ "$1" = '-v' ] && set -x

OUTPUT_DIR="bin"
PACKAGES_DIR="packages"
TEMP_DIR="temp_package"
VERSION=$(git describe --tags --always --dirty="-dev")
CHECKSUMS_FILE="$PACKAGES_DIR/checksums.txt"

make -f Makefile crossbuild

rm -rf $PACKAGES_DIR $TEMP_DIR

mkdir -p $PACKAGES_DIR $TEMP_DIR

echo "" > $CHECKSUMS_FILE

for binary in $OUTPUT_DIR/chatlog_*_*; do
    binary_name=$(basename $binary)

    # quick start
    if [[ $binary_name == "chatlog_darwin_amd64" ]]; then
        cp "$binary" "$PACKAGES_DIR/chatlog_macos"
        echo "$(sha256sum $PACKAGES_DIR/chatlog_macos | sed "s|$PACKAGES_DIR/||")" >> $CHECKSUMS_FILE
    elif [[ $binary_name == "chatlog_windows_amd64" ]]; then
        cp "$binary" "$PACKAGES_DIR/chatlog_windows.exe"
        echo "$(sha256sum $PACKAGES_DIR/chatlog_windows.exe | sed "s|$PACKAGES_DIR/||")" >> $CHECKSUMS_FILE
    elif [[ $binary_name == "chatlog_linux_amd64" ]]; then
        cp "$binary" "$PACKAGES_DIR/chatlog_linux"
        echo "$(sha256sum $PACKAGES_DIR/chatlog_linux | sed "s|$PACKAGES_DIR/||")" >> $CHECKSUMS_FILE
    fi

    cp "README.md" "LICENSE" $TEMP_DIR

    package_name=""
    os_arch=$(echo $binary_name | cut -d'_' -f 2-)
    if [[ $binary_name == *"_windows_"* ]]; then
        cp "$binary" "$TEMP_DIR/chatlog.exe"
        package_name="chatlog_${VERSION}_${os_arch}.zip"
        zip -j "$PACKAGES_DIR/$package_name" -r $TEMP_DIR/*
    else
        cp "$binary" "$TEMP_DIR/chatlog"
        package_name="chatlog_${VERSION}_${os_arch}.tar.gz"
        tar -czf "$PACKAGES_DIR/$package_name" -C $TEMP_DIR .
    fi

    rm -rf $TEMP_DIR/*

    if [[ ! -z "$package_name" ]]; then
        echo "$(sha256sum $PACKAGES_DIR/$package_name | sed "s|$PACKAGES_DIR/||")" >> $CHECKSUMS_FILE
    fi

done

rm -rf $TEMP_DIR

echo "📦 All packages and their sha256 checksums have been created in $PACKAGES_DIR/"