from PIL import Image
import sys

jpg_path = "/e/jpg.jpg"
print(sys.argv)
if len(sys.argv) != 3:
    sys.exit(1)
img = Image.open(sys.argv[1])
img.save(sys.argv[2])

