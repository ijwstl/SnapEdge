import os
from PIL import Image, ImageOps, ImageDraw, ImageFont
from PIL.ExifTags import TAGS
from datetime import datetime
from fractions import Fraction

"""
    首先确认自己有python3环境
    执行pip3 install pillow
    利用ide打开该文件，然后替换自己的照片目录即可
"""

def get_exif_data(image):
    """
    提取图像的Exif信息
    Args:
        image:
        draw:
        exif_data:
        border_height:

    Returns:

    """
    exif_data = {}
    try:
        exif = image._getexif()
        if exif:
            for tag, value in exif.items():
                tag_name = TAGS.get(tag, tag)
                exif_data[tag_name] = value
    except AttributeError:
        print("无法读取 EXIF 数据")
    return exif_data

def add_expose_info(image, draw, exif_data, border_height):
    """
    添加照片的曝光参数
    Args:
        image:
        draw:
        exif_data:
        border_height:

    Returns:

    """
    expose_info = f"{int(exif_data.get('FocalLength'))}mm    f {exif_data.get('FNumber')}    {Fraction(exif_data.get('ExposureTime')).limit_denominator()}    ISO {exif_data.get('ISOSpeedRatings')}"

    font_size = int(border_height * 0.23)  # 字体大小设置为边框高度的一半
    try:
        font = ImageFont.truetype("/System/Library/Fonts/SFCompactItalic.ttf", size=font_size)
    except IOError:
        font = ImageFont.load_default()  # 如果无法加载字体文件，使用默认字体

    # 计算文本位置
    text_bbox_expose_info = draw.textbbox((0, 0), expose_info, font=font)
    text_width, text_height = text_bbox_expose_info[2] - text_bbox_expose_info[0], text_bbox_expose_info[3] - text_bbox_expose_info[1]
    x_position = image.width - text_width - 50  # 离右边缘 10 像素
    y_position = image.height + (border_height - text_height) * 1 // 4  # 边框区域1/4

    # 绘制文本
    for dx in (-1, 0, 1):
        for dy in (-1, 0, 1):
            draw.text((x_position + dx, y_position + dy), expose_info, font=font, fill="black")

def add_shot_time(image, draw, exif_data, border_height):
    """
    添加照片拍摄时间
    Args:
        image:
        draw:
        exif_data:
        border_height:

    Returns:

    """
    shot_time_info = f"{datetime.strftime(datetime.strptime(exif_data.get('DateTimeOriginal'), '%Y:%m:%d %X'), '%Y-%m-%d %X')}"
    font_size = int(border_height * 0.23)  # 字体大小设置为边框高度的一半
    try:
        font = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", size=font_size)
    except IOError:
        font = ImageFont.load_default()  # 如果无法加载字体文件，使用默认字体

    # 计算文本位置
    text_bbox_expose_info = draw.textbbox((0, 0), shot_time_info, font=font)
    text_width, text_height = text_bbox_expose_info[2] - text_bbox_expose_info[0], text_bbox_expose_info[3] - text_bbox_expose_info[1]
    x_position = image.width - text_width - 50  # 离右边缘 50 像素
    y_position = image.height + (border_height - text_height) * 3// 4  # 边框区域3/4
    # 绘制文本
    draw.text((x_position, y_position), shot_time_info, fill="black", font=font)

def add_device_info(image, draw, exif_data, border_height, logo_width):
    """
    添加照片的相机和镜头信息
    Args:
        image:
        draw:
        exif_data:
        border_height:

    Returns:

    """
    device_info = f"{"Lumix" if exif_data.get('Make') == "Panasonic" else exif_data.get('Make')}    {exif_data.get('Model')}"
    lens_info = f"{exif_data.get('LensModel')}"

    font_size = int(border_height * 0.23)  # 字体大小设置为边框高度的一半
    try:
        font = ImageFont.truetype("/System/Library/Fonts/SFCompactItalic.ttf", size=font_size)
    except IOError:
        font = ImageFont.load_default()  # 如果无法加载字体文件，使用默认字体

    # 计算文本位置
    text_bbox_expose_info = draw.textbbox((0, 0), device_info, font=font)
    text_width, text_height = text_bbox_expose_info[2] - text_bbox_expose_info[0], text_bbox_expose_info[3] - text_bbox_expose_info[1]
    x_position = logo_width + 50  # 离左边缘logo 100 像素
    y_position = image.height + (border_height - text_height) * 1 // 4  # 边框区域的1/3
    lens_y_position = image.height + (border_height - text_height) * 3// 4  # 边框区域的1/3

    # 绘制文本
    # 通过多次绘制，加粗
    for dx in (-1, 0, 1):
        for dy in (-1, 0, 1):
            draw.text((x_position + dx, y_position + dy), device_info, font=font, fill="black")
    draw.text((x_position, lens_y_position), lens_info, fill="black", font=font)


def add_logo(image, image_with_border, exif_data, border_height):
    """
    添加logo
    Args:
        image:
        image_with_border:
        exif_data:
        border_height:

    Returns:

    """
    logo_path = os.path.join(os.path.curdir, "logo", exif_data.get('Make') + ".png")
    logo = Image.open(logo_path)
    logo_size = int(border_height * 0.8)  # 设置 LOGO 大小为边框高度的60%
    logo = logo.resize((int(logo.size[0] * logo_size/ logo.size[1]), logo_size), Image.LANCZOS)
    logo_position = (10, image.height + (border_height - logo_size) // 2)  # 左边距10像素，垂直居中
    # 粘贴 LOGO 到图像
    image_with_border.paste(logo, logo_position)
    return logo.size[0]


def add_white_border_with_text(image_path, output_path=None):
    image = Image.open(image_path)

    # 获取图像的 EXIF 信息
    exif_data = get_exif_data(image)
    # 计算边框大小，边框高度为图片宽度的4% 或 8%
    if (image.width > image.height):
        border_height = int(image.width * 0.04)
    else:
        border_height = int(image.width * 0.07)

    border = (0, 0, 0, border_height)

    # 添加白色边框
    image_with_border = ImageOps.expand(image, border=border, fill='white')

    # 在白色边框上写入文本
    draw = ImageDraw.Draw(image_with_border)
    add_shot_time(image, draw, exif_data, border_height)
    add_expose_info(image, draw, exif_data, border_height)
    logo_width = add_logo(image, image_with_border, exif_data, border_height)
    add_device_info(image, draw, exif_data, border_height, logo_width)

    # 保存带有边框和文本的图像
    image_with_border.save(output_path)
    print(f"带边框图片已保存至：{output_path}")

def main(image_path, output_path=None):
    if os.path.isdir(image_path):
        for picture in os.listdir(image_path):
            if not (picture.endswith(".jpg") or picture.endswith(".png")):
                continue
            main(os.path.join(image_path, picture), output_path=None)
    elif os.path.isfile(image_path):
        if output_path is None:
            tmp_path = os.path.join(os.path.dirname(image_path), "tmp");
            if not os.path.exists(tmp_path):
                os.mkdir(tmp_path)
            output_path = os.path.join(tmp_path, os.path.basename(image_path))
        add_white_border_with_text(image_path, output_path)
    else:
        raise FileNotFoundError

if __name__ == "__main__":
    main("/Users/wangqi/Desktop/2.35")