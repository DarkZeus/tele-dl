import pathlib


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