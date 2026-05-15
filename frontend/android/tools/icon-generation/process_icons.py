#!/usr/bin/env python3
"""
Android App Icon Processor
Generates optimized launcher icons for all Android screen densities
"""

from PIL import Image, ImageOps
import os

script_dir = os.path.dirname(os.path.abspath(__file__))
project_dir = os.path.abspath(os.path.join(script_dir, "..", ".."))

source_path = os.path.join(project_dir, "design-assets", "AlfQ app logo设计.png")
target_dir = os.path.join(project_dir, "app", "src", "main", "res")

# Android mipmap densities and their sizes
# Using square icons as per Android guidelines
densities = {
    "mipmap-mdpi": 48,
    "mipmap-hdpi": 72,
    "mipmap-xhdpi": 96,
    "mipmap-xxhdpi": 144,
    "mipmap-xxxhdpi": 192,
}

def process_icon(source_path, target_size):
    """
    Process the source image to create a launcher icon
    - Maintains aspect ratio
    - Adds padding if necessary to make it square
    - Ensures proper size
    """
    with Image.open(source_path) as img:
        # Convert to RGBA if needed (for transparency handling)
        if img.mode != 'RGBA':
            img = img.convert('RGBA')

        # Get original dimensions
        original_width, original_height = img.size

        # Calculate the target square size
        target_dim = target_size

        # Calculate scaling factor to fit the largest dimension within target
        scale = target_dim / max(original_width, original_height)

        # Resize maintaining aspect ratio
        new_width = int(original_width * scale)
        new_height = int(original_height * scale)
        img = img.resize((new_width, new_height), Image.Resampling.LANCZOS)

        # Create a new square image with transparent padding
        # Use the center of the original image for padding
        new_img = Image.new('RGBA', (target_dim, target_dim), (0, 0, 0, 0))

        # Calculate paste position (center)
        paste_x = (target_dim - new_width) // 2
        paste_y = (target_dim - new_height) // 2

        # Paste the resized image
        new_img.paste(img, (paste_x, paste_y))

        # Convert back to RGB with white background (for icons without transparency)
        # This ensures compatibility with all Android launchers
        final_img = Image.new('RGB', (target_dim, target_dim), (255, 255, 255))
        final_img.paste(new_img, (0, 0), new_img)

        return final_img

def main():
    print(f"Processing Android launcher icons from: {source_path}")
    print("=" * 60)

    # Create directories and generate icons
    for density_dir, size in densities.items():
        dir_path = os.path.join(target_dir, density_dir)

        # Create directory if it doesn't exist
        os.makedirs(dir_path, exist_ok=True)

        # Generate icon
        icon = process_icon(source_path, size)

        # Save with optimization
        output_path = os.path.join(dir_path, "ic_launcher.png")

        # Save as PNG with maximum compression
        icon.save(output_path, "PNG", optimize=True)

        # Get file size
        file_size = os.path.getsize(output_path)

        print(f"✓ {density_dir}: {size}x{size} -> {output_path}")
        print(f"  File size: {file_size / 1024:.1f} KB")

    print("=" * 60)
    print("Icon generation complete!")

    # Also create the original sized version in drawable for reference
    drawable_dir = os.path.join(target_dir, "drawable")
    os.makedirs(drawable_dir, exist_ok=True)

    with Image.open(source_path) as img:
        if img.mode != 'RGBA':
            img = img.convert('RGBA')

        # Resize to a reasonable display size (e.g., 512x512 for high-res display)
        display_size = 512
        img.thumbnail((display_size, display_size), Image.Resampling.LANCZOS)

        output_path = os.path.join(drawable_dir, "ic_launcher_display.png")
        img.save(output_path, "PNG", optimize=True)

        file_size = os.path.getsize(output_path)
        print(f"\n✓ High-resolution reference: {output_path}")
        print(f"  File size: {file_size / 1024:.1f} KB")

if __name__ == "__main__":
    main()