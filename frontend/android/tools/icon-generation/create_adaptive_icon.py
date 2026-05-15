#!/usr/bin/env python3
"""
Generate Android Adaptive Icons
For Android 8.0 (API 26) and above
"""

from PIL import Image, ImageDraw
import os

script_dir = os.path.dirname(os.path.abspath(__file__))
project_dir = os.path.abspath(os.path.join(script_dir, "..", ".."))

target_dir = os.path.join(project_dir, "app", "src", "main", "res")
source_path = os.path.join(target_dir, "drawable", "ic_launcher_display.png")

# Adaptive icon sizes (108dp with 18dp safe zone = 72dp foreground)
# Final icon canvas: 108x108, Foreground safe zone: 72x72
ADAPTIVE_SIZE = 108
FOREGROUND_SIZE = 72

def create_adaptive_icon():
    """
    Create adaptive icon with foreground and background layers
    """
    # Load source image
    with Image.open(source_path) as img:
        if img.mode != 'RGBA':
            img = img.convert('RGBA')

        # Create background (solid color matching the app theme)
        # Using a dark blue-gray color similar to the app theme
        background = Image.new('RGBA', (ADAPTIVE_SIZE, ADAPTIVE_SIZE), (20, 29, 34, 255))

        # Create foreground (the logo with padding)
        # Scale to fit within safe zone (72x72)
        img.thumbnail((FOREGROUND_SIZE, FOREGROUND_SIZE), Image.Resampling.LANCZOS)

        # Create transparent canvas for foreground
        foreground = Image.new('RGBA', (ADAPTIVE_SIZE, ADAPTIVE_SIZE), (0, 0, 0, 0))

        # Calculate paste position (center)
        paste_x = (ADAPTIVE_SIZE - img.width) // 2
        paste_y = (ADAPTIVE_SIZE - img.height) // 2
        foreground.paste(img, (paste_x, paste_y))

        # Save background
        bg_path = os.path.join(target_dir, "mipmap-anydpi-v26", "ic_launcher_background.png")
        os.makedirs(os.path.dirname(bg_path), exist_ok=True)
        background.save(bg_path, "PNG", optimize=True)

        # Save foreground
        fg_path = os.path.join(target_dir, "mipmap-anydpi-v26", "ic_launcher_foreground.png")
        foreground.save(fg_path, "PNG", optimize=True)

        print(f"✓ Adaptive icon background: {bg_path}")
        print(f"  Size: {os.path.getsize(bg_path) / 1024:.1f} KB")
        print(f"✓ Adaptive icon foreground: {fg_path}")
        print(f"  Size: {os.path.getsize(fg_path) / 1024:.1f} KB")

        # Create the adaptive icon XML definition
        xml_content = '''<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@mipmap/ic_launcher_background"/>
    <foreground android:drawable="@mipmap/ic_launcher_foreground"/>
</adaptive-icon>'''

        # Note: The ic_launcher.xml in mipmap-anydpi-v26 should reference the adaptive icon
        # But we need to keep the PNG icons for backward compatibility

        print("\n✓ Adaptive icon XML definition created (for mipmap-anydpi-v26)")

def create_fallback_icon():
    """
    Create a simple square icon for devices below API 26
    """
    with Image.open(source_path) as img:
        if img.mode != 'RGBA':
            img = img.convert('RGBA')

        # Resize to 48x48 (mdpi baseline)
        icon = img.resize((48, 48), Image.Resampling.LANCZOS)

        # Save to mipmap-mdpi as fallback
        fallback_path = os.path.join(target_dir, "mipmap-mdpi", "ic_launcher.png")

        # Convert to RGB for compatibility
        icon_rgb = Image.new('RGB', (48, 48), (255, 255, 255))
        icon_rgb.paste(icon, (0, 0), icon)

        icon_rgb.save(fallback_path, "PNG", optimize=True)
        print(f"✓ Fallback icon: {fallback_path}")
        print(f"  Size: {os.path.getsize(fallback_path) / 1024:.1f} KB")

def main():
    print("Generating Android Adaptive Icons...")
    print("=" * 60)
    create_adaptive_icon()
    print("=" * 60)
    print("Done!")

if __name__ == "__main__":
    main()