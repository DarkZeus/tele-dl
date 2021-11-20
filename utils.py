from pathlib import Path
from parser import arguments
import mimetypes

parser = arguments()


IMAGE_FILE_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.webp']

def convert_bytes(num: int) -> str | float:
    for x in ['bytes', 'KB', 'MB', 'GB', 'TB']:
        if num < 1024.0:
            return f'{num:.2f} {x}'
        num /= 1024.0


def getsize(path: Path) -> dict:
    raw_size = 0
    formatted_size = 0
    if path.is_file():
        if path.exists():
            raw_size = path.stat().st_size
            formatted_size = convert_bytes(raw_size)
    else:
        raw_size = sum(f.stat().st_size for f in path.glob('**/*') if f.is_file())
        formatted_size = convert_bytes(raw_size)

    return {'raw': raw_size, 'formatted': formatted_size}


def append_extension(path_string: str, extension: str) -> Path:
    return Path(f"{''.join([path_string.removesuffix(ext) for ext in IMAGE_FILE_EXTENSIONS if path_string.endswith(ext)])}.{extension}")


def is_image_by_url(path: Path) -> bool:
    return any(str(path).endswith(ext) for ext in IMAGE_FILE_EXTENSIONS)


def create_directory(folder: Path):
    if not folder.exists():
        try:
            Path(folder).mkdir(parents=True, exist_ok=True)
        except OSError:
            log(f"[os] Creation of the directory {folder} failed")
        else:
            log(f"[os] Successfully created the directory \"{folder}\"")


def log(data: str | list) -> None:
    if parser.parse_args().explicit:
        [print(message) for message in data] if isinstance(data, list) else print(data)

