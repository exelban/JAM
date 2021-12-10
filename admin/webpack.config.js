const webpack = require("webpack")
const path = require("path")
const { VueLoaderPlugin } = require("vue-loader")
const MiniCssExtractPlugin = require("mini-css-extract-plugin")
const { CleanWebpackPlugin } = require("clean-webpack-plugin")
const HtmlWebpackPlugin = require("html-webpack-plugin")
const PACKAGE = require("./package.json")

const PATHS = {
    SRC: path.resolve(__dirname, "src"),
    DIST: path.resolve(__dirname, "dist"),
    NODE_MODULES: path.resolve(__dirname, "node_modules"),
    PUBLIC: path.resolve(__dirname, "/"),
}

module.exports = (env, argv) => {
    const mode = argv.mode
    const prod = mode === "production"

    return {
        target: "web",
        resolve: {
            modules: [PATHS.SRC, PATHS.NODE_MODULES],
            alias: {
                'vue$': 'vue/dist/vue.esm-bundler.js'
            },
            extensions: ["*", ".ts", ".js", ".vue", ".json", ".scss", ".css"],
            mainFields: ["browser", "module", "main"]
        },
        stats: {
            children: false,
        },
        entry: `${PATHS.SRC}/index.js`,
        output: {
            path: __dirname + "/dist",
            filename: "[name].js",
            chunkFilename: "[name].js"
        },
        module: {
            rules: [
                {
                    test: /\.js$/,
                    exclude: /node_modules/,
                    use: {
                        loader: 'babel-loader',
                        options: {
                            presets: ['@babel/preset-env']
                        }
                    }
                },
                {
                    test: /\.vue$/,
                    loader: "vue-loader",
                },
                {
                    test: /\.(sa|sc|c)ss$/,
                    use: [
                        MiniCssExtractPlugin.loader,
                        "css-loader",
                        "sass-loader",
                    ],
                },
                {
                    test: /\.pug$/,
                    loader: "pug-plain-loader",
                }
            ]
        },
        plugins: [
            new VueLoaderPlugin(),
            new CleanWebpackPlugin(),
            new HtmlWebpackPlugin({
                template: `${PATHS.SRC}/index.html`,
                filename: "index.html",
                inject: true,
                minify: {
                    collapseWhitespace: false,
                    collapseInlineTagWhitespace: false,
                    removeComments: true,
                    removeRedundantAttributes: true,
                },
                inlineSource: /^.*$/,
            }),
            new MiniCssExtractPlugin({
                filename: "[name].css"
            }),
            new webpack.DefinePlugin({
                // __VUE_OPTIONS_API__: false,
                // __VUE_PROD_DEVTOOLS__: false,
                "process.env": {
                    VERSION: JSON.stringify(PACKAGE.version),
                    DATE: JSON.stringify(Date.now()),
                },
            })
        ],
        optimization: {
            splitChunks: {
                cacheGroups: {
                    vendor: {
                        test: /node_modules/,
                        chunks: "initial",
                        name: "vendor",
                        enforce: true,
                    },
                    main: {
                        test: /src/,
                        chunks: "initial",
                        name: "main",
                        enforce: true,
                    },
                },
            },
        },
        mode,
        devtool: prod ? false : "source-map",
        devServer: {
            host: "localhost",
            port: 8090,
            compress: true,
            historyApiFallback: true,
            devMiddleware: {
                index: true,
                publicPath: "/",
                serverSideRender: true,
                writeToDisk: true,
            },
        },
    }
}