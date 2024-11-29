一个给照片添加边框的简单脚本

看下效果

![Screenshot-2024-11-29-at-22 25 37](https://github.com/user-attachments/assets/7a033da3-653a-4c16-bc04-e17216c4f6f4)


使用方法很简单，修改配置文件后运行对应系统的脚本即可

```json
[SnapEdge-mac-arm](SnapEdge-mac-arm)        macos m系列芯片
[SnapEdge-mac-x86](SnapEdge-mac-x86)        macos 英特尔芯片
[SnapEdge-win-32.exe](SnapEdge-win-32.exe)  windows 32位系统
[SnapEdge-win-64.exe](SnapEdge-win-64.exe)  windows 64位系统
```

下面介绍配置文件内容含义


```json
{
  "imagePath": "/Users/wangqi/Desktop/2.35",      // 待处理照片文件地址
  "outputPath": "/Users/wangqi/Desktop/tt",       // 处理后文件存储地址
  "quality": 100,                                 // 文件输出质量，0-100
  "border": {                                     // 边框配置
    "borderWidth": 0.01,                          // 左右边框宽度
    "borderBottomHeight": 0.07,                   // 下边框宽度
    "borderTopHeight": 0.01,                      // 上边框宽度
    "borderLocation": "all",                      // 边框位置  all、 top、 bottom、 left、 right
    "borderColor": "all"                          // 边框颜色
  },
  "logo": {                                       // logo配置
    "on": false,                                  // 是否开启logo展示
    "filePath": "default",                        // logo地址，默认default为根据相机厂商进行展示
    "resize": "auto"                              // logo缩放比例，默认auto为0.8，自定义请定义字符串格式浮点数，例如“0.8”，代表将LOGO图片缩放到80%
  },
  "upperLeft": {                                  // 下边框左上角文字配置
    "on": true,                                   // 是否开启
    "text": "default",                            // 文字内容，默认为厂商 + 相机型号
    "fontPath": "default",                        // 字体文件地址，默认为SFCompactItalic字体
    "fontSize": "auto",                           // 字体大小，默认为下边框的25%， 自定义请设置字符串格式浮点数，例如“0.4”，代表为下边框的40%
    "bold": 0                                     // 加粗细数，0代表不加粗，设置值为正整数
  },
  "lowerLeft": {                                  // 下边框左下角文字配置
    "on": true,                                   // 是否开启
    "text": "default",                            // 文字内容，默认为拍摄镜头
    "fontPath": "default",                        // 字体文件地址，默认为SFCompactItalic字体
    "fontSize": "auto",                           // 字体大小，默认为下边框的25%， 自定义请设置字符串格式浮点数，例如“0.4”，代表为下边框的40%
    "bold": 0                                     // 加粗细数，0代表不加粗，设置值为正整数
  },
  "upperRight": {                                 // 下边框右上角文字配置
    "on": true,                                   // 是否开启
    "text": "default",                            // 文字内容，默认为曝光参数
    "fontPath": "default",                        // 字体文件地址，默认为SFCompactItalic字体
    "fontSize": "auto",                           // 字体大小，默认为下边框的25%， 自定义请设置字符串格式浮点数，例如“0.4”，代表为下边框的40%
    "bold": 0                                     // 加粗细数，0代表不加粗，设置值为正整数
  },
  "lowerRight": {                                 // 下边框右上角文字配置
    "on": true,                                   // 是否开启
    "text": "default",                            // 文字内容，默认为拍摄时间
    "fontPath": "default",                        // 字体文件地址，默认为SFCamera字体
    "fontSize": "auto",                           // 字体大小，默认为下边框的25%， 自定义请设置字符串格式浮点数，例如“0.4”，代表为下边框的40%
    "bold": 0                                     // 加粗细数，0代表不加粗，设置值为正整数
  }
}
```
