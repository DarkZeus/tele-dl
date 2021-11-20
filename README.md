tele-dl — download media from telegra.ph

# Description
tele-dl is a command-line program which can help you download all media (both images and videos) from a telegra.ph webpage. It requires Python 3.10+ interpreter. It should work wherever you can install Python. 

# Usage
```
main.py [-h] --link LINK [--folder FOLDER] [--explicit] [--compress]

options:
  -h, --help            show this help message and exit
  --link LINK, -L LINK  Enter the full link to the page. Example:
                        "https://telegra.ph/What-Was-TON-And-Why-It-Is-
                        Over-05-12"
  --folder FOLDER, -F FOLDER
                        Specify the folder where to extract images
  --explicit, -E        Show all messages
  --compress, -C        Compress all photos to WebP format, which is smaller
                        than JPEG or PNG at equivalent quality


```
# TODO
 - [ ] Implement import from CSV, JSON and other data formats
 - [ ] Implement modularity for the ability to download not only from telegra.ph.
