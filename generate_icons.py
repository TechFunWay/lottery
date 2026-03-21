#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from PIL import Image, ImageDraw, ImageFont
import os

def create_lottery_icon(size):
    """创建彩票助手图标"""
    # 创建透明背景
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    # 配置参数
    corner_radius = int(size * 0.25)

    # 橙色背景 (#FF8C00 - 鲜艳橙色)
    bg_color = (255, 140, 0, 255)

    # 创建圆角蒙版（四个角透明，中间填充）
    mask = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.rounded_rectangle(
        [(0, 0), (size, size)],
        radius=corner_radius,
        fill=(255, 255, 255, 255)
    )

    # 绘制完整背景（通过蒙版控制四个角透明）
    for y in range(size):
        for x in range(size):
            if mask.getpixel((x, y))[3] > 0:  # 如果蒙版有像素
                draw.point((x, y), bg_color)

    # 添加微妙渐变效果
    max_layers = int(size * 0.6)
    for i in range(max_layers):
        alpha = int(20 * (1 - i / max_layers))
        # 渐变到稍微深一点的橙色
        color = (230, 120, 0, alpha)

        # 计算当前层的位置
        x1 = i
        y1 = i
        x2 = size - i
        y2 = size - i

        # 检查矩形是否有效
        if x2 <= x1 or y2 <= y1:
            break

        # 当前层的圆角半径
        current_radius = max(0, corner_radius - i)

        # 创建渐变层蒙版
        grad_mask = Image.new('RGBA', (size, size), (0, 0, 0, 0))
        grad_draw = ImageDraw.Draw(grad_mask)
        grad_draw.rounded_rectangle(
            [(x1, y1), (x2, y2)],
            radius=current_radius,
            fill=(255, 255, 255, alpha)
        )

        # 合成渐变层
        for y in range(y1, y2):
            for x in range(x1, x2):
                if grad_mask.getpixel((x, y))[3] > 0 and mask.getpixel((x, y))[3] > 0:
                    current = img.getpixel((x, y))
                    # 混合颜色
                    new_r = int((current[0] * (255 - alpha) + color[0] * alpha) / 255)
                    new_g = int((current[1] * (255 - alpha) + color[1] * alpha) / 255)
                    new_b = int((current[2] * (255 - alpha) + color[2] * alpha) / 255)
                    draw.point((x, y), (new_r, new_g, new_b, 255))

    # 白色文字
    text_color = (255, 255, 255, 255)
    
    # 加载艺术字体
    try:
        font_title = ImageFont.truetype("/System/Library/Fonts/PingFang.ttc", int(size * 0.22), index=4)  # 增大标题字体
        font_number = ImageFont.truetype("/System/Library/Fonts/PingFang.ttc", int(size * 0.24), index=4)  # 调小数字字体
    except:
        try:
            font_title = ImageFont.truetype("/System/Library/Fonts/STHeiti Medium.ttc", int(size * 0.22))  # 增大标题字体
            font_number = ImageFont.truetype("/System/Library/Fonts/STHeiti Medium.ttc", int(size * 0.24))  # 调小数字字体
        except:
            font_title = ImageFont.load_default()
            font_number = ImageFont.load_default()
    
    # 绘制上方"彩票助手"
    title_text = "彩票助手"
    bbox = draw.textbbox((0, 0), title_text, font=font_title)
    title_width = bbox[2] - bbox[0]
    title_height = bbox[3] - bbox[1]
    
    title_x = (size - title_width) // 2
    title_y = int(size * 0.14)  # 稍微上移，给更大的文字留出空间
    
    # 绘制文字阴影（增强立体感）- 改为深色阴影
    shadow_offsets = [
        (int(size * 0.02), int(size * 0.02), (80, 40, 0, 180)),
        (int(size * 0.012), int(size * 0.012), (100, 50, 0, 200)),
        (int(size * 0.006), int(size * 0.006), (120, 60, 0, 220)),
    ]
    
    for offset_x, offset_y, color in shadow_offsets:
        draw.text((title_x + offset_x, title_y + offset_y), title_text, font=font_title, fill=color)
    
    # 绘制主文字
    draw.text((title_x, title_y), title_text, font=font_title, fill=text_color)
    
    # 绘制装饰线条
    line_y = title_y + title_height + int(size * 0.08)
    line_width = int(size * 0.55)
    line_x1 = (size - line_width) // 2
    line_x2 = line_x1 + line_width
    line_height = int(size * 0.012)
    
    # 主装饰线
    draw.rounded_rectangle(
        [(line_x1, line_y), (line_x2, line_y + line_height)],
        radius=line_height // 2,
        fill=text_color
    )
    
    # 两侧装饰点（改为白色）
    deco_size = int(size * 0.025)
    for i in range(3):
        deco_x = line_x1 + (i + 1) * (line_width // 4)
        draw.ellipse(
            [(deco_x - deco_size, line_y + line_height // 2 - deco_size),
             (deco_x + deco_size, line_y + line_height // 2 + deco_size)],
            fill=(255, 255, 255, 180)
        )
    
    # 绘制下方三个圆圈中的"888"
    circle_radius = int(size * 0.13)
    circle_gap = int(size * 0.08)
    total_width = 3 * circle_radius * 2 + 2 * circle_gap
    start_x = (size - total_width) // 2 + circle_radius
    circle_y = line_y + line_height + int(size * 0.12) + circle_radius
    
    # 三个圆圈的颜色
    ball_colors = [
        (231, 76, 60, 255),   # 红色
        (52, 152, 219, 255),  # 蓝色
        (46, 204, 113, 255)   # 绿色
    ]
    
    for i in range(3):
        circle_x = start_x + i * (2 * circle_radius + circle_gap)
        ball_color = ball_colors[i]
        
        # 绘制圆圈阴影
        draw.ellipse(
            [(circle_x - circle_radius + int(size * 0.008), circle_y - circle_radius + int(size * 0.008)),
             (circle_x + circle_radius + int(size * 0.008), circle_y + circle_radius + int(size * 0.008))],
            outline=(150, 100, 50, 150),
            width=int(size * 0.015)
        )
        
        # 绘制圆圈填充（彩色）
        draw.ellipse(
            [(circle_x - circle_radius, circle_y - circle_radius),
             (circle_x + circle_radius, circle_y + circle_radius)],
            fill=ball_color
        )
        
        # 绘制圆圈边框
        draw.ellipse(
            [(circle_x - circle_radius, circle_y - circle_radius),
             (circle_x + circle_radius, circle_y + circle_radius)],
            outline=text_color,
            width=int(size * 0.018)
        )
        
        # 绘制数字"8"（白色）
        number_text = "8"
        bbox = draw.textbbox((0, 0), number_text, font=font_number)
        num_width = bbox[2] - bbox[0]
        num_height = bbox[3] - bbox[1]
        
        num_x = circle_x - num_width // 2
        num_y = circle_y - num_height // 2
        
        # 数字阴影
        draw.text((num_x + int(size * 0.004), num_y + int(size * 0.004)), number_text, font=font_number, fill=(0, 0, 0, 120))
        
        # 主数字（白色）
        draw.text((num_x, num_y), number_text, font=font_number, fill=(255, 255, 255, 255))
    
    # 添加底部装饰元素（改为白色）
    deco_size = int(size * 0.025)
    deco_y = size - int(size * 0.12)
    for i in range(5):
        deco_x = int(size * 0.25) + i * int(size * 0.125)
        alpha = 255 - abs(2 - i) * 50
        draw.ellipse(
            [(deco_x - deco_size, deco_y - deco_size),
             (deco_x + deco_size, deco_y + deco_size)],
            fill=(255, 255, 255, alpha)
        )

    return img

def main():
    sizes = [64, 256]

    output_dir = "/Users/weiyi/develop/gitee/TechFunWay/lottery/techfunway-lottery"
    app_ui_dir = os.path.join(output_dir, "app/ui/images")

    # 创建目录
    os.makedirs(app_ui_dir, exist_ok=True)

    for size in sizes:
        # 生成图标
        icon = create_lottery_icon(size)

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
    print("  - 圆角正方形，四个角透明，中间区域填充")
    print("  - 橙色背景 (#FF8C00)")
    print("  - 上方：艺术字'彩票助手'（增强阴影）")
    print("  - 下方：三个彩色圆圈圈住的'888'（红、蓝、绿，数字更小）")

if __name__ == "__main__":
    main()
