import viteCfgFactory from "../../vite.config.js";
import { resolve } from "path";

const prod = false;

export default viteCfgFactory(
    resolve(import.meta.dirname, "index.html"),
    "dist",
    [{
        input: resolve(import.meta.dirname, ".") + "/**",
        output: "/"
    }],
    [`GOOS=js GOARCH=wasm go build ${prod ? '-ldflags="-s -w"' : ''} -o examples/action_lifecycle_demo/main.wasm examples/action_lifecycle_demo/main.go`],
    "action_lifecycle_demo"
);