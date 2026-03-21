---
name: frontend-build
description: This skill should be used when building the Vue 3 frontend for the Lottery Assistant project. Use it to compile, package, and prepare the frontend for production deployment or integration with the Go backend.
---

# Frontend Build

## Overview

This skill enables building the Vue 3 frontend for the Lottery Assistant project. It automates the compilation process, handles dependencies, generates production-optimized assets, and prepares the frontend for deployment or embedding into the Go backend.

## When to Use

Use this skill when:
- Building the frontend for production deployment
- Preparing frontend assets for Go embed integration
- Need to create optimized production builds
- Deploying the frontend to web servers
- Integrating frontend with backend for standalone binary distribution

## Quick Start

To build the frontend, run the build script:

```bash
.codebuddy/skills/frontend-build/scripts/build.sh
```

The script will:
1. Check and install dependencies if needed
2. Clean existing dist directory
3. Build the project using Vite
4. Display build statistics

## Build Process

### Step 1: Dependency Check

The script automatically checks if `node_modules` exists in the frontend directory. If not found, it runs `npm install` to install all required dependencies.

### Step 2: Clean Previous Build

Before building, the script removes the existing `dist` directory to ensure a clean build without any stale files.

### Step 3: Compile Project

The script runs `npm run build`, which executes:
- `vue-tsc -b` - Type checking with TypeScript compiler
- `vite build` - Production build with Vite bundler

### Step 4: Build Verification

After successful build, the script displays:
- Total output size
- File count
- List of main files (index.html, JS/CSS assets)

## Project Structure

The frontend project uses:
- **Framework**: Vue 3 with Composition API
- **Language**: TypeScript
- **Build Tool**: Vite 5
- **Styling**: Tailwind CSS
- **Router**: Vue Router 4
- **Charts**: ECharts with vue-echarts
- **Icons**: Lucide Vue Next

## Output Directory

Built files are output to `frontend/dist/`:
- `index.html` - Main HTML entry point
- `assets/` - Compiled JavaScript and CSS files
  - `*.js` - Optimized JavaScript bundles
  - `*.css` - Extracted CSS styles

## Build Scripts Available

In `frontend/package.json`:
- `npm run dev` - Start development server (default port 5176)
- `npm run build` - Build for production (used by this skill)
- `npm run preview` - Preview production build locally

## Integration with Go Backend

After building, the `dist/` directory can be:
1. Deployed directly to a web server
2. Embedded into Go binary using Go embed package
3. Used with the `go-cross-platform-build` skill for multi-platform distribution

## Troubleshooting

### Build Fails

If the build fails, check:
1. Node.js and npm are installed correctly
2. All dependencies are installed (run `npm install` manually)
3. TypeScript errors in source files
4. Network connectivity for downloading dependencies

### Type Errors

If TypeScript compilation fails:
1. Check `frontend/src/types/index.ts` for type definitions
2. Ensure all API imports match backend endpoints
3. Verify component props and emits are properly typed

### Missing Dependencies

If dependencies are missing:
1. Delete `node_modules` and `package-lock.json`
2. Run `npm install` to reinstall all dependencies

## Resources

### scripts/build.sh

Automated build script that:
- Validates project environment
- Installs dependencies automatically
- Cleans previous builds
- Runs production build with error handling
- Displays build statistics
- Provides colored output for easy reading

Execute directly from project root or skill directory.
