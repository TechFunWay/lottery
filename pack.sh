#!/bin/bash

# 全自动打包前端、Go、飞牛 NAS

scripts/frontend-build.sh
scripts/cross-platform-compile.sh
scripts/package-multiplatform.sh
