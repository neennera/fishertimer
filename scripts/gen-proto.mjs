// Regenerates the Go code for every .proto file under proto/.
//
// Usage (from the repo root):  pnpm proto:gen
//
// Requires protoc plus the two Go plugins:
//   protoc              brew install protobuf   (or https://grpc.io/docs/protoc-installation/)
//   protoc-gen-go       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//   protoc-gen-go-grpc  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//
// The generated *.pb.go files are committed, so only whoever changes a .proto
// file needs these tools installed.
import { execFileSync } from 'child_process';
import fs from 'fs';
import path from 'path';

const rootDir = process.cwd();
const protoDir = path.join(rootDir, 'proto');

// `go install` puts the plugins in $GOPATH/bin, which is often not on PATH.
function withGoBinOnPath() {
  const env = { ...process.env };
  try {
    const goPath = execFileSync('go', ['env', 'GOPATH'], { encoding: 'utf8' }).trim();
    env.PATH = [path.join(goPath, 'bin'), env.PATH].join(path.delimiter);
  } catch {
    // Go is not installed; protoc will report the missing plugins below.
  }
  return env;
}

function findProtoFiles(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) return findProtoFiles(fullPath);
    return entry.name.endsWith('.proto') ? [fullPath] : [];
  });
}

const protoFiles = findProtoFiles(protoDir).map((file) => path.relative(protoDir, file));

if (protoFiles.length === 0) {
  console.error('No .proto files found under proto/');
  process.exit(1);
}

try {
  execFileSync(
    'protoc',
    [
      `--proto_path=${protoDir}`,
      `--go_out=${protoDir}`,
      '--go_opt=paths=source_relative',
      `--go-grpc_out=${protoDir}`,
      '--go-grpc_opt=paths=source_relative',
      ...protoFiles,
    ],
    { stdio: 'inherit', env: withGoBinOnPath() }
  );
} catch {
  console.error('\nprotoc failed. Check that protoc and both Go plugins are installed (see the top of this script).');
  process.exit(1);
}

for (const file of protoFiles) {
  console.log(`generated  proto/${file}`);
}
