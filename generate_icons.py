#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from PIL import Image, ImageDraw, ImageFont
import os

def process_icon_from_image(input_image_path, size):
    """从图片裁剪并处理图标"""
    # 打开原始图片
    img = Image.open(input_image_path)

    # 转换为RGBA模式（如果需要）
    if img.mode != 'RGBA':
        img = img.convert('RGBA')

    width, height = img.size

    # 去掉右下角水印（假设水印在右下角约10%区域）
    watermark_ratio = 0.10
    watermark_width = int(width * watermark_ratio)
    watermark_height = int(height * watermark_ratio)
    watermark_left = width - watermark_width
    watermark_top = height - watermark_height

    # 将水印区域设为透明
    for y in range(watermark_top, height):
        for x in range(watermark_left, width):
            pixel = img.getpixel((x, y))
            img.putpixel((x, y), (pixel[0], pixel[1], pixel[2], 0))

    # 缩放到指定尺寸（保持比例，居中裁剪）
    # 计算缩放比例
    scale = max(size / width, size / height)
    new_width = int(width * scale)
    new_height = int(height * scale)

    # 缩放图片
    img = img.resize((new_width, new_height), Image.Resampling.LANCZOS)

    # 计算居中裁剪位置
    left = (new_width - size) // 2
    top = (new_height - size) // 2
    right = left + size
    bottom = top + size

    # 裁剪为正方形
    img = img.crop((left, top, right, bottom))

    # 创建圆角蒙版
    corner_radius = int(size * 0.25)
    mask = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.rounded_rectangle(
        [(0, 0), (size, size)],
        radius=corner_radius,
        fill=(255, 255, 255, 255)
    )

    # 应用圆角蒙版
    result = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    for y in range(size):
        for x in range(size):
            if mask.getpixel((x, y))[3] > 0:  # 如果蒙版有像素
                result.putpixel((x, y), img.getpixel((x, y)))

    return result

def main():
    sizes = [64, 256]

    # 源图片路径
    source_image = "/Users/weiyi/develop/gitee/TechFunWay/lottery/已移除背景的icon3.png"
    
    output_dir = "/Users/weiyi/develop/gitee/TechFunWay/lottery/techfunway-lottery"
    app_ui_dir = os.path.join(output_dir, "app/ui/images")

    # 创建目录
    os.makedirs(app_ui_dir, exist_ok=True)

    for size in sizes:
        # 生成图标
        icon = process_icon_from_image(source_image, size)

        # 保存到不同位置
        icon_64_path = os.path.join(output_dir, f"ICON_{'256' if size == 256 else ''}.PNG")
        if size == 64:
            icon_64_path = os.path.join(output_dir, "ICON.PNG")

        icon_ui_path = os.path.join(app_ui_dir, f"icon_{size}.png")

        # 保存
        icon.save(icon_64_path, "PNG")
        icon.save(icon_ui_path, "PNG")

        print(f"✓ 已生成 {size}x{size} 图标:")
        print(f"  - {icon_64_path}")
        print(f"  - {icon_ui_path}")

    print("\n✅ 所有图标生成完成！")
    print("设计特点：")
    print("  - 从源图片裁剪并缩放")
    print("  - 圆角正方形，四个角透明，中间区域填充")
    print("  - 自动居中裁剪为正方形")

if __name__ == "__main__":
    main()
