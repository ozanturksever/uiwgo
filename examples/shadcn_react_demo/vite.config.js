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
    [`GOOS=js GOARCH=wasm go build ${prod ? '-ldflags="-s -w"' : ''} -o examples/shadcn_react_demo/main.wasm examples/shadcn_react_demo/main.go`],
    "shadcn_react_demo"
);