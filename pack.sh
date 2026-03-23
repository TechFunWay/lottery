#!/bin/bash

# 全自动打包前端、go、飞牛nas，直接调用他们的skill

.codebuddy/skills/frontend-build/scripts/build.sh
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
.codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh