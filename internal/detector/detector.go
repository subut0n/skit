package detector

import "os"

// PackageManager represents a Node.js package manager.
type PackageManager int

const (
	NPM PackageManager = iota
	Yarn
	PNPM
	Bun
)

// Info holds the detected package manager and its run command.
type Info struct {
	Manager PackageManager
	Name    string // "npm", "yarn", "pnpm", "bun"
	RunCmd  string // "npm run", "yarn run", "pnpm run", "bun run"
}

// Detect examines a directory for lockfiles and returns the appropriate package manager.
// Priority: bun > pnpm > yarn > npm.
func Detect(dir string) Info {
	// Bun
	if fileExists(dir+"/bun.lockb") || fileExists(dir+"/bun.lock") {
		return Info{Manager: Bun, Name: "bun", RunCmd: "bun run"}
	}

	// PNPM
	if fileExists(dir + "/pnpm-lock.yaml") {
		return Info{Manager: PNPM, Name: "pnpm", RunCmd: "pnpm run"}
	}

	// Yarn
	if fileExists(dir + "/yarn.lock") {
		return Info{Manager: Yarn, Name: "yarn", RunCmd: "yarn run"}
	}

	// npm
	if fileExists(dir + "/package-lock.json") {
		return Info{Manager: NPM, Name: "npm", RunCmd: "npm run"}
	}

	// Default: npm
	return Info{Manager: NPM, Name: "npm", RunCmd: "npm run"}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
