import fs from 'fs';
import path from 'path';

const rootDir = process.cwd();
console.log(`Setting up Turborepo in: ${rootDir}`);

// 1. Copy files from temp-turbo to root
const tempTurboDir = path.join(rootDir, 'temp-turbo');

if (fs.existsSync(tempTurboDir)) {
  console.log('Copying base template from temp-turbo...');
  
  // Copy root files
  const rootFiles = ['.gitignore', '.npmrc', 'package.json', 'turbo.json', 'pnpm-workspace.yaml', 'pnpm-lock.yaml'];
  for (const file of rootFiles) {
    const src = path.join(tempTurboDir, file);
    const dest = path.join(rootDir, file);
    if (fs.existsSync(src)) {
      fs.copyFileSync(src, dest);
    }
  }

  // Copy packages
  const pkgsSrc = path.join(tempTurboDir, 'packages');
  const pkgsDest = path.join(rootDir, 'packages');
  if (fs.existsSync(pkgsSrc)) {
    fs.cpSync(pkgsSrc, pkgsDest, { recursive: true });
  }

  // Copy apps/web only (apps/docs is not needed as docs are in /docs/phase1)
  const webSrc = path.join(tempTurboDir, 'apps', 'web');
  const webDest = path.join(rootDir, 'apps', 'web');
  if (fs.existsSync(webSrc)) {
    fs.cpSync(webSrc, webDest, { recursive: true });
  }

  // Remove temp-turbo
  fs.rmSync(tempTurboDir, { recursive: true, force: true });
  console.log('Cleaned up temp-turbo.');
}

// 2. Configure pnpm-workspace.yaml
const workspaceYaml = `packages:
  - "apps/*"
  - "services/*"
  - "packages/*"
`;
fs.writeFileSync(path.join(rootDir, 'pnpm-workspace.yaml'), workspaceYaml);

// 3. Update root package.json
const rootPackageJsonPath = path.join(rootDir, 'package.json');
const rootPkg = JSON.parse(fs.readFileSync(rootPackageJsonPath, 'utf8'));
rootPkg.name = "fishertimer";
rootPkg.description = "Fisher Timer - Community Study Timer & Gamified Focus Monorepo";
rootPkg.scripts = {
  "build": "turbo build",
  "dev": "turbo dev",
  "lint": "turbo lint",
  "test": "turbo test",
  "check-types": "turbo check-types",
  "clean": "turbo clean"
};
fs.writeFileSync(rootPackageJsonPath, JSON.stringify(rootPkg, null, 2));

// 4. Update turbo.json
const turboConfig = {
  "$schema": "https://turbo.build/schema.json",
  "ui": "tui",
  "globalPassThroughEnv": [
    "LocalAppData",
    "LOCALAPPDATA",
    "GOPATH",
    "GOROOT",
    "GOCACHE",
    "PATH",
    "USERPROFILE",
    "SystemRoot",
    "HOME"
  ],
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$", ".env*"],
      "outputs": [".next/**", "!.next/cache/**", "dist/**", "bin/**"]
    },
    "lint": {
      "dependsOn": ["^lint"]
    },
    "check-types": {
      "dependsOn": ["^check-types"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$"]
    },
    "clean": {
      "cache": false
    }
  }
};
fs.writeFileSync(path.join(rootDir, 'turbo.json'), JSON.stringify(turboConfig, null, 2));

// 5. Define services from docs/phase1/microservice.md
const services = [
  { name: 'account', port: 8082, desc: 'Google OAuth authentication, user profile management, and statistics dashboard' },
  { name: 'study-session', port: 8083, desc: 'Study session room lifecycle, participant rosters, and room capacity' },
  { name: 'study-timer', port: 8084, desc: 'Independent user timer execution, focus cycles, and rest intervals' },
  { name: 'reward', port: 8085, desc: 'Gamification reward drop calculation and inventory progression' },
  { name: 'leaderboard', port: 8086, desc: 'Read-optimized rankings filtered by timeframe' },
  { name: 'admin', port: 8087, desc: 'Real-time room monitoring, user reporting, and moderation' }
];

const servicesDir = path.join(rootDir, 'services');
if (!fs.existsSync(servicesDir)) {
  fs.mkdirSync(servicesDir, { recursive: true });
}

