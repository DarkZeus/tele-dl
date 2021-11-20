import argparse
from pathlib import Path


def arguments() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser()
    parser.add_argument('--link', '-L', help='Enter the full link to the page. Example: '
                                             '"https://telegra.ph/What-Was-TON-And-Why-It-Is-Over-05-12"', type=str,
                        required=True)
    parser.add_argument('--folder', '-F', help='Specify the folder where to extract images', type=Path,
                        default=Path().absolute())
    parser.add_argument('--explicit', '-E', help='Show all messages', action="store_true")
    parser.add_argument('--compress', '-C', help='Compress all photos to WebP format, is 25 – 34% smaller than JPEG '
                                                 'or PNG at equivalent quality', action="store_true")

    return parser
