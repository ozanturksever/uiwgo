import viteCfgFactory, { parseCliArgs } from "../../vite.config.js";
import { resolve } from "path";

const { prod } = parseCliArgs();

export default viteCfgFactory(
    resolve(import.meta.dirname, "index.html"),
    "dist",
    [{
        input: resolve(import.meta.dirname, ".") + "/**",
        output: "/"
    }],
    [`GOOS=js GOARCH=wasm go build ${prod ? '-ldflags="-s -w"' : ''} -o examples/router_test/main.wasm examples/router_test/main.go`],
    "router_test"
);