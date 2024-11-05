<template>
  <div v-if="showPreview" class="preview-container" @click="onClose">
    <img :src="imageSrc" class="preview-image" alt="Preview Image" />
  </div>
</template>

<script>
export default {
  name: "ImagePreview",
  props: {
    imageSrc: String, // 传入图片的路径
    showPreview: Boolean, // 控制是否显示预览
    onClose: Function, 
  },
};
</script>

<style scoped>
.preview-container {
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.8); /* 添加背景色 */
  position: fixed;
  top: 0;
  left: 0;
  z-index: 1000;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
</style>
