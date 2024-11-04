const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  outputDir: 'dist',
  publicPath: './', // 配置相对路径，确保 Electron 能正确加载资源文件
})
