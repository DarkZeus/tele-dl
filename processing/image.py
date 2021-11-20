from pathlib import Path

import pyvips

from utils import getsize, log


def compress_image(_bytes: bytes, destination: Path) -> None:
    image = pyvips.Image.new_from_buffer(_bytes, "")
    target = pyvips.Target.new_to_file(str(destination))
    image.write_to_target(target, ".webp", preset="photo", Q=80, smart_subsample=True)

    log(f"[cwebp] {destination.name} — {getsize(destination)['formatted']}")