for (const svc of services) {
  const svcDir = path.join(servicesDir, svc.name);
  const cmdDir = path.join(svcDir, 'cmd');
  fs.mkdirSync(cmdDir, { recursive: true });

  // package.json for Turborepo integration
  const svcPackageJson = {
    "name": `@fishertimer/${svc.name}-service`,
    "version": "0.1.0",
    "private": true,
    "description": svc.desc,
    "scripts": {
      "dev": "go run cmd/main.go",
      "build": "go build -o bin/server cmd/main.go",
      "test": "go test ./...",
      "lint": "go vet ./...",
      "clean": "go clean"
    }
  };
  fs.writeFileSync(path.join(svcDir, 'package.json'), JSON.stringify(svcPackageJson, null, 2));

  // go.mod
  const goModContent = `module github.com/neennera/fishertimer/services/${svc.name}

go 1.22
`;
  fs.writeFileSync(path.join(svcDir, 'go.mod'), goModContent);

  // cmd/main.go scaffold (minimal health-check server, no business logic code)
  const mainGoContent = `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HealthResponse struct {
	Service   string \`json:"service"\`
	Status    string \`json:"status"\`
	Port      int    \`json:"port"\`
	Timestamp string \`json:"timestamp"\`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := HealthResponse{
		Service:   "${svc.name}",
		Status:    "healthy",
		Port:      ${svc.port},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := ${svc.port}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("${svc.name} service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down ${svc.name} service...")
}
`;
  fs.writeFileSync(path.join(cmdDir, 'main.go'), mainGoContent);

  // Service README.md
  const svcReadme = `# ${svc.name.charAt(0).toUpperCase() + svc.name.slice(1)} Service

## Overview
${svc.desc}

## Port
Default port: \`${svc.port}\`

## Scripts
- \`pnpm dev\` : Runs the service locally with Go
- \`pnpm build\` : Compiles the Go binary to \`bin/server\`
- \`pnpm test\` : Runs Go tests
- \`pnpm lint\` : Runs Go static analysis (\`go vet\`)
`;
  fs.writeFileSync(path.join(svcDir, 'README.md'), svcReadme);

  // .gitignore for service
  const svcGitignore = `bin/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
`;
  fs.writeFileSync(path.join(svcDir, '.gitignore'), svcGitignore);

  console.log(`Created service scaffold: ${svc.name} on port ${svc.port}`);
}

// 6. Create root go.work for multi-module Go workspace
const goWorkContent = `go 1.22

use (
	./services/account
	./services/study-session
	./services/study-timer
	./services/reward
	./services/leaderboard
	./services/admin
)
`;
fs.writeFileSync(path.join(rootDir, 'go.work'), goWorkContent);
console.log('Created root go.work for Go workspace.');

// 7. Create shared-types package for contracts/DTOs between services & frontend
const sharedTypesDir = path.join(rootDir, 'packages', 'shared-types');
fs.mkdirSync(path.join(sharedTypesDir, 'src'), { recursive: true });

const sharedTypesPkg = {
  "name": "@fishertimer/shared-types",
  "version": "0.1.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "scripts": {
    "check-types": "tsc --noEmit"
  },
  "devDependencies": {
    "@repo/typescript-config": "workspace:*",
    "typescript": "7.0.2"
  }
};
fs.writeFileSync(path.join(sharedTypesDir, 'package.json'), JSON.stringify(sharedTypesPkg, null, 2));

const sharedTypesTsConfig = {
  "extends": "@fishertimer/typescript-config/base.json",
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
};
fs.writeFileSync(path.join(sharedTypesDir, 'tsconfig.json'), JSON.stringify(sharedTypesTsConfig, null, 2));

const sharedTypesIndex = `// Shared API Contracts and Domain Models for Fisher Timer

export interface User {
  id: string;
  email: string;
  displayName: string;
  isBanned: boolean;
  createdAt: string;
}

export interface StudySession {
  id: string;
  name: string;
  creatorId: string;
  participantLimit: number;
  status: 'ACTIVE' | 'ENDED';
  createdAt: string;
}

export interface SessionParticipant {
  sessionId: string;
  userId: string;
  joinedAt: string;
  leftAt?: string;
}

export type TimerStatus = 'STOPPED' | 'RUNNING' | 'PAUSED' | 'RESTING';

export interface TimerSession {
  sessionId: string;
  userId: string;
  status: TimerStatus;
  workDurationMinutes: number;
  restDurationMinutes: number;
  currentCycle: number;
}

export interface Reward {
  id: string;
  userId: string;
  name: string;
  rarity: 'COMMON' | 'UNCOMMON' | 'RARE' | 'EPIC' | 'LEGENDARY';
  awardedAt: string;
}

export interface LeaderboardEntry {
  userId: string;
  displayName: string;
  rank: number;
  totalRewardCount: number;
  totalFocusMinutes: number;
}
`;
fs.writeFileSync(path.join(sharedTypesDir, 'src', 'index.ts'), sharedTypesIndex);
console.log('Created @fishertimer/shared-types package.');

// 8. Update root .gitignore
const rootGitignore = `# Dependencies
node_modules/
.pnpm-store/

# Next.js & Turborepo
.next/
out/
.turbo/

# Go binaries
bin/
*.exe
*.test

# Environment & IDE
.env*
!.env.example
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db
`;
fs.writeFileSync(path.join(rootDir, '.gitignore'), rootGitignore);

console.log('Turborepo scaffold completed successfully!');
