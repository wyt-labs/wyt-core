#! /bin/bash
set -e

base_dir=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)
export WYT_CORE_PATH=${base_dir}

${base_dir}/tools/control.sh status