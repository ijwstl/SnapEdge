<template>
  <SnapEdge class="app-container"/>
</template>

<script>
import SnapEdge from "./components/SnapEdge.vue";

export default {
  name: "App",
  components: {
    SnapEdge,
  },
};
</script>

<style>
html,
body,#app{
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  font-family: Avenir, Helvetica, Arial, sans-serif;
  overflow: hidden; /* 禁止页面滚动 */
}

.app-container {
  height: 100%;
  width: 100%;
  padding: 0%;
  font-family: Avenir, Helvetica, Arial, sans-serif;
  display: flex;
  overflow: hidden; /* 防止内容溢出 */
  background-color: #f0f0f0; /* 可选：背景颜色 */
}
</style>
