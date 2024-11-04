<template>
  <n-config-provider>
    <n-space vertical>
      <n-layout has-sider sider-placement="right">
        <n-layout-content content-style="padding: 24px;">
          <div v-if="imageSrc" :style="imageContainerStyle">
            <img :src="imageSrc" :style="imageStyle" />
          </div>
        </n-layout-content>
        <n-layout-sider
          collapse-mode="transform"
          :collapsed-width="10"
          :width="240"
          :native-scrollbar="false"
          show-trigger="bar"
          content-style="padding: 24px;"
          bordered
        >
          <n-button type="primary" @click="selectImage">选择 JPG 图片</n-button>
          <n-slider
            v-model:value="borderWidth"
            :max="50"
            :min="0"
            step="1"
            label="边框宽度"
            style="width: 300px"
          />
        </n-layout-sider>
      </n-layout>
    </n-space>
  </n-config-provider>
</template>

<script setup>
import { ref, computed } from "vue";
import {
  NButton,
  NConfigProvider,
  NSlider,
  NSpace,
  NLayout,
  NLayoutContent,
  NLayoutSider,
} from "naive-ui";

const imageSrc = ref(null);
const borderWidth = ref(10);

const selectImage = async () => {
  const filePath = await window.electronAPI.selectImageFile();
  if (filePath) {
    // 使用主进程读取文件内容
    const imageBase64 = await window.electronAPI.readImageFile(filePath);
    imageSrc.value = imageBase64; // 设置为 base64 数据
  }
};

const imageContainerStyle = computed(() => ({
  padding: `${borderWidth.value}px`,
  backgroundColor: "white",
  display: "inline-block",
  borderRadius: "4px",
  boxShadow: "0 0 10px rgba(0, 0, 0, 0.2)",
}));

const imageStyle = {
  width: "100%",
  height: "auto",
};
</script>

<style scoped>
.n-layout-sider {
  background: rgba(128, 128, 128, 0.3);
}

.n-layout-content {
  background: rgba(128, 128, 128, 0.4);
}
</style>
