#!/bin/bash

# 全自动打包前端、go、飞牛nas，直接调用他们的skill

.skill/frontend-build/scripts/build.sh
.skill/cross-platform-compile/scripts/compile.sh
.skill/fnnas-packager/scripts/package-multiplatform.sh