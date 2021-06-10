# MIT License
#
# Copyright (c) 2020 Shyam Jesalpura and EASE lab
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

#!/bin/bash

if [ -z $1 ] || [ -z $2 ] || [ -z $3 ] || [ -z $4 ]; then
    echo "Parameters missing"
    echo "USAGE: start_runners.sh <num of runners> https://github.com/<OWNER>/<REPO> <Github Access key> xdt [restart]"
    exit -1
fi

NUM_OF_RUNNERS=$1
RUNNER_LABEL=$4
RESTART_FLAG=$5

# fetch runner token using access token
ACCESS_TOKEN=$3
REPO_URL=$2

# pull latest image
docker pull myoung34/github-runner:ubuntu-bionic


case "$RUNNER_LABEL" in
"xdt")

    if [ "$RESTART_FLAG" == "restart" ]; then
        docker container stop $(docker ps --format "{{.Names}}" | grep xdt-github_runner)
        docker container rm $(docker ps -a --format "{{.Names}}" | grep xdt-github_runner)
    fi
    for number in $(seq 1 $NUM_OF_RUNNERS)
    do
        # create access token as mentioned here (https://github.com/myoung34/docker-github-actions-runner#create-github-personal-access-token)
        docker run -d --restart always \
            --name "xdt-github_runner-${HOSTNAME}-${number}" \
            -e REPO_URL="${REPO_URL}" \
            -e RUNNER_NAME="${HOSTNAME}-${number}" \
            -e ACCESS_TOKEN="${ACCESS_TOKEN}" \
            -e RUNNER_WORKDIR="/tmp/github-XDTrunner" \
            -v /var/run/docker.sock:/var/run/docker.sock \
            myoung34/github-runner:ubuntu-bionic
    done
    ;;
*)
    echo "Invalid label"
    ;;
esac
