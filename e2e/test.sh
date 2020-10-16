#!/usr/bin/env bash
set -e

for dir in */
do
    dir=${dir%*/}
    echo "  * Running ${dir}"
    (
        cd ${dir}
        if [[ -f test.sh ]]; then
            ./test.sh
        else
            ../../halfpipe 1> pipelineActual.yml
            diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
            if command -v fly > /dev/null; then
                fly validate-pipeline -c pipelineActual.yml > /dev/null
            fi
        fi
    )
done
