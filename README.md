tele-dl — download media from telegra.ph

# Description
tele-dl is a command-line program which can help you download all media (both images and videos) from a telegra.ph webpage. It requires Python 3.9+ interpreter. It should work wherever you can install Python. 

# Usage
main.py [-h] --link LINK [--folder FOLDER] [--explicit]
               [--mode {ordered,fast}]

required arguments:
  --link LINK, -L LINK  Enter the full link to the page. Example:
                        "https://telegra.ph/What-Was-TON-And-Why-It-Is-
                        Over-05-12"

optional arguments:
  -h, --help            Show this help message and exit
  --folder FOLDER, -F FOLDER
                        Specify the folder where to extract images
                        Default: current directory
  --explicit, -E        Enable logging
  --mode {ordered,fast}, -M {ordered,fast}
                        Choose mode to download.
                        Ordered — media will download in order of JSON schema from api.
                        Fast — media will download asynchronously orderless.
                        Default: ordered
