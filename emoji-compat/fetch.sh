#!/usr/bin/env bash

# Copyright (C) 2022 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Helper script for fetching new emoji compat fonts from github source
# of truth

# This script is very basic, please extend or replace to handle your
# needs (e.g. pulling specific commits, releases, branches) as needed.

#set -o xtrace
set +e

METADATA_GIT="https://github.com/googlefonts/emojicompat.git"
FONT_GIT="https://github.com/googlefonts/noto-emoji.git"

SCRIPT_DIR=$(readlink -f $(dirname -- "$0"))
TMP_DIR=$(mktemp -d)

GIT_VERSION=$(git --version)
if [ $? -ne 0 ]; then
   echo -e "ERROR: git not found"
   exit 1
fi

TTX_VERSION=$(ttx --version)

if [ $? -ne 0 ]; then
   echo "ERROR ttx required to check font"
   echo -e "\t python3 -m venv venv"
   echo -e "\t source venv/bin/activate"
   echo -e "\t pip install fonttools"
   exit 127
fi

echo "METADATA:    $METADATA_GIT"
echo "FONT:        $FONT_GIT"
echo "Updating in: $SCRIPT_DIR"

# confirm directory is clean
pushd $SCRIPT_DIR > /dev/null
UNCOMMITED_CHANGES=$(git status --porcelain)
popd > /dev/null
if [[ "$UNCOMMITED_CHANGES" ]]; then
   echo "$UNCOMMITED_CHANGES"
   read -p "Uncommited changes. Continue? [y/N]:" uncommited
   if [[ ! $uncommited =~ ^[Yy] ]]; then
      exit 3
   fi
fi

function confirm_git_commit() {
   pushd $TMP_DIR/$1 > /dev/null
   RESULT=$(git log -1)
   echo "$RESULT"
   read -p "Continue for repo $1? [y/N]: " yn
   if [[ ! $yn =~ ^[Yy] ]]; then
      exit 2
   fi
   popd > /dev/null
}

pushd $TMP_DIR > /dev/null

git clone --quiet --depth 1 --branch main $METADATA_GIT
confirm_git_commit "emojicompat"
METADATA_FILE="./emojicompat/src/emojicompat/emoji_metadata.txt"
# adjust newlines to avoid giant diffs
cat $METADATA_FILE | awk 'sub("$", "\r")' > emoji_metadata.txt

# pull the font
git clone --quiet --depth 1 --branch main $FONT_GIT
confirm_git_commit "noto-emoji"
cp ./noto-emoji/fonts/NotoColorEmoji-emojicompat.ttf ./NewFont.ttf

ttx -o NewFont.ttx NewFont.ttf 2> /dev/null
grep -q 'header version="2.0"' NewFont.ttx

if [ $? -ne 0 ]; then
   echo -e "WRONG HEADER VERSION IN FONT FILE (breaks API23)"
   echo -e "Expected 'header version=\"2.0\""
   echo -e "Found: "
   grep 'header version' NewFont.ttx
   exit 128
fi

# concat new codepoints to emojis.txt
NEW_LINES=$(comm -23 emoji_metadata.txt $SCRIPT_DIR/data/emoji_metadata.txt)
NEW_CODEPOINTS=$(echo "$NEW_LINES" | cut -d" " -f4-100 | sed 's/\r//')

if [[ "$NEW_CODEPOINTS" ]]; then
    echo "$NEW_CODEPOINTS"
    read -p "New codpoints found in metadata. Append emojis.txt? [y/N]:" emojiAppend
    if [[ "$emojiAppend" =~ ^[Yy] ]]; then
        echo "$NEW_CODEPOINTS" >> $SCRIPT_DIR/supported-emojis/emojis.txt
        echo "Updated ${SCRIPT_DIR}/supported-emojis/emojis.txt"
    fi
fi

cp emoji_metadata.txt $SCRIPT_DIR/data/emoji_metadata.txt
echo "Updated ${SCRIPT_DIR}/data/emoji_metadata.txt"
cp NewFont.ttf $SCRIPT_DIR/font/NotoColorEmojiCompat.ttf
echo "Updated ${SCRIPT_DIR}/font/NotoColorEmojiCompat.ttf"

popd > /dev/null
rm -rf $TMP_DIR
