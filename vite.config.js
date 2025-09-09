import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path, { resolve } from "path";
import DynamicPublicDirectory from "vite-multiple-assets";
import fs from 'fs';
import { exec } from 'child_process';
import { promisify } from 'util';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);

// Resolve project root based on this config file's location (not process.cwd())
const __filename = fileURLToPath(import.meta.url);
const projectRoot = path.dirname(__filename);

// Custom plugin to build WASM
function wasmBuildPlugin(exampleName) {
    const buildWasm = async () => {
        const wasmPath = resolve(projectRoot, `examples/${exampleName}/main.wasm`);
        const goFile = resolve(projectRoot, `examples/${exampleName}/main.go`);
        
        try {
            console.log(`[wasm-build] Building WASM for ${exampleName}...`);
            const { stdout, stderr } = await execAsync(`GOOS=js GOARCH=wasm go build -o ${wasmPath} ${goFile}`);
            if (stderr) console.warn(`[wasm-build] Warning: ${stderr}`);
            console.log(`[wasm-build] WASM build complete for ${exampleName}`);
        } catch (error) {
            console.error(`[wasm-build] Build failed:`, error.message);
            throw error;
        }
    };

    return {
        name: 'wasm-build',
        buildStart() {
            // Build WASM on startup
            return buildWasm();
        },
        configureServer(server) {
            // Watch for Go file changes
            const chokidar = server.watcher;
            chokidar.add('**/*.go');
            chokidar.on('change', (path) => {
                if (path.endsWith('.go') && !path.includes('vendor') && !path.includes('.devenv')) {
                    console.log(`[wasm-build] Go file changed: ${path}`);
                    buildWasm().then(() => {
                        server.ws.send({ type: 'full-reload' });
                    });
                }
            });
        }
    };
}

// Function to parse CLI arguments
export function parseCliArgs() {
    const cmdArgs = new Set(process.argv.slice(6));
    return {
        prod: cmdArgs.has("prod"),
        dontRun: cmdArgs.has("dontRun"),
    };
}

// Parse CLI arguments
const { prod, dontRun } = parseCliArgs();

const baseCfg = (pubDirs, preCmds = [], exampleName = "counter") => {
    let buildArgs = [`--tags ${prod ? "prod" : "dev"} -ldflags "-w -s"`];
    return defineConfig({
        root: resolve(projectRoot, `examples/${exampleName}`),
        plugins: [
            react(),
            wasmBuildPlugin(exampleName),
            // Serve fixed wasm_exec.js and per-example main.wasm
            {
                name: 'wasm-asset-middleware',
                configureServer(server) {
                    const fixedWasmExecPath = resolve(projectRoot, 'examples/wasm_exec.js');
                    const exampleWasmPath = resolve(projectRoot, `examples/${exampleName}/main.wasm`);
                    server.middlewares.use(async (req, res, next) => {
                        try {
                            if (!req.url) return next();
                            if (req.method !== 'GET' && req.method !== 'HEAD') return next();
                            if (req.url === '/wasm_exec.js') {
                                // Always serve the shared wasm_exec.js from examples/wasm_exec.js
                                if (!fs.existsSync(fixedWasmExecPath)) {
                                    res.statusCode = 404;
                                    return res.end('wasm_exec.js not found');
                                }
                                res.setHeader('Content-Type', 'application/javascript');
                                return fs.createReadStream(fixedWasmExecPath).pipe(res);
                            }
                            if (req.url === '/main.wasm') {
                                // Serve the example's compiled wasm
                                if (!fs.existsSync(exampleWasmPath)) {
                                    res.statusCode = 404;
                                    return res.end('main.wasm not found');
                                }
                                res.setHeader('Content-Type', 'application/wasm');
                                return fs.createReadStream(exampleWasmPath).pipe(res);
                            }
                            // Serve React bridge files from compat/react/
                            if (req.url.startsWith('/compat/react/')) {
                                const fileName = req.url.replace('/compat/react/', '');
                                const filePath = resolve(projectRoot, 'compat/react', fileName);
                                if (fs.existsSync(filePath)) {
                                    res.setHeader('Content-Type', 'application/javascript');
                                    return fs.createReadStream(filePath).pipe(res);
                                } else {
                                    res.statusCode = 404;
                                    return res.end(`${fileName} not found`);
                                }
                            }
                        } catch (e) {
                            // let Vite handle errors/logging
                        }
                        return next();
                    });
                }
            },
            tailwindcss(),
        ],
        optimizeDeps: {
            exclude: ["vendor", ".devenv"]
        },
        resolve: {
            alias: {
                "@": path.resolve(projectRoot, "examples"),
                "/compat/react": path.resolve(projectRoot, "compat/react")
            }
        },
        server: {
            port: 3000,
            watch: {
                ignored: ["**/.devenv/**", "**/.cache/**", "**/vendor/**", "**/node_modules/**"]
            }
        },
        build: {
            cssCodeSplit: false,
            minify: false,
            emptyOutDir: true,
            manifest: true,
            rollupOptions: {
                manualChunks: undefined,
                input: {}
            }
        }
    });
};

export default function (entry, outDir, pubDirs, preCmds = [], exampleName = "counter") {
    const config = baseCfg(pubDirs, preCmds, exampleName);
    config.build.outDir = outDir;
    config.build.rollupOptions.input = entry;
    return config;
}

// Export the base config for direct use
export { baseCfg };