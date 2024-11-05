<template>
  <el-row :gutter="20">
    <el-col :span="20">
      <el-card>
        <div class="thumbnail-container">
          <img
            :src="imageSrc"
            alt="Thumbnail"
            class="thumbnail"
            @click="openPreview"
          />
        </div>
        <image-preview
          :imageSrc="imageSrc"
          :showPreview="isPreviewVisible"
          :onClose="closePreview"
        />
      </el-card>
    </el-col>
    <el-col :span="4">
      <el-card style="max-width: 480px">
        <template #header>
          <div class="card-header">
            <span>配置</span>
          </div>
        </template>
        <el-button type="primary" @click="selectImage">选择 JPG 图片</el-button>
        <edit-config></edit-config>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref } from "vue";
import ImagePreview from "./ImagePreview.vue";
import EditConfig from "./EditConfig.vue";

const imageSrc = ref(null);
const isPreviewVisible = ref(false);
// const borderWidth = ref(10);

const selectImage = async () => {
  const filePath = await window.electronAPI.selectImageFile();
  if (filePath) {
    // 使用主进程读取文件内容
    const imageBase64 = await window.electronAPI.readImageFile(filePath);
    imageSrc.value = imageBase64; // 设置为 base64 数据
  }
};

const openPreview = () => {
  isPreviewVisible.value = true;
};

const closePreview = () => {
  isPreviewVisible.value = false;
};
</script>

<style scoped>
.thumbnail-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  border: 1px solid #ccc; /* 可选：添加边框以更好地看到容器 */
}

.thumbnail {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
</style>
