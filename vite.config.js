import { defineConfig } from "vite";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import vue from '@vitejs/plugin-vue'


// https://vite.dev/config/
export default defineConfig({
  transpileDependencies: true,
  outputDir: "dist",
  publicPath: "./", // 配置相对路径，确保 Electron 能正确加载资源文件
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
});
