import pathlib
import argparse


def convert_bytes(num):
    for x in ['bytes', 'KB', 'MB', 'GB', 'TB']:
        if num < 1024.0:
            return f'{num:.2f} {x}'
        num /= 1024.0


def getsize(path):
    path_object = pathlib.Path(path)
    raw_size = 0
    formatted_size = 0
    if path_object.is_file():
        if path_object.exists():
            raw_size = path_object.stat().st_size
            formatted_size = convert_bytes(raw_size)
    else:
        raw_size = sum(f.stat().st_size for f in path_object.glob('**/*') if f.is_file())
        formatted_size = convert_bytes(raw_size)

    return {'raw': raw_size, 'formatted': formatted_size}


def arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument('--link', '-L', help='Enter the full link to the page. Example: '
                                             '"https://telegra.ph/What-Was-TON-And-Why-It-Is-Over-05-12"', type=str,
                        required=True)
    parser.add_argument('--folder', '-F', help='Specify the folder where to extract images', type=pathlib.Path,
                        default=pathlib.Path().absolute())
    parser.add_argument('--explicit', '-E', help='Show all messages', action="store_true")

    return parser